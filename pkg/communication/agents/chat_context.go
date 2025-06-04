package agents

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type ChatInvokeContext struct {
	Context     context.Context
	ControlChan chan CtrlType

	Memory     *memory.Memory
	InputItems []messages.InputItem
	Response   *messages.AgentMessage

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

func (c *ChatInvokeContext) WithAgentMessage(message *messages.AgentMessage) *ChatInvokeContext {
	c.Response = message
	return c
}

func (c *ChatInvokeContext) send(event *messages.ChatEvent) error {
	logrus.Infof("ChatInvokeContext 尝试发送事件 [%s] 类型: %s", event.EventID, event.EventType)
	err := c.Stream.Send(event)
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
	return c.Stream.Send(errorEvent)
}

func (c *ChatInvokeContext) sendAIChatCreated(message *messages.AgentMessage) error {
	createdEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageCreated,
		Event: &messages.CreatedEvent{
			AgentMessage: message,
		},
	}
	return c.Stream.Send(createdEvent)
}

func (c *ChatInvokeContext) sendAIChatInProgress(message *messages.AgentMessage) error {
	inProgressEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageInProgress,
		Event: &messages.InProgressEvent{
			AgentMessage: message,
		},
	}
	return c.Stream.Send(inProgressEvent)
}

func (c *ChatInvokeContext) sendAIChatFailed(message *messages.AgentMessage, errorCode messages.ResponseErrorCode, errorMsg string) error {
	message.Status = messages.AgentMessageStatusFailed
	message.Error = &messages.ResponseError{
		Code:    errorCode,
		Message: errorMsg,
	}

	failedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageFailed,
		Event: &messages.FailedEvent{
			AgentMessage: message,
		},
	}
	return c.Stream.Send(failedEvent)
}

func (c *ChatInvokeContext) sendAIChatCompleted(message *messages.AgentMessage) error {
	message.Status = messages.AgentMessageStatusCompleted
	message.UpdatedAt = time.Now().UnixMilli()

	completedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageCompleted,
		Event: &messages.CompletedEvent{
			AgentMessage: message,
		},
	}
	return c.Stream.Send(completedEvent)
}

func (c *ChatInvokeContext) sendOutputItemAdded(outputIndex int, item messages.OutputItem) error {
	outputItemAddedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageOutputItemAdded,
		Event: &messages.OutputItemAddedEvent{
			OutputIndex: outputIndex,
			Item:        item,
		},
	}
	return c.Stream.Send(outputItemAddedEvent)
}

func (c *ChatInvokeContext) sendOutputItemDone(outputIndex int, item messages.OutputItem) error {
	outputItemDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageOutputItemDone,
		Event: &messages.OutputItemDoneEvent{
			OutputIndex: outputIndex,
			Item:        item,
		},
	}
	return c.Stream.Send(outputItemDoneEvent)
}

func (c *ChatInvokeContext) sendContentPartAdded(itemID string, outputIndex, contentIndex int, part messages.OutputContent) error {
	contentPartAddedEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageContentPartAdded,
		Event: &messages.ContentPartAddedEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Part:         part,
		},
	}
	return c.Stream.Send(contentPartAddedEvent)
}

func (c *ChatInvokeContext) sendContentPartDone(itemID string, outputIndex, contentIndex int, part messages.OutputContent) error {
	contentPartDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageContentPartDone,
		Event: &messages.ContentPartDoneEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Part:         part,
		},
	}
	return c.Stream.Send(contentPartDoneEvent)
}

func (c *ChatInvokeContext) sendTextDelta(itemID string, outputIndex, contentIndex int, delta string) error {
	textDeltaEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageOutputTextDelta,
		Event: &messages.TextDeltaEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Delta:        delta,
		},
	}
	return c.Stream.Send(textDeltaEvent)
}

func (c *ChatInvokeContext) sendTextDone(itemID string, outputIndex, contentIndex int, text string) error {
	textDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageOutputTextDone,
		Event: &messages.TextDoneEvent{
			ItemID:       itemID,
			OutputIndex:  outputIndex,
			ContentIndex: contentIndex,
			Text:         text,
		},
	}
	return c.Stream.Send(textDoneEvent)
}

func (c *ChatInvokeContext) sendFunctionCallArgumentsDelta(itemID string, outputIndex int, delta string) error {
	argsDeltaEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageFunctionCallArgumentsDelta,
		Event: &messages.FunctionCallArgumentsDeltaEvent{
			ItemID:      itemID,
			OutputIndex: outputIndex,
			Delta:       delta,
		},
	}
	return c.Stream.Send(argsDeltaEvent)
}

func (c *ChatInvokeContext) sendFunctionCallArgumentsDone(itemID string, outputIndex int, arguments string) error {
	argsDoneEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageFunctionCallArgumentsDone,
		Event: &messages.FunctionCallArgumentsDoneEvent{
			ItemID:      itemID,
			OutputIndex: outputIndex,
			Arguments:   arguments,
		},
	}
	return c.Stream.Send(argsDoneEvent)
}

func (c *ChatInvokeContext) sendAIChatIncomplete(message *messages.AgentMessage) error {
	incompleteEvent := &messages.ChatEvent{
		EventID:   uuid.New().String(),
		EventType: messages.EventTypeAgentMessageIncomplete,
		Event: &messages.IncompleteEvent{
			AgentMessage: message,
		},
	}
	return c.Stream.Send(incompleteEvent)
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
