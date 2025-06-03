package chat

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/agents"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
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

	ctx, cancel := context.WithTimeout(actorCtx.Context, 5*time.Minute)

	invokeCtx := agents.NewChatInvokeContext(ctx).
		WithInputItems(inputItems).
		WithAgentMessage(&respondMessage.Content.(*messages.AgentMessageContent).AgentMessage).
		WithMemory(actor.memory)

	go func() {
		defer cancel()
		actor.HandleResponseStream(actorCtx, invokeCtx.Stream)
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
