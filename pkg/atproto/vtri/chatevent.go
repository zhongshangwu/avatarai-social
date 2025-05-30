// Code generated by cmd/lexgen (see Makefile's lexgen); DO NOT EDIT.

package vtri

// schema: app.vtri.chat.event

import (
	"encoding/json"
	"fmt"

	"github.com/bluesky-social/indigo/lex/util"
)

// ChatEvent is a "main" in the app.vtri.chat.event schema.
//
// 通用聊天事件定义
type ChatEvent struct {
	Event *ChatEvent_Event `json:"event" cborgen:"event"`
	// eventId: 事件ID
	EventId string `json:"eventId" cborgen:"eventId"`
	// eventType: 事件类型
	EventType string `json:"eventType" cborgen:"eventType"`
}

// ChatEvent_AiChatMsg is a "aiChatMsg" in the app.vtri.chat.event schema.
//
// # AI聊天消息体
//
// RECORDTYPE: ChatEvent_AiChatMsg
type ChatEvent_AiChatMsg struct {
	LexiconTypeID string                              `json:"$type,const=app.vtri.chat.event#aiChatMsg" cborgen:"$type,const=app.vtri.chat.event#aiChatMsg"`
	Content       []*ChatEvent_AiChatMsg_Content_Elem `json:"content" cborgen:"content"`
	Role          string                              `json:"role" cborgen:"role"`
}

type ChatEvent_AiChatMsg_Content_Elem struct {
	ChatAiChat_InputMessage     *ChatAiChat_InputMessage
	ChatAiChat_FunctionToolCall *ChatAiChat_FunctionToolCall
}

func (t *ChatEvent_AiChatMsg_Content_Elem) MarshalJSON() ([]byte, error) {
	if t.ChatAiChat_InputMessage != nil {
		t.ChatAiChat_InputMessage.LexiconTypeID = "app.vtri.chat.aiChat#InputMessage"
		return json.Marshal(t.ChatAiChat_InputMessage)
	}
	if t.ChatAiChat_FunctionToolCall != nil {
		t.ChatAiChat_FunctionToolCall.LexiconTypeID = "app.vtri.chat.aiChat#FunctionToolCall"
		return json.Marshal(t.ChatAiChat_FunctionToolCall)
	}
	return nil, fmt.Errorf("cannot marshal empty enum")
}
func (t *ChatEvent_AiChatMsg_Content_Elem) UnmarshalJSON(b []byte) error {
	typ, err := util.TypeExtract(b)
	if err != nil {
		return err
	}

	switch typ {
	case "app.vtri.chat.aiChat#InputMessage":
		t.ChatAiChat_InputMessage = new(ChatAiChat_InputMessage)
		return json.Unmarshal(b, t.ChatAiChat_InputMessage)
	case "app.vtri.chat.aiChat#FunctionToolCall":
		t.ChatAiChat_FunctionToolCall = new(ChatAiChat_FunctionToolCall)
		return json.Unmarshal(b, t.ChatAiChat_FunctionToolCall)

	default:
		return nil
	}
}

type ChatEvent_Event struct {
	ChatEvent_SendMsgEvent                          *ChatEvent_SendMsgEvent
	ChatAiChatStream_InterruptEvent                 *ChatAiChatStream_InterruptEvent
	ChatAiChatStream_CompletedEvent                 *ChatAiChatStream_CompletedEvent
	ChatAiChatStream_ContentPartAddedEvent          *ChatAiChatStream_ContentPartAddedEvent
	ChatAiChatStream_ContentPartDoneEvent           *ChatAiChatStream_ContentPartDoneEvent
	ChatAiChatStream_CreatedEvent                   *ChatAiChatStream_CreatedEvent
	ChatAiChatStream_ErrorEvent                     *ChatAiChatStream_ErrorEvent
	ChatAiChatStream_InProgressEvent                *ChatAiChatStream_InProgressEvent
	ChatAiChatStream_FailedEvent                    *ChatAiChatStream_FailedEvent
	ChatAiChatStream_IncompleteEvent                *ChatAiChatStream_IncompleteEvent
	ChatAiChatStream_OutputItemAddedEvent           *ChatAiChatStream_OutputItemAddedEvent
	ChatAiChatStream_OutputItemDoneEvent            *ChatAiChatStream_OutputItemDoneEvent
	ChatAiChatStream_ReasoningSummaryPartAddedEvent *ChatAiChatStream_ReasoningSummaryPartAddedEvent
	ChatAiChatStream_ReasoningSummaryPartDoneEvent  *ChatAiChatStream_ReasoningSummaryPartDoneEvent
	ChatAiChatStream_ReasoningSummaryTextDeltaEvent *ChatAiChatStream_ReasoningSummaryTextDeltaEvent
	ChatAiChatStream_ReasoningSummaryTextDoneEvent  *ChatAiChatStream_ReasoningSummaryTextDoneEvent
	ChatAiChatStream_RefusalDeltaEvent              *ChatAiChatStream_RefusalDeltaEvent
	ChatAiChatStream_RefusalDoneEvent               *ChatAiChatStream_RefusalDoneEvent
	ChatAiChatStream_TextAnnotationDeltaEvent       *ChatAiChatStream_TextAnnotationDeltaEvent
	ChatAiChatStream_TextDeltaEvent                 *ChatAiChatStream_TextDeltaEvent
	ChatAiChatStream_TextDoneEvent                  *ChatAiChatStream_TextDoneEvent
}

func (t *ChatEvent_Event) MarshalJSON() ([]byte, error) {
	if t.ChatEvent_SendMsgEvent != nil {
		t.ChatEvent_SendMsgEvent.LexiconTypeID = "app.vtri.chat.event#sendMsgEvent"
		return json.Marshal(t.ChatEvent_SendMsgEvent)
	}
	if t.ChatAiChatStream_InterruptEvent != nil {
		t.ChatAiChatStream_InterruptEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#InterruptEvent"
		return json.Marshal(t.ChatAiChatStream_InterruptEvent)
	}
	if t.ChatAiChatStream_CompletedEvent != nil {
		t.ChatAiChatStream_CompletedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#CompletedEvent"
		return json.Marshal(t.ChatAiChatStream_CompletedEvent)
	}
	if t.ChatAiChatStream_ContentPartAddedEvent != nil {
		t.ChatAiChatStream_ContentPartAddedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ContentPartAddedEvent"
		return json.Marshal(t.ChatAiChatStream_ContentPartAddedEvent)
	}
	if t.ChatAiChatStream_ContentPartDoneEvent != nil {
		t.ChatAiChatStream_ContentPartDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ContentPartDoneEvent"
		return json.Marshal(t.ChatAiChatStream_ContentPartDoneEvent)
	}
	if t.ChatAiChatStream_CreatedEvent != nil {
		t.ChatAiChatStream_CreatedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#CreatedEvent"
		return json.Marshal(t.ChatAiChatStream_CreatedEvent)
	}
	if t.ChatAiChatStream_ErrorEvent != nil {
		t.ChatAiChatStream_ErrorEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ErrorEvent"
		return json.Marshal(t.ChatAiChatStream_ErrorEvent)
	}
	if t.ChatAiChatStream_InProgressEvent != nil {
		t.ChatAiChatStream_InProgressEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#InProgressEvent"
		return json.Marshal(t.ChatAiChatStream_InProgressEvent)
	}
	if t.ChatAiChatStream_FailedEvent != nil {
		t.ChatAiChatStream_FailedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#FailedEvent"
		return json.Marshal(t.ChatAiChatStream_FailedEvent)
	}
	if t.ChatAiChatStream_IncompleteEvent != nil {
		t.ChatAiChatStream_IncompleteEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#IncompleteEvent"
		return json.Marshal(t.ChatAiChatStream_IncompleteEvent)
	}
	if t.ChatAiChatStream_OutputItemAddedEvent != nil {
		t.ChatAiChatStream_OutputItemAddedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#OutputItemAddedEvent"
		return json.Marshal(t.ChatAiChatStream_OutputItemAddedEvent)
	}
	if t.ChatAiChatStream_OutputItemDoneEvent != nil {
		t.ChatAiChatStream_OutputItemDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#OutputItemDoneEvent"
		return json.Marshal(t.ChatAiChatStream_OutputItemDoneEvent)
	}
	if t.ChatAiChatStream_ReasoningSummaryPartAddedEvent != nil {
		t.ChatAiChatStream_ReasoningSummaryPartAddedEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ReasoningSummaryPartAddedEvent"
		return json.Marshal(t.ChatAiChatStream_ReasoningSummaryPartAddedEvent)
	}
	if t.ChatAiChatStream_ReasoningSummaryPartDoneEvent != nil {
		t.ChatAiChatStream_ReasoningSummaryPartDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ReasoningSummaryPartDoneEvent"
		return json.Marshal(t.ChatAiChatStream_ReasoningSummaryPartDoneEvent)
	}
	if t.ChatAiChatStream_ReasoningSummaryTextDeltaEvent != nil {
		t.ChatAiChatStream_ReasoningSummaryTextDeltaEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ReasoningSummaryTextDeltaEvent"
		return json.Marshal(t.ChatAiChatStream_ReasoningSummaryTextDeltaEvent)
	}
	if t.ChatAiChatStream_ReasoningSummaryTextDoneEvent != nil {
		t.ChatAiChatStream_ReasoningSummaryTextDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#ReasoningSummaryTextDoneEvent"
		return json.Marshal(t.ChatAiChatStream_ReasoningSummaryTextDoneEvent)
	}
	if t.ChatAiChatStream_RefusalDeltaEvent != nil {
		t.ChatAiChatStream_RefusalDeltaEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#RefusalDeltaEvent"
		return json.Marshal(t.ChatAiChatStream_RefusalDeltaEvent)
	}
	if t.ChatAiChatStream_RefusalDoneEvent != nil {
		t.ChatAiChatStream_RefusalDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#RefusalDoneEvent"
		return json.Marshal(t.ChatAiChatStream_RefusalDoneEvent)
	}
	if t.ChatAiChatStream_TextAnnotationDeltaEvent != nil {
		t.ChatAiChatStream_TextAnnotationDeltaEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#TextAnnotationDeltaEvent"
		return json.Marshal(t.ChatAiChatStream_TextAnnotationDeltaEvent)
	}
	if t.ChatAiChatStream_TextDeltaEvent != nil {
		t.ChatAiChatStream_TextDeltaEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#TextDeltaEvent"
		return json.Marshal(t.ChatAiChatStream_TextDeltaEvent)
	}
	if t.ChatAiChatStream_TextDoneEvent != nil {
		t.ChatAiChatStream_TextDoneEvent.LexiconTypeID = "app.vtri.chat.aiChatStream#TextDoneEvent"
		return json.Marshal(t.ChatAiChatStream_TextDoneEvent)
	}
	return nil, fmt.Errorf("cannot marshal empty enum")
}
func (t *ChatEvent_Event) UnmarshalJSON(b []byte) error {
	typ, err := util.TypeExtract(b)
	if err != nil {
		return err
	}

	switch typ {
	case "app.vtri.chat.event#sendMsgEvent":
		t.ChatEvent_SendMsgEvent = new(ChatEvent_SendMsgEvent)
		return json.Unmarshal(b, t.ChatEvent_SendMsgEvent)
	case "app.vtri.chat.aiChatStream#InterruptEvent":
		t.ChatAiChatStream_InterruptEvent = new(ChatAiChatStream_InterruptEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_InterruptEvent)
	case "app.vtri.chat.aiChatStream#CompletedEvent":
		t.ChatAiChatStream_CompletedEvent = new(ChatAiChatStream_CompletedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_CompletedEvent)
	case "app.vtri.chat.aiChatStream#ContentPartAddedEvent":
		t.ChatAiChatStream_ContentPartAddedEvent = new(ChatAiChatStream_ContentPartAddedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ContentPartAddedEvent)
	case "app.vtri.chat.aiChatStream#ContentPartDoneEvent":
		t.ChatAiChatStream_ContentPartDoneEvent = new(ChatAiChatStream_ContentPartDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ContentPartDoneEvent)
	case "app.vtri.chat.aiChatStream#CreatedEvent":
		t.ChatAiChatStream_CreatedEvent = new(ChatAiChatStream_CreatedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_CreatedEvent)
	case "app.vtri.chat.aiChatStream#ErrorEvent":
		t.ChatAiChatStream_ErrorEvent = new(ChatAiChatStream_ErrorEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ErrorEvent)
	case "app.vtri.chat.aiChatStream#InProgressEvent":
		t.ChatAiChatStream_InProgressEvent = new(ChatAiChatStream_InProgressEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_InProgressEvent)
	case "app.vtri.chat.aiChatStream#FailedEvent":
		t.ChatAiChatStream_FailedEvent = new(ChatAiChatStream_FailedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_FailedEvent)
	case "app.vtri.chat.aiChatStream#IncompleteEvent":
		t.ChatAiChatStream_IncompleteEvent = new(ChatAiChatStream_IncompleteEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_IncompleteEvent)
	case "app.vtri.chat.aiChatStream#OutputItemAddedEvent":
		t.ChatAiChatStream_OutputItemAddedEvent = new(ChatAiChatStream_OutputItemAddedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_OutputItemAddedEvent)
	case "app.vtri.chat.aiChatStream#OutputItemDoneEvent":
		t.ChatAiChatStream_OutputItemDoneEvent = new(ChatAiChatStream_OutputItemDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_OutputItemDoneEvent)
	case "app.vtri.chat.aiChatStream#ReasoningSummaryPartAddedEvent":
		t.ChatAiChatStream_ReasoningSummaryPartAddedEvent = new(ChatAiChatStream_ReasoningSummaryPartAddedEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ReasoningSummaryPartAddedEvent)
	case "app.vtri.chat.aiChatStream#ReasoningSummaryPartDoneEvent":
		t.ChatAiChatStream_ReasoningSummaryPartDoneEvent = new(ChatAiChatStream_ReasoningSummaryPartDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ReasoningSummaryPartDoneEvent)
	case "app.vtri.chat.aiChatStream#ReasoningSummaryTextDeltaEvent":
		t.ChatAiChatStream_ReasoningSummaryTextDeltaEvent = new(ChatAiChatStream_ReasoningSummaryTextDeltaEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ReasoningSummaryTextDeltaEvent)
	case "app.vtri.chat.aiChatStream#ReasoningSummaryTextDoneEvent":
		t.ChatAiChatStream_ReasoningSummaryTextDoneEvent = new(ChatAiChatStream_ReasoningSummaryTextDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_ReasoningSummaryTextDoneEvent)
	case "app.vtri.chat.aiChatStream#RefusalDeltaEvent":
		t.ChatAiChatStream_RefusalDeltaEvent = new(ChatAiChatStream_RefusalDeltaEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_RefusalDeltaEvent)
	case "app.vtri.chat.aiChatStream#RefusalDoneEvent":
		t.ChatAiChatStream_RefusalDoneEvent = new(ChatAiChatStream_RefusalDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_RefusalDoneEvent)
	case "app.vtri.chat.aiChatStream#TextAnnotationDeltaEvent":
		t.ChatAiChatStream_TextAnnotationDeltaEvent = new(ChatAiChatStream_TextAnnotationDeltaEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_TextAnnotationDeltaEvent)
	case "app.vtri.chat.aiChatStream#TextDeltaEvent":
		t.ChatAiChatStream_TextDeltaEvent = new(ChatAiChatStream_TextDeltaEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_TextDeltaEvent)
	case "app.vtri.chat.aiChatStream#TextDoneEvent":
		t.ChatAiChatStream_TextDoneEvent = new(ChatAiChatStream_TextDoneEvent)
		return json.Unmarshal(b, t.ChatAiChatStream_TextDoneEvent)

	default:
		return nil
	}
}

// ChatEvent_SendMsgEvent is a "sendMsgEvent" in the app.vtri.chat.event schema.
//
// 发送消息事件
//
// RECORDTYPE: ChatEvent_SendMsgEvent
type ChatEvent_SendMsgEvent struct {
	LexiconTypeID string `json:"$type,const=app.vtri.chat.event#sendMsgEvent" cborgen:"$type,const=app.vtri.chat.event#sendMsgEvent"`
	// body: 消息体
	Body *ChatEvent_SendMsgEvent_Body `json:"body" cborgen:"body"`
	// msgType: 消息类型
	MsgType int64 `json:"msgType" cborgen:"msgType"`
	// quoteId: 引用ID
	QuoteId *string `json:"quoteId,omitempty" cborgen:"quoteId,omitempty"`
	// roomId: 房间ID
	RoomId string `json:"roomId" cborgen:"roomId"`
	// senderAt: 发送时间
	SenderAt *string `json:"senderAt,omitempty" cborgen:"senderAt,omitempty"`
	// senderId: 发送者ID
	SenderId string `json:"senderId" cborgen:"senderId"`
	// threadId: 话题ID
	ThreadId *string `json:"threadId,omitempty" cborgen:"threadId,omitempty"`
}

// 消息体
type ChatEvent_SendMsgEvent_Body struct {
	ChatEvent_TextMsg   *ChatEvent_TextMsg
	ChatEvent_AiChatMsg *ChatEvent_AiChatMsg
}

func (t *ChatEvent_SendMsgEvent_Body) MarshalJSON() ([]byte, error) {
	if t.ChatEvent_TextMsg != nil {
		t.ChatEvent_TextMsg.LexiconTypeID = "app.vtri.chat.event#textMsg"
		return json.Marshal(t.ChatEvent_TextMsg)
	}
	if t.ChatEvent_AiChatMsg != nil {
		t.ChatEvent_AiChatMsg.LexiconTypeID = "app.vtri.chat.event#aiChatMsg"
		return json.Marshal(t.ChatEvent_AiChatMsg)
	}
	return nil, fmt.Errorf("cannot marshal empty enum")
}
func (t *ChatEvent_SendMsgEvent_Body) UnmarshalJSON(b []byte) error {
	typ, err := util.TypeExtract(b)
	if err != nil {
		return err
	}

	switch typ {
	case "app.vtri.chat.event#textMsg":
		t.ChatEvent_TextMsg = new(ChatEvent_TextMsg)
		return json.Unmarshal(b, t.ChatEvent_TextMsg)
	case "app.vtri.chat.event#aiChatMsg":
		t.ChatEvent_AiChatMsg = new(ChatEvent_AiChatMsg)
		return json.Unmarshal(b, t.ChatEvent_AiChatMsg)

	default:
		return nil
	}
}

// ChatEvent_TextMsg is a "textMsg" in the app.vtri.chat.event schema.
//
// 文本消息体
//
// RECORDTYPE: ChatEvent_TextMsg
type ChatEvent_TextMsg struct {
	LexiconTypeID string `json:"$type,const=app.vtri.chat.event#textMsg" cborgen:"$type,const=app.vtri.chat.event#textMsg"`
	// text: 文本消息内容
	Text string `json:"text" cborgen:"text"`
}
