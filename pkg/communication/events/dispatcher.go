package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TracedDispatcher 是支持跟踪功能的 EventDispatcher 接口扩展
type TracedDispatcher[T Event] interface {
	EventDispatcher[T]
	SetTracer(tracer EventBusTracer[T])
}

type DefaultDispatcher[T Event] struct {
	typeHandlers map[string]map[SubscriptionID]EventHandler[T]
	mu           sync.RWMutex
	tracer       EventBusTracer[T]
}

func NewDefaultDispatcher[T Event]() *DefaultDispatcher[T] {
	return &DefaultDispatcher[T]{
		typeHandlers: make(map[string]map[SubscriptionID]EventHandler[T]),
		tracer:       NewNoOpTracer[T](),
	}
}

func (d *DefaultDispatcher[T]) SetTracer(tracer EventBusTracer[T]) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tracer = tracer
}

func (d *DefaultDispatcher[T]) Register(eventType string, handler EventHandler[T]) (SubscriptionID, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	subID := SubscriptionID(uuid.New().String())

	if _, exists := d.typeHandlers[eventType]; !exists {
		d.typeHandlers[eventType] = make(map[SubscriptionID]EventHandler[T])
	}

	d.typeHandlers[eventType][subID] = handler

	return subID, nil
}

func (d *DefaultDispatcher[T]) Unregister(subscriptionID SubscriptionID) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for eventType, handlers := range d.typeHandlers {
		if _, exists := handlers[subscriptionID]; exists {
			delete(d.typeHandlers[eventType], subscriptionID)
			if len(d.typeHandlers[eventType]) == 0 {
				delete(d.typeHandlers, eventType)
			}
			return nil
		}
	}

	return ErrSubscriptionNotFound
}

func (d *DefaultDispatcher[T]) Dispatch(ctx context.Context, event T) error {
	d.mu.RLock()

	var handlerEntries []struct {
		id      SubscriptionID
		handler EventHandler[T]
	}

	if handlers, exists := d.typeHandlers[event.Type()]; exists {
		for id, handler := range handlers {
			handlerEntries = append(handlerEntries, struct {
				id      SubscriptionID
				handler EventHandler[T]
			}{id, handler})
		}
	}

	if handlers, exists := d.typeHandlers["*"]; exists {
		for id, handler := range handlers {
			handlerEntries = append(handlerEntries, struct {
				id      SubscriptionID
				handler EventHandler[T]
			}{id, handler})
		}
	}

	tracer := d.tracer
	d.mu.RUnlock()

	var errs []error

	for _, entry := range handlerEntries {
		handlerID := string(entry.id)

		tracer.EventHandlerStarted(ctx, event, handlerID)

		startTime := time.Now()
		err := entry.handler(ctx, event)
		duration := time.Since(startTime)

		tracer.EventHandlerFinished(ctx, event, handlerID, duration, err)

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}

	return nil
}
