package messages

import (
	"encoding/json"
	"fmt"
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
	case EventTypeAgentMessageInterrupt:
		eventBody = &InterruptEvent{}
	case EventTypeAgentMessageCompleted:
		eventBody = &CompletedEvent{}
	case EventTypeAgentMessageContentPartAdded:
		eventBody = &ContentPartAddedEvent{}
	case EventTypeAgentMessageContentPartDone:
		eventBody = &ContentPartDoneEvent{}
	case EventTypeAgentMessageCreated:
		eventBody = &CreatedEvent{}
	case EventTypeAgentMessageInProgress:
		eventBody = &InProgressEvent{}
	case EventTypeAgentMessageFailed:
		eventBody = &FailedEvent{}
	case EventTypeAgentMessageIncomplete:
		eventBody = &IncompleteEvent{}
	case EventTypeAgentMessageOutputItemAdded:
		eventBody = &OutputItemAddedEvent{}
	case EventTypeAgentMessageOutputItemDone:
		eventBody = &OutputItemDoneEvent{}
	case EventTypeAgentMessageReasoningSummaryPartAdded:
		eventBody = &ReasoningSummaryPartAddedEvent{}
	case EventTypeAgentMessageReasoningSummaryPartDone:
		eventBody = &ReasoningSummaryPartDoneEvent{}
	case EventTypeAgentMessageReasoningSummaryTextDelta:
		eventBody = &ReasoningSummaryTextDeltaEvent{}
	case EventTypeAgentMessageReasoningSummaryTextDone:
		eventBody = &ReasoningSummaryTextDoneEvent{}
	case EventTypeAgentMessageRefusalDelta:
		eventBody = &RefusalDeltaEvent{}
	case EventTypeAgentMessageRefusalDone:
		eventBody = &RefusalDoneEvent{}
	case EventTypeAgentMessageOutputTextAnnotationAdded:
		eventBody = &TextAnnotationDeltaEvent{}
	case EventTypeAgentMessageOutputTextDelta:
		eventBody = &TextDeltaEvent{}
	case EventTypeAgentMessageOutputTextDone:
		eventBody = &TextDoneEvent{}
	case EventTypeAgentMessageFunctionCallArgumentsDelta:
		eventBody = &FunctionCallArgumentsDeltaEvent{}
	case EventTypeAgentMessageFunctionCallArgumentsDone:
		eventBody = &FunctionCallArgumentsDoneEvent{}
	case EventTypeAgentMessageFileSearchCallInProgress:
		eventBody = &FileSearchCallInProgressEvent{}
	case EventTypeAgentMessageFileSearchCallSearching:
		eventBody = &FileSearchCallSearchingEvent{}
	case EventTypeAgentMessageFileSearchCallCompleted:
		eventBody = &FileSearchCallCompletedEvent{}
	case EventTypeAgentMessageWebSearchCallInProgress:
		eventBody = &WebSearchCallInProgressEvent{}
	case EventTypeAgentMessageWebSearchCallSearching:
		eventBody = &WebSearchCallSearchingEvent{}
	case EventTypeAgentMessageWebSearchCallCompleted:
		eventBody = &WebSearchCallCompletedEvent{}
	case EventTypeAgentMessageCodeInterpreterCallInProgress:
		eventBody = &CodeInterpreterCallInProgressEvent{}
	case EventTypeAgentMessageCodeInterpreterCallInterpreting:
		eventBody = &CodeInterpreterCallInterpretingEvent{}
	case EventTypeAgentMessageCodeInterpreterCallCompleted:
		eventBody = &CodeInterpreterCallCompletedEvent{}
	case EventTypeAgentMessageCodeInterpreterCallCodeDelta:
		eventBody = &CodeInterpreterCallCodeDeltaEvent{}
	case EventTypeAgentMessageCodeInterpreterCallCodeDone:
		eventBody = &CodeInterpreterCallCodeDoneEvent{}
	case EventTypeAgentMessageComputerCallInProgress:
		eventBody = &ComputerCallInProgressEvent{}
	case EventTypeAgentMessageComputerCallCompleted:
		eventBody = &ComputerCallCompletedEvent{}
	case EventTypeAgentMessageAudioDelta:
		eventBody = &AudioDeltaEvent{}
	case EventTypeAgentMessageAudioDone:
		eventBody = &AudioDoneEvent{}
	case EventTypeAgentMessageAudioTranscriptDelta:
		eventBody = &AudioTranscriptDeltaEvent{}
	case EventTypeAgentMessageAudioTranscriptDone:
		eventBody = &AudioTranscriptDoneEvent{}
	default:
		return fmt.Errorf("未知的事件类型: %s", e.EventType)
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
