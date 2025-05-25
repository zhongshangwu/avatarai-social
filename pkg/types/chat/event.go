package chat

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

type ChatEvent struct {
	EventID   string        `json:"eventId"`   // 事件ID
	EventType ChatEventType `json:"eventType"` // 事件类型
	Event     ChatEventBody `json:"event"`     // 事件内容
}

func (c *ChatEvent) ID() string {
	return c.EventID
}

func (c *ChatEvent) Type() string {
	return string(c.EventType)
}

type ChatEventBody interface {
	isChatEventBody()
}

func (e *ChatEvent) UnmarshalJSON(data []byte) error {
	type Alias ChatEvent
	aux := &struct {
		Event json.RawMessage `json:"event"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 根据EventType确定Event的具体类型
	var eventBody ChatEventBody
	switch e.EventType {
	case EventTypeSendMsg:
		eventBody = &SendMsgEvent{}
	case EventTypeAIChatInterrupt:
		eventBody = &InterruptEvent{}
	case EventTypeAIChatCompleted:
		eventBody = &CompletedEvent{}
	case EventTypeAIChatContentPartAdded:
		eventBody = &ContentPartAddedEvent{}
	case EventTypeAIChatContentPartDone:
		eventBody = &ContentPartDoneEvent{}
	case EventTypeAIChatCreated:
		eventBody = &CreatedEvent{}
	case EventTypeAIChatInProgress:
		eventBody = &InProgressEvent{}
	case EventTypeAIChatFailed:
		eventBody = &FailedEvent{}
	case EventTypeAIChatIncomplete:
		eventBody = &IncompleteEvent{}
	case EventTypeAIChatOutputItemAdded:
		eventBody = &OutputItemAddedEvent{}
	case EventTypeAIChatOutputItemDone:
		eventBody = &OutputItemDoneEvent{}
	case EventTypeAIChatReasoningSummaryPartAdded:
		eventBody = &ReasoningSummaryPartAddedEvent{}
	case EventTypeAIChatReasoningSummaryPartDone:
		eventBody = &ReasoningSummaryPartDoneEvent{}
	case EventTypeAIChatReasoningSummaryTextDelta:
		eventBody = &ReasoningSummaryTextDeltaEvent{}
	case EventTypeAIChatReasoningSummaryTextDone:
		eventBody = &ReasoningSummaryTextDoneEvent{}
	case EventTypeAIChatRefusalDelta:
		eventBody = &RefusalDeltaEvent{}
	case EventTypeAIChatRefusalDone:
		eventBody = &RefusalDoneEvent{}
	case EventTypeAIChatOutputTextAnnotationAdded:
		eventBody = &TextAnnotationDeltaEvent{}
	case EventTypeAIChatOutputTextDelta:
		eventBody = &TextDeltaEvent{}
	case EventTypeAIChatOutputTextDone:
		eventBody = &TextDoneEvent{}
	default:
		return fmt.Errorf("unknown event type: %s", e.EventType)
	}

	if err := json.Unmarshal(aux.Event, eventBody); err != nil {
		return err
	}

	e.Event = eventBody
	return nil
}

func (e *ChatEvent) MarshalJSON() ([]byte, error) {
	type Alias ChatEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}

type SendMsgEvent struct {
	RoomID   string               `json:"roomId"`             // 房间ID
	MsgType  messages.MessageType `json:"msgType"`            // 消息类型
	Body     SendMsgBody          `json:"body"`               // 消息体
	SenderID string               `json:"senderId"`           // 发送者ID
	ThreadID string               `json:"threadId,omitempty"` // 话题ID
	QuoteID  string               `json:"quoteId,omitempty"`  // 引用ID
	SenderAt int64                `json:"senderAt,omitempty"` // 发送时间
}

func (s *SendMsgEvent) isChatEventBody() {}

func (s *SendMsgEvent) UnmarshalJSON(data []byte) error {
	type Alias SendMsgEvent
	aux := &struct {
		Body json.RawMessage `json:"body"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 尝试解析为TextMsg
	textMsg := &TextMsg{}
	if err := json.Unmarshal(aux.Body, textMsg); err == nil {
		// 检查是否真的是TextMsg（至少有text字段）
		var check struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(aux.Body, &check); err == nil && check.Text != "" {
			s.Body = textMsg
			return nil
		} else {
			logrus.WithError(err).Errorf("failed to unmarshal message body3")
		}
	}

	// 尝试解析为AIChatMsg
	aiChatMsg := &AIChatMsg{}
	if err := json.Unmarshal(aux.Body, aiChatMsg); err == nil {
		s.Body = aiChatMsg
		return nil
	} else {
		logrus.WithError(err).Errorf("failed to unmarshal message body1")
	}

	return fmt.Errorf("failed to unmarshal message body")
}

type SendMsgBody interface {
	isSendMsgBody()
}

type TextMsg struct {
	Text string `json:"text"` // 文本消息内容
}

func (t *TextMsg) isSendMsgBody() {}

type AIChatMsg struct {
	MessageItems []InputItem `json:"messageItems"` // 内容可以是InputMessage或FunctionToolCall
}

func (a *AIChatMsg) isSendMsgBody() {}

func (a *AIChatMsg) UnmarshalJSON(data []byte) error {
	type Alias AIChatMsg
	aux := &struct {
		MessageItems []json.RawMessage `json:"messageItems"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.MessageItems = make([]InputItem, len(aux.MessageItems))
	for i, raw := range aux.MessageItems {
		// 尝试解析为InputMessage
		inputMsg := &InputMessage{}
		if err := json.Unmarshal(raw, inputMsg); err == nil {
			a.MessageItems[i] = inputMsg
			continue
		}

		// 尝试解析为FunctionToolCall
		toolCall := &FunctionToolCall{}
		if err := json.Unmarshal(raw, toolCall); err == nil {
			a.MessageItems[i] = toolCall
			continue
		}

		return fmt.Errorf("failed to unmarshal content item at index %d", i)
	}

	return nil
}
