package agents

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

type ChatInvokeContext struct {
	Context     context.Context
	ControlChan chan CtrlType

	Memory     *memory.Memory
	InputItems []messages.InputItem
	Response   *messages.AIChatMessage

	CurrentOutputItemIdx int
	CurrentOutputMessage *messages.OutputMessage
	CurrentTextContent   *messages.OutputTextContent

	Stream *streams.Stream[*messages.ChatEvent]
	mu     sync.RWMutex
}

func NewChatInvokeContext(
	ctx context.Context,
) *ChatInvokeContext {
	return &ChatInvokeContext{
		Context:     ctx,
		ControlChan: make(chan CtrlType, 10), // 增加缓冲区大小
		Stream:      streams.NewStream[*messages.ChatEvent](ctx, 100),
		mu:          sync.RWMutex{},
	}
}

func (c *ChatInvokeContext) isInvokeContext() {}

func (c *ChatInvokeContext) WithInputItems(items []messages.InputItem) *ChatInvokeContext {
	c.InputItems = items
	return c
}

func (c *ChatInvokeContext) WithMemory(memory *memory.Memory) *ChatInvokeContext {
	c.Memory = memory
	return c
}

func (c *ChatInvokeContext) WithAIChatMessage(message *messages.AIChatMessage) *ChatInvokeContext {
	c.Response = message
	return c
}

func (c *ChatInvokeContext) send(event *messages.ChatEvent) error {
	logrus.Infof("ChatInvokeContext 尝试发送事件 [%s] 类型: %s", event.EventID, event.EventType)
	err := c.Stream.Send(c.Context, event)
	if err != nil {
		logrus.Errorf("ChatInvokeContext 发送事件失败: %v", err)
		return err
	}
	logrus.Infof("ChatInvokeContext 发送事件 [%s] 成功", event.EventID)
	return nil
}

func (c *ChatInvokeContext) sendError(errorCode string, errorMsg string) error {
	errorEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeError,
		Event: &messages.ErrorEvent{
			Code:    &errorCode,
			Message: errorMsg,
		},
	}
	return c.Stream.Send(c.Context, errorEvent)
}

func (c *ChatInvokeContext) sendAIChatCreated(message *messages.AIChatMessage) error {
	createdEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatCreated,
		Event: &messages.CreatedEvent{
			Response: message,
		},
	}
	return c.Stream.Send(c.Context, createdEvent)
}

func (c *ChatInvokeContext) sendAIChatInProgress(message *messages.AIChatMessage) error {
	inProgressEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatInProgress,
		Event: &messages.InProgressEvent{
			Response: message,
		},
	}
	return c.Stream.Send(c.Context, inProgressEvent)
}

func (c *ChatInvokeContext) sendAIChatFailed(message *messages.AIChatMessage, errorCode, errorMsg string) error {
	message.Status = messages.AiChatMessageStatusFailed
	message.Error = &messages.ResponseError{
		Code:    messages.ResponseErrorCode(errorCode),
		Message: errorMsg,
	}

	failedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatFailed,
		Event: &messages.FailedEvent{
			Response: message,
		},
	}
	return c.Stream.Send(c.Context, failedEvent)
}

func (c *ChatInvokeContext) sendAIChatCompleted(message *messages.AIChatMessage) error {
	message.Status = messages.AiChatMessageStatusCompleted
	message.UpdatedAt = time.Now().UnixMilli()

	completedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatCompleted,
		Event: &messages.CompletedEvent{
			Response: message,
		},
	}
	return c.Stream.Send(c.Context, completedEvent)
}

func (c *ChatInvokeContext) sendOutputItemAdded(outputIndex int, item messages.OutputItem) error {
	outputItemAddedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatOutputItemAdded,
		Event: &messages.OutputItemAddedEvent{
			OutputIndex: outputIndex,
			Item:        item,
		},
	}
	return c.Stream.Send(c.Context, outputItemAddedEvent)
}

func (c *ChatInvokeContext) sendOutputItemDone(outputIndex int, item messages.OutputItem) error {
	outputItemDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatOutputItemDone,
		Event: &messages.OutputItemDoneEvent{
			OutputIndex: outputIndex,
			Item:        item,
		},
	}
	return c.Stream.Send(c.Context, outputItemDoneEvent)
}

func (c *ChatInvokeContext) sendContentPartAdded(itemID string, outputIndex, contentIndex int, part messages.OutputContent) error {
	contentPartAddedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatContentPartAdded,
		Event: &messages.ContentPartAddedEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Part:         part,
		},
	}
	return c.Stream.Send(c.Context, contentPartAddedEvent)
}

func (c *ChatInvokeContext) sendContentPartDone(itemID string, outputIndex, contentIndex int, part messages.OutputContent) error {
	contentPartDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatContentPartDone,
		Event: &messages.ContentPartDoneEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Part:         part,
		},
	}
	return c.Stream.Send(c.Context, contentPartDoneEvent)
}

func (c *ChatInvokeContext) sendTextDelta(itemID string, outputIndex, contentIndex int, delta string) error {
	textDeltaEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatOutputTextDelta,
		Event: &messages.TextDeltaEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Delta:        delta,
		},
	}
	return c.Stream.Send(c.Context, textDeltaEvent)
}

func (c *ChatInvokeContext) sendTextDone(itemID string, outputIndex, contentIndex int, text string) error {
	textDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatOutputTextDone,
		Event: &messages.TextDoneEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Text:         text,
		},
	}
	return c.Stream.Send(c.Context, textDoneEvent)
}

func (c *ChatInvokeContext) sendFunctionCallArgumentsDelta(itemID string, outputIndex int, delta string) error {
	argsDeltaEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatFunctionCallArgumentsDelta,
		Event: &messages.FunctionCallArgumentsDeltaEvent{
			ItemID:      itemID,
			OutputIndex: outputIndex,
			Delta:       delta,
		},
	}
	return c.Stream.Send(c.Context, argsDeltaEvent)
}

func (c *ChatInvokeContext) sendFunctionCallArgumentsDone(itemID string, outputIndex int, arguments string) error {
	argsDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatFunctionCallArgumentsDone,
		Event: &messages.FunctionCallArgumentsDoneEvent{
			ItemID:      itemID,
			OutputIndex: outputIndex,
			Arguments:   arguments,
		},
	}
	return c.Stream.Send(c.Context, argsDoneEvent)
}

func (c *ChatInvokeContext) sendAIChatIncomplete(message *messages.AIChatMessage) error {
	incompleteEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAIChatIncomplete,
		Event: &messages.IncompleteEvent{
			Response: message,
		},
	}
	return c.Stream.Send(c.Context, incompleteEvent)
}

type ChatControlContext struct {
	Context    context.Context
	CtrlType   CtrlType
	ResponseID string
}

func NewChatControlContext(ctx context.Context, ctrlType CtrlType, responseID string) *ChatControlContext {
	return &ChatControlContext{
		Context:    ctx,
		CtrlType:   ctrlType,
		ResponseID: responseID,
	}
}
