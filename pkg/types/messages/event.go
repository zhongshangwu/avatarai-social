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
	case EventTypeAIChatFunctionCallArgumentsDelta:
		eventBody = &FunctionCallArgumentsDeltaEvent{}
	case EventTypeAIChatFunctionCallArgumentsDone:
		eventBody = &FunctionCallArgumentsDoneEvent{}
	case EventTypeAIChatFileSearchCallInProgress:
		eventBody = &FileSearchCallInProgressEvent{}
	case EventTypeAIChatFileSearchCallSearching:
		eventBody = &FileSearchCallSearchingEvent{}
	case EventTypeAIChatFileSearchCallCompleted:
		eventBody = &FileSearchCallCompletedEvent{}
	case EventTypeAIChatWebSearchCallInProgress:
		eventBody = &WebSearchCallInProgressEvent{}
	case EventTypeAIChatWebSearchCallSearching:
		eventBody = &WebSearchCallSearchingEvent{}
	case EventTypeAIChatWebSearchCallCompleted:
		eventBody = &WebSearchCallCompletedEvent{}
	case EventTypeAIChatCodeInterpreterCallInProgress:
		eventBody = &CodeInterpreterCallInProgressEvent{}
	case EventTypeAIChatCodeInterpreterCallInterpreting:
		eventBody = &CodeInterpreterCallInterpretingEvent{}
	case EventTypeAIChatCodeInterpreterCallCompleted:
		eventBody = &CodeInterpreterCallCompletedEvent{}
	case EventTypeAIChatCodeInterpreterCallCodeDelta:
		eventBody = &CodeInterpreterCallCodeDeltaEvent{}
	case EventTypeAIChatCodeInterpreterCallCodeDone:
		eventBody = &CodeInterpreterCallCodeDoneEvent{}
	case EventTypeAIChatComputerCallInProgress:
		eventBody = &ComputerCallInProgressEvent{}
	case EventTypeAIChatComputerCallCompleted:
		eventBody = &ComputerCallCompletedEvent{}
	case EventTypeAIChatAudioDelta:
		eventBody = &AudioDeltaEvent{}
	case EventTypeAIChatAudioDone:
		eventBody = &AudioDoneEvent{}
	case EventTypeAIChatAudioTranscriptDelta:
		eventBody = &AudioTranscriptDeltaEvent{}
	case EventTypeAIChatAudioTranscriptDone:
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
