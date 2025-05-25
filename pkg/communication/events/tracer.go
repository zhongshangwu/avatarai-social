package events

import (
	"context"
	"fmt"
	"time"
)

// EventBusMetrics 接口定义了事件总线的指标收集功能
type EventBusMetrics interface {
	EventPublished(eventType string)
	EventProcessed(eventType string, duration time.Duration)
	EventError(eventType string, err error)
}

// EventBusTracer 接口定义了事件总线的追踪功能
type EventBusTracer[T Event] interface {
	// StartEventTrace 在事件开始处理时被调用，返回带有追踪信息的上下文
	StartEventTrace(ctx context.Context, event T) context.Context

	// EndEventTrace 在事件处理完成时被调用
	EndEventTrace(ctx context.Context, err error)

	// EventQueued 在事件被加入队列时调用
	EventQueued(ctx context.Context, event T)

	// EventPublished 在事件被发布到流时调用
	EventPublished(ctx context.Context, event T)

	// EventDispatchStarted 在事件开始分发到处理程序前调用
	EventDispatchStarted(ctx context.Context, event T)

	// EventHandlerStarted 在事件处理程序开始处理事件时调用
	EventHandlerStarted(ctx context.Context, event T, handlerID string)

	// EventHandlerFinished 在事件处理程序完成处理事件时调用
	EventHandlerFinished(ctx context.Context, event T, handlerID string, duration time.Duration, err error)

	// EventDispatchFinished 在事件分发完成时调用
	EventDispatchFinished(ctx context.Context, event T, duration time.Duration, err error)
}

// NoOpMetrics 提供一个不执行任何操作的指标收集实现
type NoOpMetrics struct{}

func (m *NoOpMetrics) EventPublished(eventType string)                         {}
func (m *NoOpMetrics) EventProcessed(eventType string, duration time.Duration) {}
func (m *NoOpMetrics) EventError(eventType string, err error)                  {}

// NoOpTracer 提供一个不执行任何操作的追踪实现
type NoOpTracer[T Event] struct{}

func NewNoOpTracer[T Event]() *NoOpTracer[T] {
	return &NoOpTracer[T]{}
}

func (t *NoOpTracer[T]) StartEventTrace(ctx context.Context, event T) context.Context       { return ctx }
func (t *NoOpTracer[T]) EndEventTrace(ctx context.Context, err error)                       {}
func (t *NoOpTracer[T]) EventQueued(ctx context.Context, event T)                           {}
func (t *NoOpTracer[T]) EventPublished(ctx context.Context, event T)                        {}
func (t *NoOpTracer[T]) EventDispatchStarted(ctx context.Context, event T)                  {}
func (t *NoOpTracer[T]) EventHandlerStarted(ctx context.Context, event T, handlerID string) {}
func (t *NoOpTracer[T]) EventHandlerFinished(ctx context.Context, event T, handlerID string, duration time.Duration, err error) {
}
func (t *NoOpTracer[T]) EventDispatchFinished(ctx context.Context, event T, duration time.Duration, err error) {
}

// LoggingTracer 实现基于日志的事件追踪
type LoggingTracer[T Event] struct {
	logFunc func(format string, args ...interface{})
}

// traceKey 是上下文中存储追踪信息的键
type traceKey struct{}

// traceInfo 存储在上下文中的追踪信息
type traceInfo struct {
	eventID   string
	eventType string
	startTime time.Time
}

// NewLoggingTracer 创建一个新的基于日志的追踪器
func NewLoggingTracer[T Event](logFunc func(format string, args ...interface{})) *LoggingTracer[T] {
	if logFunc == nil {
		logFunc = func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		}
	}

	return &LoggingTracer[T]{
		logFunc: logFunc,
	}
}

func (t *LoggingTracer[T]) StartEventTrace(ctx context.Context, event T) context.Context {
	info := &traceInfo{
		eventID:   event.ID(),
		eventType: event.Type(),
		startTime: time.Now(),
	}

	t.logFunc("开始追踪事件 [%s] 类型: %s", info.eventID, info.eventType)
	return context.WithValue(ctx, traceKey{}, info)
}

func (t *LoggingTracer[T]) EndEventTrace(ctx context.Context, err error) {
	info, ok := ctx.Value(traceKey{}).(*traceInfo)
	if !ok {
		return
	}

	duration := time.Since(info.startTime)
	if err != nil {
		t.logFunc("结束追踪事件 [%s] 类型: %s, 耗时: %v, 错误: %v",
			info.eventID, info.eventType, duration, err)
	} else {
		t.logFunc("结束追踪事件 [%s] 类型: %s, 耗时: %v",
			info.eventID, info.eventType, duration)
	}
}

func (t *LoggingTracer[T]) EventQueued(ctx context.Context, event T) {
	t.logFunc("事件队列 [%s] 类型: %s", event.ID(), event.Type())
}

func (t *LoggingTracer[T]) EventPublished(ctx context.Context, event T) {
	t.logFunc("事件发布 [%s] 类型: %s", event.ID(), event.Type())
}

func (t *LoggingTracer[T]) EventDispatchStarted(ctx context.Context, event T) {
	t.logFunc("事件开始分发 [%s] 类型: %s", event.ID(), event.Type())
}

func (t *LoggingTracer[T]) EventHandlerStarted(ctx context.Context, event T, handlerID string) {
	t.logFunc("处理程序开始 [%s] 类型: %s, 处理程序: %s", event.ID(), event.Type(), handlerID)
}

func (t *LoggingTracer[T]) EventHandlerFinished(ctx context.Context, event T, handlerID string, duration time.Duration, err error) {
	if err != nil {
		t.logFunc("处理程序完成 [%s] 类型: %s, 处理程序: %s, 耗时: %v, 错误: %v",
			event.ID(), event.Type(), handlerID, duration, err)
	} else {
		t.logFunc("处理程序完成 [%s] 类型: %s, 处理程序: %s, 耗时: %v",
			event.ID(), event.Type(), handlerID, duration)
	}
}

func (t *LoggingTracer[T]) EventDispatchFinished(ctx context.Context, event T, duration time.Duration, err error) {
	if err != nil {
		t.logFunc("事件分发完成 [%s] 类型: %s, 耗时: %v, 错误: %v",
			event.ID(), event.Type(), duration, err)
	} else {
		t.logFunc("事件分发完成 [%s] 类型: %s, 耗时: %v",
			event.ID(), event.Type(), duration)
	}
}
