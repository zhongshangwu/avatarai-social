package chat

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/agents"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

func (actor *ChatActor) AIRespond(actorCtx events.ActorContext[*messages.ChatEvent], message *messages.Message) error {
	logrus.Info("开始处理 AI 聊天消息")

	inputItems, err := actor.convertMsgToInputItems(message)
	if err != nil {
		logrus.Errorf("消息转换失败: %v", err)
		return actor.sendError(actorCtx, "conversion_failed", "消息转换失败")
	}

	respondMessage, err := actor.InitRespondMessage(message)
	if err != nil {
		logrus.Errorf("初始化响应消息失败: %v", err)
		return actor.sendError(actorCtx, "init_respond_message_failed", "初始化响应消息失败")
	}
	actor.sendMsgReceived(actorCtx, respondMessage)

	ctx, cancel := context.WithTimeout(actorCtx.Context, 5*time.Minute)

	mem := memory.NewSimpleThreadMemory(actor.DB, message.RoomID, message.ThreadID)

	invokeCtx := agents.NewChatInvokeContext(ctx).
		WithInputItems(inputItems).
		WithAgentMessage(&respondMessage.Content.(*messages.AgentMessageContent).AgentMessage).
		WithMemory(mem)

	go func() {
		defer cancel()
		actor.HandleAIResponseStream(invokeCtx)
		logrus.Info("所有响应处理完成")
	}()

	if err := actor.runner.Invoke(invokeCtx); err != nil {
		logrus.Errorf("AI 聊天智能体执行失败: %v", err)
		return actor.sendError(actorCtx, "ai_respond_failed", "AI 聊天智能体执行失败")
	} else {
		logrus.Info("AI 处理成功完成")
	}

	logrus.Info("AI 处理完成，已关闭响应流，等待响应处理完成...")
	return nil
}

func (actor *ChatActor) HandleAIResponseStream(
	invokeCtx *agents.ChatInvokeContext,
) {
	logrus.Info("开始处理响应流...")
	defer logrus.Info("响应流处理器退出")

	var currentAgentMessageID string
	if invokeCtx.Response != nil {
		currentAgentMessageID = invokeCtx.Response.ID
	}

	for {
		result := invokeCtx.Stream.Recv()

		if result.HasData {
			serverEvent := result.Data
			logrus.Infof("收到事件响应: %v", serverEvent)

			if err := actor.handleEventPersistence(serverEvent, currentAgentMessageID); err != nil {
				logrus.Errorf("持久化事件失败: %v", err)
			}

			if err := actor.PublishToOutbox(invokeCtx.Context, serverEvent); err != nil {
				logrus.Errorf("发布响应到 outbox 失败: %v", err)
				return
			}
			continue
		}

		if result.Completed {
			logrus.Info("响应流已关闭")
			if result.Error != nil {
				logrus.Errorf("接收事件响应失败: %v", result.Error)
				return
			}
			return
		}
	}
}

func (actor *ChatActor) handleEventPersistence(event *messages.ChatEvent, agentMessageID string) error {
	switch event.EventType {
	case messages.EventTypeAgentMessageCreated:
		return actor.handleAgentMessageCreated(event)
	case messages.EventTypeAgentMessageInProgress:
		return actor.handleAgentMessageInProgress(event)
	case messages.EventTypeAgentMessageCompleted:
		return actor.handleAgentMessageCompleted(event)
	case messages.EventTypeAgentMessageFailed:
		return actor.handleAgentMessageFailed(event)
	case messages.EventTypeAgentMessageIncomplete:
		return actor.handleAgentMessageIncomplete(event)
	case messages.EventTypeAgentMessageOutputItemAdded:
		return actor.handleOutputItemAdded(event, agentMessageID)
	case messages.EventTypeAgentMessageOutputItemDone:
		return actor.handleOutputItemDone(event, agentMessageID)
	default:
		// 对于不需要持久化的事件，直接返回nil
		return nil
	}
}

func (actor *ChatActor) handleAgentMessageCreated(event *messages.ChatEvent) error {
	createdEvent, ok := event.Event.(*messages.CreatedEvent)
	if !ok {
		return nil
	}

	agentMessage := createdEvent.AgentMessage
	logrus.Infof("持久化AI消息创建事件: %s", agentMessage.ID)

	return actor.MessageRepo.UpdateAgentMessageStatus(agentMessage.ID, string(agentMessage.Status))
}

func (actor *ChatActor) handleAgentMessageInProgress(event *messages.ChatEvent) error {
	inProgressEvent, ok := event.Event.(*messages.InProgressEvent)
	if !ok {
		return nil
	}

	agentMessage := inProgressEvent.AgentMessage
	logrus.Infof("持久化AI消息进行中事件: %s", agentMessage.ID)

	return actor.MessageRepo.UpdateAgentMessageStatus(agentMessage.ID, string(agentMessage.Status))
}

func (actor *ChatActor) handleAgentMessageCompleted(event *messages.ChatEvent) error {
	completedEvent, ok := event.Event.(*messages.CompletedEvent)
	if !ok {
		return nil
	}

	agentMessage := completedEvent.AgentMessage
	logrus.Infof("持久化AI消息完成事件: %s", agentMessage.ID)

	return actor.MessageRepo.UpdateAgentMessageWithUsage(
		agentMessage.ID,
		string(agentMessage.Status),
		agentMessage.Usage,
		agentMessage.AltText,
	)
}

func (actor *ChatActor) handleAgentMessageFailed(event *messages.ChatEvent) error {
	failedEvent, ok := event.Event.(*messages.FailedEvent)
	if !ok {
		return nil
	}

	agentMessage := failedEvent.AgentMessage
	logrus.Infof("持久化AI消息失败事件: %s", agentMessage.ID)

	return actor.MessageRepo.UpdateAgentMessageWithError(
		agentMessage.ID,
		string(agentMessage.Status),
		agentMessage.Error,
	)
}

func (actor *ChatActor) handleAgentMessageIncomplete(event *messages.ChatEvent) error {
	incompleteEvent, ok := event.Event.(*messages.IncompleteEvent)
	if !ok {
		return nil
	}

	agentMessage := incompleteEvent.AgentMessage
	logrus.Infof("持久化AI消息不完整事件: %s", agentMessage.ID)

	return actor.MessageRepo.UpdateAgentMessageIncomplete(
		agentMessage.ID,
		agentMessage.InterruptType,
		agentMessage.Error,
		agentMessage.IncompleteDetails,
	)
}

func (actor *ChatActor) handleOutputItemAdded(event *messages.ChatEvent, agentMessageID string) error {
	outputItemEvent, ok := event.Event.(*messages.OutputItemAddedEvent)
	if !ok {
		return nil
	}

	logrus.Infof("持久化输出项添加事件: AgentMessageID=%s, OutputIndex=%d, ItemType=%s",
		agentMessageID, outputItemEvent.OutputIndex, outputItemEvent.Item.GetType())

	itemJSON, err := json.Marshal(outputItemEvent.Item)
	if err != nil {
		logrus.Errorf("序列化输出项失败: %v", err)
		return err
	}

	agentMessageItem := &repositories.AgentMessageItem{
		ID:             uuid.New().String(),
		AgentMessageID: agentMessageID,
		ItemType:       outputItemEvent.Item.GetType(),
		Item:           string(itemJSON),
		Position:       outputItemEvent.OutputIndex,
		CreatedAt:      time.Now().UnixMilli(),
		UpdatedAt:      time.Now().UnixMilli(),
		Deleted:        false,
	}

	return actor.MessageRepo.InsertAgentMessageItem(agentMessageItem)
}

func (actor *ChatActor) handleOutputItemDone(event *messages.ChatEvent, agentMessageID string) error {
	outputItemEvent, ok := event.Event.(*messages.OutputItemDoneEvent)
	if !ok {
		return nil
	}

	logrus.Infof("持久化输出项完成事件: AgentMessageID=%s, OutputIndex=%d, ItemType=%s",
		agentMessageID, outputItemEvent.OutputIndex, outputItemEvent.Item.GetType())

	itemJSON, err := json.Marshal(outputItemEvent.Item)
	if err != nil {
		logrus.Errorf("序列化输出项失败: %v", err)
		return err
	}

	updates := map[string]interface{}{
		"item": string(itemJSON),
	}

	return actor.MessageRepo.UpdateAgentMessageItemByPosition(agentMessageID, outputItemEvent.OutputIndex, updates)
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
