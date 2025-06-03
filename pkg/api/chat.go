package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/chat"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/events"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 在生产环境中应该进行适当的来源检查
		},
		EnableCompression: true,
	}
)

type EventHandler func(event messages.ChatEvent, conn *websocket.Conn) error

func (a *AvatarAIAPI) ChatStream(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	logrus.Info("ChatStream connected")

	connCtx, connCancel := context.WithCancel(c.Request().Context())
	defer connCancel()

	tracer := events.NewLoggingTracer[*messages.ChatEvent](func(format string, args ...interface{}) {
		logrus.Infof("[ChatEventTracer] "+format, args...)
	})

	eventBus := events.NewEventBus[*messages.ChatEvent](
		events.BusWithBufferSize[*messages.ChatEvent](100),
		events.BusWithWorkerCount[*messages.ChatEvent](1),
		events.BusWithErrorHandler[*messages.ChatEvent](func(err error) {
			logrus.Errorf("ChatStream event bus error: %v", err)
		}),
		events.BusWithTracer[*messages.ChatEvent](tracer),
	)

	if err := eventBus.Start(connCtx); err != nil {
		logrus.Errorf("Failed to start event bus: %v", err)
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		eventBus.Stop(shutdownCtx)
	}()

	outbox := streams.NewStream[*messages.ChatEvent](connCtx, 100)
	defer outbox.CloseSend()

	chatActor := chat.NewChatActor("chat",
		a.metaStore.DB,
		a.Config,
		events.ActorWithCustomOutbox[*messages.ChatEvent](outbox),
	)

	if err := chatActor.Start(connCtx); err != nil {
		logrus.Errorf("Failed to start chat actor: %v", err)
		return err
	}
	defer chatActor.Stop()

	if _, err := eventBus.Subscribe(string(messages.EventTypeSendMsg), chatActor.Send); err != nil {
		logrus.Errorf("ChatStream subscribe event error: %v", err)
		sendErrorEvent(conn, "subscribe_event_error", "订阅事件失败")
		return err
	}
	if _, err := eventBus.Subscribe(string(messages.EventTypeAgentMessageInterrupt), chatActor.Send); err != nil {
		logrus.Errorf("ChatStream subscribe event error: %v", err)
		sendErrorEvent(conn, "subscribe_event_error", "订阅事件失败")
		return err
	}

	go func() {
		a.handleChatStreamResponse(connCtx, outbox, conn)
	}()

	for {
		select {
		case <-connCtx.Done():
			logrus.Info("连接上下文已取消，退出消息处理循环")
			return nil
		default:
			// conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Infof("WebSocket连接正常关闭: %v", err)
					return nil
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					logrus.Errorf("WebSocket意外错误: %v", err)
				}
				return err
			}

			if msgType == websocket.TextMessage {
				if err := a.handleWebSocketMessage(connCtx, eventBus, conn, msg); err != nil {
					logrus.Errorf("处理 WebSocket 消息失败: %v", err)
					continue
				}
			}
		}
	}
}

func (a *AvatarAIAPI) handleWebSocketMessage(
	ctx context.Context,
	eventBus events.EventBus[*messages.ChatEvent],
	conn *websocket.Conn,
	msg []byte,
) error {
	logrus.Infof("ChatStream msgType: TextMessage, msg: %s", string(msg))

	var event messages.ChatEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		logrus.Errorf("ChatStream event unmarshal error: %v", err)
		sendErrorEvent(conn, "invalid_event_format", "无效的事件格式")
		return err
	}

	eventCtx := context.WithValue(ctx, "eventID", event.EventID)
	eventCtx = context.WithValue(eventCtx, "eventType", event.EventType)
	eventCtx, cancel := context.WithTimeout(eventCtx, 30*time.Second)
	defer cancel()

	logrus.Infof("ChatStream event: %+v", event)
	logrus.Infof("开始发布事件到 EventBus: %s", event.EventID)

	if err := eventBus.Publish(eventCtx, &event); err != nil {
		logrus.Errorf("ChatStream publish event error: %v", err)
		return err
	}

	logrus.Infof("事件已发布到 EventBus: %s", event.EventID)
	return nil
}

func (a *AvatarAIAPI) handleChatStreamResponse(ctx context.Context, outbox *streams.Stream[*messages.ChatEvent], conn *websocket.Conn) {
	logrus.Info("启动 handleChatStreamResponse 处理器")
	defer logrus.Info("handleChatStreamResponse 处理器退出")

	for {
		serverEvent, closed, err := outbox.Recv()
		if closed {
			logrus.Info("ChatStream recv response finished - stream closed")
			return
		}
		if errors.Is(err, streams.ErrContextAlreadyDone) {
			logrus.Info("ChatStream recv response finished - context done")
			return
		}
		if err != nil {
			logrus.Errorf("ChatStream recv response error: %v", err)
			return
		}

		if serverEvent == nil {
			logrus.Info("ChatStream recv response finished - serverEvent is nil")
			continue
		}

		logrus.Infof("收到响应事件: ID=%s, 类型=%s", serverEvent.EventID, serverEvent.EventType)

		if err := a.sendResponseToClient(ctx, conn, serverEvent); err != nil {
			logrus.Errorf("发送响应到客户端失败: %v", err)
			return
		}
	}
}

func (a *AvatarAIAPI) sendResponseToClient(ctx context.Context, conn *websocket.Conn, event *messages.ChatEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Errorf("ChatStream response marshal error: %v", err)
		return err
	}

	logrus.Infof("发送响应到客户端: %s", string(data))

	// conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			logrus.Info("客户端连接已关闭，停止发送响应")
		} else {
			logrus.Errorf("ChatStream write response error: %v", err)
		}
		return err
	}

	logrus.Info("响应已发送到客户端")
	return nil
}

func sendErrorEvent(conn *websocket.Conn, errorCode string, errorMsg string) error {
	errorEvent := messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeError,
		Event: &messages.ErrorEvent{
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
