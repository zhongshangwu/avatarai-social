package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

var (
	ErrEventBusStopped      = errors.New("event bus already stopped")
	ErrEventBusNotStarted   = errors.New("event bus not started")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidEventType     = errors.New("invalid event type")
	ErrEventPublishTimeout  = errors.New("event publish timeout")
)

type Event interface {
	ID() string
	Type() string
}

type EventHandler[T Event] func(ctx context.Context, event T) error

type EventBus[T Event] interface {
	Publish(ctx context.Context, event T) error
	Subscribe(eventType string, handler EventHandler[T]) (SubscriptionID, error)
	Unsubscribe(subscriptionID SubscriptionID) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Wait() error
}

type SubscriptionID string

type EventDispatcher[T Event] interface {
	Dispatch(ctx context.Context, event T) error
	Register(eventType string, handler EventHandler[T]) (SubscriptionID, error)
	Unregister(subscriptionID SubscriptionID) error
}

type DefaultEventBus[T Event] struct {
	eventStream *streams.Stream[T]
	dispatcher  EventDispatcher[T]

	// 配置选项
	bufferSize  int
	workerCount int

	// 状态管理
	mu      sync.RWMutex
	started bool
	stopped bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// 可观测性
	metrics      EventBusMetrics
	tracer       EventBusTracer[T]
	errorHandler func(error)
}

type EventBusOption[T Event] func(*DefaultEventBus[T])

func NewEventBus[T Event](options ...EventBusOption[T]) *DefaultEventBus[T] {
	bus := &DefaultEventBus[T]{
		bufferSize:  1000, // 默认缓冲区大小
		workerCount: 10,   // 默认工作线程数
		started:     false,
		stopped:     false,
		errorHandler: func(err error) {
			fmt.Printf("EventBus error: %v\n", err)
		},
		metrics: &NoOpMetrics{},
		tracer:  NewNoOpTracer[T](),
	}

	for _, option := range options {
		option(bus)
	}

	dispatcher := NewDefaultDispatcher[T]()
	dispatcher.SetTracer(bus.tracer)
	bus.dispatcher = dispatcher

	return bus
}

func (eb *DefaultEventBus[T]) Start(ctx context.Context) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.started {
		return nil // 已经启动
	}

	if eb.stopped {
		return ErrEventBusStopped
	}

	eb.ctx, eb.cancel = context.WithCancel(ctx)
	eb.eventStream = streams.NewStream[T](eb.ctx, eb.bufferSize)
	eb.started = true

	// 启动工作线程池
	for i := 0; i < eb.workerCount; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}

	return nil
}

func (eb *DefaultEventBus[T]) worker(id int) {
	defer eb.wg.Done()

	for {
		event, finished, err := eb.eventStream.Recv()
		if finished {
			return // 流已关闭，退出工作线程
		}

		if err != nil {
			eb.handleError(fmt.Errorf("worker %d error receiving event: %w", id, err))
			continue
		}

		// 追踪事件处理开始
		ctx := eb.tracer.StartEventTrace(eb.ctx, event)

		// 记录分发开始
		eb.tracer.EventDispatchStarted(ctx, event)
		start := time.Now()

		// 分发事件
		err = eb.dispatcher.Dispatch(ctx, event)

		// 计算处理时间
		duration := time.Since(start)

		// 记录分发完成
		eb.tracer.EventDispatchFinished(ctx, event, duration, err)

		// 记录指标
		eb.metrics.EventProcessed(event.Type(), duration)

		if err != nil {
			eb.handleError(err)
			eb.metrics.EventError(event.Type(), err)
		}

		// 追踪事件处理结束
		eb.tracer.EndEventTrace(ctx, err)
	}
}

func (eb *DefaultEventBus[T]) Stop(ctx context.Context) error {
	eb.mu.Lock()
	if eb.stopped {
		eb.mu.Unlock()
		return nil // 已经停止
	}

	if !eb.started {
		eb.stopped = true
		eb.mu.Unlock()
		return nil
	}

	// 标记为停止状态，阻止新的事件发布
	eb.stopped = true
	eb.mu.Unlock()

	// 关闭事件流，停止接收新的事件
	eb.eventStream.CloseSend()

	// 等待所有工作线程完成当前处理
	waitCh := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(waitCh)
	}()

	// 使用提供的上下文超时
	select {
	case <-waitCh:
		// 所有事件处理完成后，才取消上下文
		eb.cancel()
		eb.mu.Lock()
		eb.started = false
		eb.mu.Unlock()
		return nil
	case <-ctx.Done():
		// 超时，强制退出
		eb.cancel()
		return fmt.Errorf("event bus shutdown timed out: %w", ctx.Err())
	}
}

func (eb *DefaultEventBus[T]) Wait() error {
	eb.wg.Wait()
	return nil
}

func (eb *DefaultEventBus[T]) Publish(ctx context.Context, event T) error {
	eb.mu.RLock()
	if eb.stopped {
		eb.mu.RUnlock()
		return ErrEventBusStopped
	}

	if !eb.started {
		eb.mu.RUnlock()
		return ErrEventBusNotStarted
	}
	eb.mu.RUnlock()

	// 记录事件被加入队列
	eb.tracer.EventQueued(ctx, event)

	// 记录发布指标
	eb.metrics.EventPublished(event.Type())

	// 发送事件到流
	err := eb.eventStream.Send(event)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// 记录事件已发布
	eb.tracer.EventPublished(ctx, event)

	return nil
}

func (eb *DefaultEventBus[T]) PublishWithTimeout(ctx context.Context, event T, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- eb.Publish(ctx, event)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return ErrEventPublishTimeout
	}
}

func (eb *DefaultEventBus[T]) Subscribe(eventType string, handler EventHandler[T]) (SubscriptionID, error) {
	if eventType == "" {
		return "", ErrInvalidEventType
	}

	return eb.dispatcher.Register(eventType, handler)
}

func (eb *DefaultEventBus[T]) Unsubscribe(subscriptionID SubscriptionID) error {
	return eb.dispatcher.Unregister(subscriptionID)
}

func (eb *DefaultEventBus[T]) handleError(err error) {
	if eb.errorHandler != nil {
		eb.errorHandler(err)
	}
}

func BusWithBufferSize[T Event](size int) EventBusOption[T] {
	return func(eb *DefaultEventBus[T]) {
		eb.bufferSize = size
	}
}

func BusWithWorkerCount[T Event](count int) EventBusOption[T] {
	return func(eb *DefaultEventBus[T]) {
		eb.workerCount = count
	}
}

func BusWithErrorHandler[T Event](handler func(error)) EventBusOption[T] {
	return func(eb *DefaultEventBus[T]) {
		eb.errorHandler = handler
	}
}

func BusWithMetrics[T Event](metrics EventBusMetrics) EventBusOption[T] {
	return func(eb *DefaultEventBus[T]) {
		eb.metrics = metrics
	}
}

func BusWithTracer[T Event](tracer EventBusTracer[T]) EventBusOption[T] {
	return func(eb *DefaultEventBus[T]) {
		eb.tracer = tracer
		// 如果 dispatcher 实现了 TracedDispatcher 接口，则设置 tracer
		if td, ok := eb.dispatcher.(TracedDispatcher[T]); ok && td != nil {
			td.SetTracer(tracer)
		}
	}
}
