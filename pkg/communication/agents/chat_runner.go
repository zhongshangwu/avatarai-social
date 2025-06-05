package agents

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory/converters"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/prompt"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type ChatRunner struct {
	*BaseRunner
	LLMManager *llm.ModelManager

	runnings sync.Map
}

func NewChatRunner(
	llmManager *llm.ModelManager,
) *ChatRunner {
	return &ChatRunner{
		BaseRunner: NewBaseRunner("ChatRunner", "处理 AI 聊天消息的智能体"),
		LLMManager: llmManager,
	}
}

func (a *ChatRunner) Ctrl(ctx *ChatControlContext) error {
	responseID := ctx.ResponseID

	runCtx, ok := a.runnings.Load(responseID)
	if !ok {
		logrus.Errorf("响应ID %s 不存在", responseID)
		return errors.New("响应ID不存在")
	}

	invokeCtx := runCtx.(*ChatInvokeContext)

	// 真正的Fire-and-forget：非阻塞发送，立即返回结果
	select {
	case invokeCtx.ControlChan <- ctx.CtrlType:
		logrus.Infof("成功发送控制信号 %s 到响应 %s", ctx.CtrlType, responseID)
		return nil
	default:
		// 立即失败，不等待 - 通道可能已满或接收端已停止
		logrus.Warnf("控制信号发送失败，响应 %s 的通道已满或已关闭", responseID)
		return errors.New("控制信号发送失败：通道不可用")
	}
}

func (a *ChatRunner) Invoke(ctx *ChatInvokeContext) error {
	logrus.Info("ChatRunner 开始处理 AI 聊天消息")

	go func() {
		defer ctx.Stream.CloseSend()
		if err := a._invoke(ctx); err != nil {
			logrus.Errorf("处理 AI 聊天消息失败: %v", err)
		}
	}()

	return nil
}

func (a *ChatRunner) _invoke(ctx *ChatInvokeContext) error {
	if ctx.Response == nil {
		return ctx.sendError("invalid_context", "缺少 AI 聊天消息")
	}
	responseID := ctx.Response.ID
	a.runnings.Store(responseID, ctx)
	defer a.runnings.Delete(responseID)

	if err := ctx.sendAIChatCreated(ctx.Response); err != nil {
		logrus.Errorf("发送创建事件失败: %v", err)
		return err
	}

	if err := ctx.sendAIChatInProgress(ctx.Response); err != nil {
		logrus.Errorf("发送进行中事件失败: %v", err)
		return err
	}

	chunks, err := ctx.Memory.Retrieve(&memory.MessageChunk{
		ID: ctx.Response.MessageID,
		Content: &messages.Message{
			ID: ctx.Response.MessageID,
			Content: &messages.TextMessageContent{
				Text: ctx.Response.AltText,
			},
		},
	})
	if err != nil {
		logrus.Errorf("获取记忆失败: %v", err)
		return ctx.sendAIChatFailed(ctx.Response, "memory_error", "获取记忆失败: "+err.Error())
	}

	var promptMessages []*llm.PromptMessage

	for _, chunk := range chunks {
		p := converters.ChunkToLLM(chunk)
		logrus.Infof("promptMessages: %+v", p)
		promptMessages = append(promptMessages, p)
	}
	transformer := &prompt.LLMEntitiesTransform{}
	tools := transformer.TransformTools(ctx.Response.Tools)

	if err := a.processLLMInteraction(ctx, promptMessages, tools); err != nil {
		logrus.Errorf("处理 LLM 交互失败: %v", err)
		return ctx.sendAIChatFailed(ctx.Response, "llm_error", "LLM 交互失败: "+err.Error())
	}

	return nil
}

func (a *ChatRunner) processLLMInteraction(ctx *ChatInvokeContext, promptMessages []*llm.PromptMessage, tools []llm.PromptMessageTool) error {
	modelParameters := map[string]interface{}{
		"temperature": 0.7,
	}

	logrus.Info("准备发起 LLM 请求")
	defer logrus.Info("LLM 请求处理完成")

	llmCtx, cancel := context.WithTimeout(ctx.Context, 5*time.Minute)
	defer cancel()

	chatStream, err := a.LLMManager.ChatStream(llmCtx, promptMessages, modelParameters, tools, nil)
	if err != nil {
		return ctx.sendAIChatFailed(ctx.Response, messages.ResponseErrorCodeServerError, "LLM 请求失败: "+err.Error())
	}

	// FIXME: 将 Recv 改成 non-blocking 模式
	for {
		select {
		case ctrlType := <-ctx.ControlChan:
			logrus.Infof("收到控制事件: %s", ctrlType)
			switch ctrlType {
			case CtrlTypeInterrupt:
				logrus.Info("收到中断信号，停止流处理")
				return a.handleManuallyInterrupt(ctx)
			default:
				logrus.Warnf("未知的控制事件类型: %s", ctrlType)
			}
		default:
			result := chatStream.Recv()

			if result.HasData {
				chunk := result.Data
				if err := a.processChunk(ctx, chunk); err != nil {
					logrus.Errorf("处理块失败: %v", err)
					return a.handleServerInterrupt(ctx, messages.ResponseErrorCodeServerError, "处理块失败: "+err.Error())
				}

				if chunk.Delta.FinishReason != "" && chunk.Delta.FinishReason != "stop" {
					logrus.Infof("收到完成原因: %s", chunk.Delta.FinishReason)

					if err := a.handleFinishReason(ctx, chunk.Delta.FinishReason); err != nil {
						return err
					}
				}
				continue
			}

			if result.Completed {
				logrus.Info("流已完成")

				if result.Error != nil {
					if errors.Is(result.Error, streams.ErrContextAlreadyDone) || errors.Is(result.Error, streams.ErrChannelClosed) {
						logrus.Infof("流已关闭或上下文已取消: %v", result.Error)
						return a.handleServerInterrupt(ctx, messages.ResponseErrorCodeServerError, "流已关闭或上下文已取消")
					}
					logrus.Errorf("接收流数据错误: %v", result.Error)
					return a.handleServerInterrupt(ctx, messages.ResponseErrorCodeServerError, "接收流数据错误: "+result.Error.Error())
				}

				logrus.Info("流已完成，开始处理完成事件")
				if err := a.finalizeAllOutputItems(ctx); err != nil {
					return err
				}
				return ctx.sendAIChatCompleted(ctx.Response)
			}
		}
	}
}

func (a *ChatRunner) processChunk(ctx *ChatInvokeContext, chunk *llm.LLMResultChunk) error {
	delta := chunk.Delta
	message := delta.Message

	if content, ok := message.Content.(string); ok && content != "" {
		return a.handleTextContent(ctx, content)
	}

	if len(message.ToolCalls) > 0 {
		// toolAgent := NewToolEngine(ctx.Sender, a.LLMManager)
		// return toolAgent.HandleToolCalls(ctx, message.ToolCalls)
	}

	if chunk.Delta.Usage != nil {
		ctx.Response.Usage = &messages.ResponseUsage{
			InputTokens:  chunk.Delta.Usage.PromptTokens,
			OutputTokens: chunk.Delta.Usage.CompletionTokens,
			TotalTokens:  chunk.Delta.Usage.TotalTokens,
		}
	}

	return nil
}

func (a *ChatRunner) handleTextContent(ctx *ChatInvokeContext, content string) error {
	if err := a.ensureOutputMessage(ctx); err != nil {
		return err
	}

	if err := a.ensureTextContent(ctx); err != nil {
		return err
	}

	ctx.CurrentTextContent.Text += content

	return ctx.sendTextDelta(ctx.CurrentOutputMessage.ID, ctx.CurrentOutputItemIdx, 0, content)
}

func (a *ChatRunner) ensureOutputMessage(ctx *ChatInvokeContext) error {
	if ctx.CurrentOutputMessage != nil {
		return nil
	}

	outputMessage := &messages.OutputMessage{
		ID:      uuid.New().String(),
		Type:    "message",
		Role:    "assistant",
		Content: []messages.OutputContent{},
		Status:  "in_progress",
	}

	ctx.Response.MessageItems = append(ctx.Response.MessageItems, outputMessage)
	ctx.CurrentOutputItemIdx = len(ctx.Response.MessageItems) - 1
	ctx.CurrentOutputMessage = outputMessage
	return ctx.sendOutputItemAdded(ctx.CurrentOutputItemIdx, outputMessage)
}

func (a *ChatRunner) ensureTextContent(ctx *ChatInvokeContext) error {
	if ctx.CurrentTextContent != nil {
		return nil
	}

	textContent := &messages.OutputTextContent{
		Type: "output_text",
		Text: "",
	}

	ctx.CurrentOutputMessage.Content = append(ctx.CurrentOutputMessage.Content, textContent)
	contentIndex := len(ctx.CurrentOutputMessage.Content) - 1
	ctx.CurrentTextContent = textContent

	return ctx.sendContentPartAdded(ctx.CurrentOutputMessage.ID, ctx.CurrentOutputItemIdx, contentIndex, textContent)
}

func (a *ChatRunner) finalizeAllOutputItems(ctx *ChatInvokeContext) error {
	for i, item := range ctx.Response.MessageItems {
		if outputMsg, ok := item.(*messages.OutputMessage); ok {
			if err := a.finalizeOutputMessage(ctx, outputMsg, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *ChatRunner) finalizeOutputMessage(ctx *ChatInvokeContext, outputMsg *messages.OutputMessage, index int) error {
	for j, content := range outputMsg.Content {
		if textContent, ok := content.(*messages.OutputTextContent); ok {
			if err := ctx.sendTextDone(outputMsg.ID, index, j, textContent.Text); err != nil {
				return err
			}

			if err := ctx.sendContentPartDone(outputMsg.ID, index, j, textContent); err != nil {
				return err
			}

			if ctx.Response.AltText == "" {
				ctx.Response.AltText = textContent.Text
			} else {
				ctx.Response.AltText += textContent.Text
			}
		}
	}

	outputMsg.Status = "completed"

	return ctx.sendOutputItemDone(index, outputMsg)
}
func (a *ChatRunner) handleManuallyInterrupt(ctx *ChatInvokeContext) error {
	logrus.Info("处理中断事件")
	ctx.Response.Status = messages.AgentMessageStatusIncomplete
	ctx.Response.InterruptType = int32(messages.InterruptTypeUser)
	return ctx.sendAIChatIncomplete(ctx.Response)
}

func (a *ChatRunner) handleContextCancellation(ctx *ChatInvokeContext) error {
	logrus.Info("处理上下文取消")
	return a.handleManuallyInterrupt(ctx) // 上下文取消也按中断处理
}

func (a *ChatRunner) handleServerInterrupt(ctx *ChatInvokeContext, code messages.ResponseErrorCode, msg string) error {
	logrus.Info("处理服务端中断")
	ctx.Response.Status = messages.AgentMessageStatusIncomplete
	ctx.Response.InterruptType = int32(messages.InterruptTypeSystem)
	ctx.Response.Error = &messages.ResponseError{
		Code:    code,
		Message: msg,
	}
	return ctx.sendAIChatIncomplete(ctx.Response)
}

func (a *ChatRunner) handleFinishReason(ctx *ChatInvokeContext, finishReason string) error {
	if finishReason == "stop" {
		return nil
	}

	code := messages.ResponseErrorCodeServerError
	msg := "LLM 请求终止"
	if finishReason == "length" {
		code = messages.ResponseErrorCodeLLMRequestLength
		msg = "LLM 请求长度超出限制"
	}
	return a.handleServerInterrupt(ctx, code, msg)
}
