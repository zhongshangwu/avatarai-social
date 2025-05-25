package chat

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
	"github.com/zhongshangwu/avatarai-social/pkg/types/chat"
	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

type ChatActor struct {
	*events.BaseActor[*chat.ChatEvent]
	llmManager *llm.ModelManager
	config     *config.SocialConfig

	mu sync.RWMutex
}

func NewChatActor(id string, config *config.SocialConfig, options ...events.ActorOption[*chat.ChatEvent]) *ChatActor {
	baseActor := events.NewActor[*chat.ChatEvent](id, options...)
	actor := &ChatActor{
		BaseActor:  baseActor,
		llmManager: llm.NewModelManager(config),
		config:     config,
	}

	actor.RegisterHandler(string(chat.EventTypeSendMsg), actor.handleSendMessage)
	actor.RegisterHandler(string(chat.EventTypeAIChatInterrupt), actor.handleInterrupt)
	return actor
}

func (a *ChatActor) EventBusHandler(ctx context.Context, event *chat.ChatEvent) error {
	return a.Send(ctx, event)
}

func (a *ChatActor) send(actorCtx events.ActorContext[*chat.ChatEvent], event *chat.ChatEvent) error {
	logrus.Infof("ChatActor 尝试发送事件 [%s] 类型: %s 到 outbox\n", event.EventID, event.EventType)
	err := actorCtx.Actor.PublishToOutbox(actorCtx.Context, event)
	if err != nil {
		logrus.Infof("ChatActor 发送事件到 outbox 失败: %v\n", err)
		return err
	}
	logrus.Infof("ChatActor 发送事件 [%s] 到 outbox 成功\n", event.EventID)
	return nil
}

func (a *ChatActor) sendErrorEvent(actorCtx events.ActorContext[*chat.ChatEvent], errorCode string, errorMsg string) error {
	errorEvent := &chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeError,
		Event: &chat.ErrorEvent{
			Code:    &errorCode,
			Message: errorMsg,
		},
	}

	return a.send(actorCtx, errorEvent)
}

func (a *ChatActor) handleSendMessage(actorCtx events.ActorContext[*chat.ChatEvent], event *chat.ChatEvent) error {
	// 添加调试日志
	logrus.Infof("开始处理 SendMessage 事件: %s\n", event.EventID)

	sendMsgEvent, ok := event.Event.(*chat.SendMsgEvent)
	if !ok {
		logrus.Infof("事件类型转换失败，非 SendMsgEvent 类型\n")
		return a.sendErrorEvent(actorCtx, "invalid_event", "无效的事件类型")
	}

	logrus.Infof("消息类型: %s\n", sendMsgEvent.MsgType)

	switch sendMsgEvent.MsgType {
	case messages.MessageTypeAIChat:
		aiChatBody, ok := sendMsgEvent.Body.(*chat.AIChatMsg)
		if !ok {
			logrus.Infof("消息体类型转换失败，非 AIChatMsg 类型\n")
			return a.sendErrorEvent(actorCtx, "invalid_message", "无效的消息体")
		}

		// 提取用户查询
		query := ""
		for _, item := range aiChatBody.MessageItems {
			if item.GetType() == "message" {
				for _, content := range item.(*chat.InputMessage).Content {
					if content.GetType() == "input_text" {
						query += content.(*chat.InputTextContent).Text
					}
				}
			}
		}

		logrus.Infof("提取到的用户查询: %s\n", query)

		// 使用goroutine异步处理聊天查询，避免阻塞Actor
		go a.processAIChatQuery(actorCtx, query)
		logrus.Infof("已启动异步处理 AI 聊天查询\n")
		return nil
	default:
		logrus.Infof("不支持的消息类型: %s\n", sendMsgEvent.MsgType)
		return a.sendErrorEvent(actorCtx, "unsupported_message_type", "不支持的消息类型")
	}
}

func (a *ChatActor) handleInterrupt(actorCtx events.ActorContext[*chat.ChatEvent], event *chat.ChatEvent) error {
	// 暂时留空，后续实现
	return nil
}

func (a *ChatActor) processAIChatQuery(actorCtx events.ActorContext[*chat.ChatEvent], query string) {
	logrus.Infof("开始异步处理 AI 聊天查询: %s\n", query)

	responseID := uuid.New().String()
	messageID := uuid.New().String()

	// 创建初始AI聊天消息
	aiChatMessage := &chat.AIChatMessage{
		ID:           responseID,
		MessageID:    messageID,
		Role:         chat.RoleTypeAssistant,
		Status:       chat.AiChatMessageStatusInProgress,
		UserID:       "assistant", // 这里可以替换为实际的助手ID
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
		MessageItems: []chat.OutputItem{},
	}

	logrus.Info("创建了 AI 聊天消息，准备发送 created 事件\n")

	// 发送创建事件
	createdEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatCreated,
		Event: &chat.CreatedEvent{
			Response: aiChatMessage,
		},
	}
	if err := a.send(actorCtx, &createdEvent); err != nil {
		logrus.Infof("发送创建事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送创建事件成功: %s\n", createdEvent.EventID)

	// 发送进行中事件
	inProgressEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatInProgress,
		Event: &chat.InProgressEvent{
			Response: aiChatMessage,
		},
	}
	if err := a.send(actorCtx, &inProgressEvent); err != nil {
		logrus.Infof("发送进行中事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送进行中事件成功: %s\n", inProgressEvent.EventID)

	// 创建输出消息
	outputMessage := &chat.OutputMessage{
		ID:      uuid.New().String(),
		Type:    "message",
		Role:    "assistant",
		Content: []chat.OutputContent{},
		Status:  "in_progress",
	}

	aiChatMessage.MessageItems = append(aiChatMessage.MessageItems, outputMessage)

	// 发送输出项添加事件
	outputItemAddedEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatOutputItemAdded,
		Event: &chat.OutputItemAddedEvent{
			OutputIndex: 0,
			Item:        outputMessage,
		},
	}
	if err := a.send(actorCtx, &outputItemAddedEvent); err != nil {
		logrus.Infof("发送输出项添加事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送输出项添加事件成功: %s\n", outputItemAddedEvent.EventID)

	// 创建文本内容
	textContent := &chat.OutputTextContent{
		Type: "output_text",
		Text: "",
	}

	outputMessage.Content = append(outputMessage.Content, textContent)

	// 发送内容部分添加事件
	contentPartAddedEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatContentPartAdded,
		Event: &chat.ContentPartAddedEvent{
			ItemID:       outputMessage.ID,
			OutputIndex:  0,
			ContentIndex: 0,
			Part:         textContent,
		},
	}
	if err := a.send(actorCtx, &contentPartAddedEvent); err != nil {
		logrus.Infof("发送内容部分添加事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送内容部分添加事件成功: %s\n", contentPartAddedEvent.EventID)

	// 准备LLM请求
	userMessage := llm.NewUserPromptMessage(query, "")
	promptMessages := []llm.PromptMessage{userMessage.PromptMessage}
	modelParameters := map[string]interface{}{
		"temperature": 0.7,
	}

	logrus.Infof("准备发起 LLM 请求\n")
	ctx := context.Background()

	// 发起流式聊天请求
	chatStream, err := a.llmManager.ChatStream(ctx, promptMessages, modelParameters, nil, nil)
	if err != nil {
		logrus.Infof("发起流式聊天请求失败: %v\n", err)
		// 处理错误
		aiChatMessage.Status = chat.AiChatMessageStatusFailed
		aiChatMessage.Error = &chat.ResponseError{
			Code:    "stream_error",
			Message: "Failed to stream chat completion: " + err.Error(),
		}

		failedEvent := chat.ChatEvent{
			EventID:   uuid.New().String(),
			EventType: chat.EventTypeAIChatFailed,
			Event: &chat.FailedEvent{
				Response: aiChatMessage,
			},
		}
		_ = a.send(actorCtx, &failedEvent)
		return
	}
	logrus.Info("成功启动流式聊天请求\n")

	// 处理流式响应
	logrus.Info("开始接收流式响应\n")
	for !chatStream.Closed() {
		chunk, finished, err := chatStream.Recv()
		if err != nil {
			// 如果是上下文取消或通道关闭，直接退出循环
			if err == streams.ErrContextAlreadyDone || err == streams.ErrChannelClosed {
				logrus.Infof("流已关闭或上下文已取消: %v\n", err)
				break
			}
			// 其他错误继续尝试接收
			logrus.Infof("接收流数据错误: %v，继续尝试\n", err)
			continue
		}

		if finished {
			logrus.Info("流已完成\n")
			break
		}

		delta := chunk.Delta
		// 处理文本内容
		if content, ok := delta.Message.Content.(string); ok && content != "" {
			// 更新文本内容
			textContent.Text += content
			logrus.Infof("收到文本增量: %s\n", content)

			// 发送文本增量事件
			textDeltaEvent := chat.ChatEvent{
				EventID:   uuid.New().String(),
				EventType: chat.EventTypeAIChatOutputTextDelta,
				Event: &chat.TextDeltaEvent{
					ItemID:       outputMessage.ID,
					OutputIndex:  0,
					ContentIndex: 0,
					Delta:        content,
				},
			}
			if err := a.send(actorCtx, &textDeltaEvent); err != nil {
				// 发送失败不中断流程，继续处理后续内容
				logrus.Infof("发送文本增量事件失败: %v\n", err)
				continue
			}
			logrus.Infof("发送文本增量事件成功: %s\n", textDeltaEvent.EventID)
		}

		// 处理完成信息和使用统计
		if delta.FinishReason != "" {
			logrus.Infof("收到完成原因: %s\n", delta.FinishReason)
			if delta.Usage != nil {
				logrus.Infof("使用统计: 输入=%d, 输出=%d, 总计=%d\n",
					delta.Usage.PromptTokens, delta.Usage.CompletionTokens, delta.Usage.TotalTokens)
				aiChatMessage.Usage = &chat.ResponseUsage{
					InputTokens:  delta.Usage.PromptTokens,
					OutputTokens: delta.Usage.CompletionTokens,
					TotalTokens:  delta.Usage.TotalTokens,
				}
			}
			// 收到完成原因时可以提前结束
			break
		}
	}

	logrus.Info("流式响应接收完成，准备发送完成事件\n")

	// 发送文本完成事件
	textDoneEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatOutputTextDone,
		Event: &chat.TextDoneEvent{
			ItemID:       outputMessage.ID,
			OutputIndex:  0,
			ContentIndex: 0,
			Text:         textContent.Text,
		},
	}
	if err := a.send(actorCtx, &textDoneEvent); err != nil {
		logrus.Infof("发送文本完成事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送文本完成事件成功: %s\n", textDoneEvent.EventID)

	// 发送内容部分完成事件
	contentPartDoneEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatContentPartDone,
		Event: &chat.ContentPartDoneEvent{
			ItemID:       outputMessage.ID,
			OutputIndex:  0,
			ContentIndex: 0,
			Part:         textContent,
		},
	}
	if err := a.send(actorCtx, &contentPartDoneEvent); err != nil {
		logrus.Infof("发送内容部分完成事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送内容部分完成事件成功: %s\n", contentPartDoneEvent.EventID)

	// 更新输出消息状态
	outputMessage.Status = "completed"

	// 发送输出项完成事件
	outputItemDoneEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatOutputItemDone,
		Event: &chat.OutputItemDoneEvent{
			OutputIndex: 0,
			Item:        outputMessage,
		},
	}
	if err := a.send(actorCtx, &outputItemDoneEvent); err != nil {
		logrus.Infof("发送输出项完成事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送输出项完成事件成功: %s\n", outputItemDoneEvent.EventID)

	// 更新AI聊天消息
	aiChatMessage.Status = chat.AiChatMessageStatusCompleted
	aiChatMessage.Text = textContent.Text
	aiChatMessage.UpdatedAt = time.Now().UnixMilli()

	// 如果没有获取到使用情况数据，则使用简单估算
	if aiChatMessage.Usage == nil {
		aiChatMessage.Usage = &chat.ResponseUsage{}
	}

	// 发送完成事件
	completedEvent := chat.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chat.EventTypeAIChatCompleted,
		Event: &chat.CompletedEvent{
			Response: aiChatMessage,
		},
	}
	if err := a.send(actorCtx, &completedEvent); err != nil {
		logrus.Infof("发送完成事件失败: %v\n", err)
		return
	}
	logrus.Infof("发送完成事件成功: %s\n", completedEvent.EventID)
	logrus.Info("AI 聊天查询处理完成\n")
}
