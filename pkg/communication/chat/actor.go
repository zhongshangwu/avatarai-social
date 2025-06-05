package chat

import (
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/agents"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
	"gorm.io/gorm"
)

type ChatActor struct {
	*events.BaseActor[*messages.ChatEvent]

	User         *database.Avatar       // 发送者用户
	OAuthSession *database.OAuthSession // 发送者 OAuth 会话

	DB         *gorm.DB
	llmManager *llm.ModelManager
	config     *config.SocialConfig

	runner *agents.ChatRunner
	memory memory.Memory
	mu     sync.RWMutex
}

func NewChatActor(
	id string,
	db *gorm.DB,
	config *config.SocialConfig,
	options ...events.ActorOption[*messages.ChatEvent],
) *ChatActor {
	baseActor := events.NewActor[*messages.ChatEvent](id, options...)
	llmManager := llm.NewModelManager(config)

	llm.RegisterDefaultTools(llmManager)

	actor := &ChatActor{
		BaseActor:  baseActor,
		DB:         db,
		llmManager: llmManager,
		config:     config,
	}
	runner := agents.NewChatRunner(
		actor.llmManager,
	)
	actor.runner = runner
	actor.RegisterHandler(string(messages.EventTypeMessageSend), actor.SendMsgHandler)
	actor.RegisterHandler(string(messages.EventTypeAgentMessageInterrupt), actor.InterruptHandler)
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

	message, err := actor.SendMsg(sendMsgEvent)
	if err != nil {
		logrus.Errorf("消息发送失败: %v", err)
		return actor.sendError(actorCtx, "send_failed", "消息发送失败")
	}
	actor.sendMsgSent(actorCtx, message, event)

	actor.AIRespond(actorCtx, message)
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

	ctrlCtx := agents.NewChatControlContext(actorCtx.Context, agents.CtrlTypeInterrupt, interruptEvent.AgentMessageID)
	actor.runner.Ctrl(ctrlCtx)
	return nil
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

func (actor *ChatActor) sendMsgSent(actorCtx events.ActorContext[*messages.ChatEvent], message *messages.Message, event *messages.ChatEvent) error {
	sentEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeMessageSent,
		Event: &messages.MessageSentEvent{
			MessageID: message.ID,
			EventID:   event.EventID,
		},
	}
	logrus.Infof("发送消息已发送事件: %s", sentEvent.EventID)
	return actor.PublishToOutbox(actorCtx.Context, sentEvent)
}

func (actor *ChatActor) sendMsgReceived(actorCtx events.ActorContext[*messages.ChatEvent], message *messages.Message) error {
	receivedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeMessageReceived,
		Event: &messages.MessageReceivedEvent{
			Message: message,
		},
	}
	logrus.Infof("发送消息已接收事件: %s", receivedEvent.EventID)
	return actor.PublishToOutbox(actorCtx.Context, receivedEvent)
}
