package chat

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/agents"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

type ChatActor struct {
	*events.BaseActor[*messages.ChatEvent]
	llmManager *llm.ModelManager
	config     *config.SocialConfig

	runner *agents.ChatRunner
	memory *memory.Memory
	mu     sync.RWMutex
}

func NewChatActor(id string, config *config.SocialConfig, options ...events.ActorOption[*messages.ChatEvent]) *ChatActor {
	baseActor := events.NewActor[*messages.ChatEvent](id, options...)
	llmManager := llm.NewModelManager(config)

	llm.RegisterDefaultTools(llmManager)

	actor := &ChatActor{
		BaseActor:  baseActor,
		llmManager: llmManager,
		config:     config,
	}
	runner := agents.NewChatRunner(
		actor.llmManager,
	)
	actor.runner = runner
	actor.RegisterHandler(string(messages.EventTypeSendMsg), actor.SendMsgHandler)
	actor.RegisterHandler(string(messages.EventTypeAIChatInterrupt), actor.InterruptHandler)
	return actor
}

func (actor *ChatActor) SendMsgHandler(actorCtx events.ActorContext[*messages.ChatEvent], event *messages.ChatEvent) error {
	logrus.Infof("开始处理 SendMessage 事件: %s", event.EventID)

	sendMsgEvent, ok := event.Event.(*messages.SendMsgEvent)
	if !ok {
		logrus.Error("事件类型转换失败，非 SendMsgEvent 类型")
		return actor.sendError(actorCtx, "invalid_event", "无效的事件类型")
	}

	logrus.Infof("消息类型: %s", sendMsgEvent.MsgType)

	inputItems, err := actor.convertMsgToInputItems(sendMsgEvent)
	if err != nil {
		logrus.Errorf("消息转换失败: %v", err)
		return actor.sendError(actorCtx, "conversion_failed", "消息转换失败")
	}

	actor.AIRespond(actorCtx, inputItems)
	logrus.Info("已启动异步处理 AI 聊天消息")
	return nil
}

func (actor *ChatActor) InterruptHandler(actorCtx events.ActorContext[*messages.ChatEvent], event *messages.ChatEvent) error {
	logrus.Info("收到手动中断事件，打断中...")
	interruptEvent, ok := event.Event.(*messages.InterruptEvent)
	if !ok {
		logrus.Error("事件类型转换失败，非 InterruptEvent 类型")
		return actor.sendError(actorCtx, "invalid_event", "无效的事件类型")
	}

	ctrlCtx := agents.NewChatControlContext(actorCtx.Context, agents.CtrlTypeInterrupt, interruptEvent.ResponseID)
	actor.runner.Ctrl(ctrlCtx)
	return nil
}

func (actor *ChatActor) AIRespond(actorCtx events.ActorContext[*messages.ChatEvent], inputItems []messages.InputItem) {
	logrus.Info("开始处理 AI 聊天消息")
	responseID := "default-response-id"
	messageID := "default-message-id"
	response := &messages.AIChatMessage{
		ID:           responseID,
		MID:          messageID,
		Role:         messages.RoleTypeAssistant,
		Status:       messages.AiChatMessageStatusInProgress,
		Creator:      "assistant",
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
		MessageItems: []messages.MessageItem{},
		Tools:        actor.extractTools(),
		Metadata:     make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(actorCtx.Context, 5*time.Minute)

	invokeCtx := agents.NewChatInvokeContext(ctx).
		WithInputItems(inputItems).
		WithAIChatMessage(response).
		WithMemory(actor.memory)

	go func() {
		defer cancel()
		actor.HandleResponseStream(actorCtx, invokeCtx.Stream)
		logrus.Info("所有响应处理完成")
	}()

	if err := actor.runner.Invoke(invokeCtx); err != nil {
		logrus.Errorf("AI 聊天智能体执行失败: %v", err)
	} else {
		logrus.Info("AI 处理成功完成")
	}

	logrus.Info("AI 处理完成，已关闭响应流，等待响应处理完成...")
}

func (actor *ChatActor) HandleResponseStream(
	actorCtx events.ActorContext[*messages.ChatEvent],
	responseStream *streams.Stream[*messages.ChatEvent],
) {
	logrus.Info("开始处理响应流...")
	defer logrus.Info("响应流处理器退出")

	for {
		serverEvent, closed, err := responseStream.Recv()
		if err != nil {
			if errors.Is(err, streams.ErrContextAlreadyDone) {
				logrus.Info("响应流上下文已取消")
				return
			}
			logrus.Errorf("接收事件响应失败: %v", err)
			return
		}

		logrus.Infof("收到事件响应: %v", serverEvent)

		if err := actor.PublishToOutbox(actorCtx.Context, serverEvent); err != nil {
			logrus.Errorf("发布响应到 outbox 失败: %v", err)
			return
		}

		if closed {
			logrus.Info("事件响应流发送通道已关闭")
			return
		}
	}
}

func (actor *ChatActor) sendError(actorCtx events.ActorContext[*messages.ChatEvent], errorCode string, errorMsg string) error {
	errorEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeError,
		Event: &messages.ErrorEvent{
			Code:    &errorCode,
			Message: errorMsg,
		},
	}
	logrus.Infof("发送错误事件 [%s]", errorEvent.EventID)
	return actor.PublishToOutbox(actorCtx.Context, errorEvent)
}

func (actor *ChatActor) extractTools() []map[string]interface{} {
	availableTools := actor.llmManager.GetAvailableTools()
	var tools []map[string]interface{}
	for _, tool := range availableTools {
		toolMap := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
		tools = append(tools, toolMap)
	}
	return tools
}
