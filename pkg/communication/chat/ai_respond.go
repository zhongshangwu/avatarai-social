package chat

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/agents"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
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

	invokeCtx := agents.NewChatInvokeContext(ctx).
		WithInputItems(inputItems).
		WithAgentMessage(&respondMessage.Content.(*messages.AgentMessageContent).AgentMessage).
		WithMemory(actor.memory)

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

	for {
		result := invokeCtx.Stream.Recv()

		if result.HasData {
			serverEvent := result.Data
			logrus.Infof("收到事件响应: %v", serverEvent)
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
