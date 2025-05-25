package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/chat"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
	chatTypes "github.com/zhongshangwu/avatarai-social/pkg/types/chat"
)

var (
	upgrader = websocket.Upgrader{}
)

type EventHandler func(event chatTypes.ChatEvent, conn *websocket.Conn) error

func (a *AvatarAIAPI) ChatStream(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	logrus.Info("ChatStream connected")
	ctx, cancel := context.WithCancel(c.Request().Context())

	// 创建 LoggingTracer 用于跟踪事件
	tracer := events.NewLoggingTracer[*chatTypes.ChatEvent](func(format string, args ...interface{}) {
		logrus.Infof("[ChatEventTracer] "+format, args...)
	})

	eventBus := events.NewEventBus[*chatTypes.ChatEvent](
		events.BusWithBufferSize[*chatTypes.ChatEvent](100),
		events.BusWithWorkerCount[*chatTypes.ChatEvent](1),
		events.BusWithErrorHandler[*chatTypes.ChatEvent](func(err error) {
			logrus.Errorf("ChatStream event bus error: %v", err)
		}),
		events.BusWithTracer[*chatTypes.ChatEvent](tracer),
	)
	eventBus.Start(ctx)

	respStream := streams.NewStream[*chatTypes.ChatEvent](ctx, 100)

	go func() {
		handleChatStreamResponse(ctx, respStream, conn)
		cancel()
	}()

	chatActor := chat.NewChatActor("chat", a.Config,
		events.ActorWithCustomOutbox[*chatTypes.ChatEvent](respStream),
	)
	chatActor.Start(ctx)
	defer chatActor.Stop()

	if _, err := eventBus.Subscribe(string(chatTypes.EventTypeSendMsg), chatActor.Send); err != nil {
		logrus.Errorf("ChatStream subscribe event error: %v", err)
	}

	// 添加调试日志，确认订阅是否成功
	logrus.Infof("已成功订阅 SendMsg 事件到 ChatActor")

	for {
		msgType, msg, err := conn.ReadMessage()
		logrus.Infof("ChatStream msgType: %d, msg: %s, err: %v", msgType, string(msg), err)
		if err != nil {
			// 处理连接关闭或错误
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket错误: %v", err)
			}
		}

		if msgType == websocket.TextMessage {
			var event chatTypes.ChatEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				logrus.Errorf("ChatStream event unmarshal error: %v", err)
				sendErrorEvent(conn, "invalid_event_format", "无效的事件格式")
				continue
			}

			logrus.Infof("ChatStream event: %+v", event)
			switch event.EventType {
			case chatTypes.EventTypeSendMsg:
				logrus.Infof("event type: %s", event.EventType)

				// 调试输出事件结构
				if msgEvent, ok := event.Event.(*chatTypes.SendMsgEvent); ok {
					logrus.Infof("SendMsgEvent 解析成功: msgType=%s", msgEvent.MsgType)

					// 检查消息体类型
					switch body := msgEvent.Body.(type) {
					case *chatTypes.AIChatMsg:
						logrus.Infof("AIChatMsg 消息体: 包含 %d 个消息项", len(body.MessageItems))
					case *chatTypes.TextMsg:
						logrus.Infof("TextMsg 消息体: %s", body.Text)
					default:
						logrus.Warnf("未知的消息体类型: %T", msgEvent.Body)
					}
				} else {
					logrus.Warnf("事件解析失败，event.Event 不是 SendMsgEvent 类型")
				}

				logrus.Infof("开始发布事件到 EventBus: %s", event.EventID)
				if err := eventBus.Publish(ctx, &event); err != nil {
					logrus.Errorf("ChatStream publish event error: %v", err)
					continue
				}
				logrus.Infof("事件已发布到 EventBus: %s", event.EventID)
			}
		}
	}
}

func handleChatStreamResponse(ctx context.Context, stream *streams.Stream[*chatTypes.ChatEvent], conn *websocket.Conn) {
	logrus.Info("启动 handleChatStreamResponse 处理器")
	defer stream.CloseSend()

	for {
		logrus.Info("等待接收事件响应...")
		serverEvent, closed, err := stream.Recv()
		if errors.Is(err, streams.ErrContextAlreadyDone) || closed {
			logrus.Info("ChatStream recv response finished")
			return
		}
		if err != nil {
			logrus.Errorf("ChatStream recv response error: %v", err)
			return
		}

		logrus.Infof("收到响应事件: ID=%s, 类型=%s", serverEvent.EventID, serverEvent.EventType)

		data, err := json.Marshal(serverEvent)
		if err != nil {
			logrus.Errorf("ChatStream response marshal error: %v", err)
			return
		}

		logrus.Infof("发送响应到客户端: %s", string(data))
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			logrus.Errorf("ChatStream write response error: %v", err)
			return
		}
		logrus.Info("响应已发送到客户端")
	}
}

func sendErrorEvent(conn *websocket.Conn, errorCode string, errorMsg string) error {
	errorEvent := chatTypes.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: chatTypes.EventTypeError,
		Event: &chatTypes.ErrorEvent{
			Code:    &errorCode,
			Message: errorMsg,
		},
	}

	data, err := json.Marshal(errorEvent)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

func (a *AvatarAIAPI) ChatHistoryMessages(c echo.Context) error {
	return nil
}
