package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

var (
	ErrActorStopped     = errors.New("actor already stopped")
	ErrActorNotStarted  = errors.New("actor not started")
	ErrHandlerNotFound  = errors.New("event handler not found")
	ErrDuplicateHandler = errors.New("duplicate event handler")
	ErrInvalidEvent     = errors.New("invalid event")
	ErrActorBusy        = errors.New("actor busy, inbox full")
	ErrOutboxBusy       = errors.New("actor busy, outbox full")
	ErrActorTimeout     = errors.New("actor operation timeout")
	ErrNoEvents         = errors.New("no events available")
)

type ActorContext[T Event] struct {
	Actor   Actor[T]
	Context context.Context
}

type Handler[T Event] func(actorCtx ActorContext[T], msg T) error

type Actor[T Event] interface {
	ID() string
	Send(ctx context.Context, msg T) error
	SendWithTimeout(ctx context.Context, msg T, timeout time.Duration) error
	RegisterHandler(msgType string, handler Handler[T]) error
	Start(ctx context.Context) error
	Stop() error
	IsStopped() bool
	PublishToOutbox(ctx context.Context, msg T) error
	ReceiveFromOutbox(ctx context.Context) (T, error)
	ReceiveFromOutboxWithTimeout(ctx context.Context, timeout time.Duration) (T, error)
}

type BaseActor[T Event] struct {
	id       string
	inbox    *streams.Stream[T]
	outbox   *streams.Stream[T]
	handlers map[string]Handler[T]

	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	mu        sync.RWMutex
	stopped   bool
	started   bool
	inboxCap  int
	outboxCap int

	// 可选配置
	errorHandler func(error)
	middleware   []func(Handler[T]) Handler[T]
	gracePeriod  time.Duration // 优雅关闭的等待时间

	// 自定义 Stream
	customInbox  *streams.Stream[T]
	customOutbox *streams.Stream[T]
}

type ActorOption[T Event] func(*BaseActor[T])

func ActorWithErrorHandler[T Event](handler func(error)) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.errorHandler = handler
	}
}

func ActorWithMiddleware[T Event](middleware ...func(Handler[T]) Handler[T]) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.middleware = append(a.middleware, middleware...)
	}
}

func ActorWithInboxCapacity[T Event](capacity int) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.inboxCap = capacity
	}
}

func ActorWithOutboxCapacity[T Event](capacity int) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.outboxCap = capacity
	}
}

func ActorWithGracePeriod[T Event](period time.Duration) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.gracePeriod = period
	}
}

// ActorWithCustomInbox 允许使用自定义的 Stream 作为 Actor 的收件箱
// 注意：
// 1. 提供的 Stream 应当在调用此函数前已经创建并初始化
// 2. Actor 不会主动关闭自定义 Stream，而是依赖 Stream 自身监听 Actor 的上下文取消信号
// 3. 如果需要在 Actor 停止前后控制 Stream 的生命周期，应当由 Stream 的创建者负责
func ActorWithCustomInbox[T Event](inbox *streams.Stream[T]) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.customInbox = inbox
	}
}

// ActorWithCustomOutbox 允许使用自定义的 Stream 作为 Actor 的发件箱
// 注意：
// 1. 提供的 Stream 应当在调用此函数前已经创建并初始化
// 2. Actor 不会主动关闭自定义 Stream，而是依赖 Stream 自身监听 Actor 的上下文取消信号
// 3. 如果需要在 Actor 停止前后控制 Stream 的生命周期，应当由 Stream 的创建者负责
func ActorWithCustomOutbox[T Event](outbox *streams.Stream[T]) ActorOption[T] {
	return func(a *BaseActor[T]) {
		a.customOutbox = outbox
	}
}

func NewActor[T Event](id string, options ...ActorOption[T]) *BaseActor[T] {
	actor := &BaseActor[T]{
		id:          id,
		handlers:    make(map[string]Handler[T]),
		stopped:     false,
		started:     false,
		inboxCap:    100,             // 默认收件箱容量
		outboxCap:   100,             // 默认发件箱容量
		gracePeriod: 5 * time.Second, // 默认优雅关闭等待时间
		errorHandler: func(err error) {
			// 默认错误处理器，简单记录错误
			logrus.Infof("Actor %s error: %v\n", id, err)
		},
	}

	for _, option := range options {
		option(actor)
	}

	return actor
}

func (a *BaseActor[T]) ID() string {
	return a.id
}

func (a *BaseActor[T]) Send(ctx context.Context, msg T) error {
	logrus.Infof("Actor %s 接收到消息: %v\n", a.id, msg)

	a.mu.RLock()
	if a.stopped {
		a.mu.RUnlock()
		logrus.Infof("Actor %s 已停止，消息被拒绝\n", a.id)
		return ErrActorStopped
	}

	if !a.started {
		a.mu.RUnlock()
		logrus.Infof("Actor %s 未启动，消息被拒绝\n", a.id)
		return ErrActorNotStarted
	}
	a.mu.RUnlock()

	logrus.Infof("Actor %s 发送消息到收件箱\n", a.id)
	err := a.inbox.Send(ctx, msg)
	if err != nil {
		if errors.Is(err, streams.ErrChannelClosed) {
			logrus.Infof("Actor %s 收件箱已关闭\n", a.id)
			return ErrActorStopped
		}
		if errors.Is(err, streams.ErrContextAlreadyDone) {
			logrus.Infof("Actor %s 上下文已取消\n", a.id)
			return ErrActorStopped
		}
		logrus.Infof("Actor %s 发送消息失败: %v\n", a.id, err)
		return fmt.Errorf("failed to send event: %w", err)
	}
	logrus.Infof("Actor %s 消息已成功发送到收件箱\n", a.id)
	return nil
}

func (a *BaseActor[T]) SendWithTimeout(ctx context.Context, msg T, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- a.Send(ctx, msg)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return ErrActorTimeout
	}
}

func (a *BaseActor[T]) PublishToOutbox(ctx context.Context, msg T) error {
	logrus.Infof("Actor %s 尝试发布消息到 outbox: %v\n", a.id, msg)

	a.mu.RLock()
	if a.stopped {
		a.mu.RUnlock()
		logrus.Infof("Actor %s 已停止，消息被拒绝\n", a.id)
		return ErrActorStopped
	}

	if !a.started {
		a.mu.RUnlock()
		logrus.Infof("Actor %s 未启动，消息被拒绝\n", a.id)
		return ErrActorNotStarted
	}
	a.mu.RUnlock()

	logrus.Infof("Actor %s 发送消息到 outbox\n", a.id)
	err := a.outbox.Send(ctx, msg)
	if err != nil {
		if errors.Is(err, streams.ErrChannelClosed) {
			logrus.Infof("Actor %s outbox 已关闭\n", a.id)
			return ErrActorStopped
		}
		if errors.Is(err, streams.ErrContextAlreadyDone) {
			logrus.Infof("Actor %s 上下文已取消\n", a.id)
			return ErrActorStopped
		}
		logrus.Infof("Actor %s 发布到 outbox 失败: %v\n", a.id, err)
		return fmt.Errorf("failed to publish to outbox: %w", err)
	}
	logrus.Infof("Actor %s 成功发布消息到 outbox\n", a.id)
	return nil
}

func (a *BaseActor[T]) ReceiveFromOutbox(ctx context.Context) (T, error) {
	var zero T

	a.mu.RLock()
	if a.stopped {
		a.mu.RUnlock()
		return zero, ErrActorStopped
	}

	if !a.started {
		a.mu.RUnlock()
		return zero, ErrActorNotStarted
	}
	a.mu.RUnlock()

	msg, finished, err := a.outbox.Recv()
	if err != nil {
		return zero, fmt.Errorf("failed to receive from outbox: %w", err)
	}

	if finished {
		return zero, ErrActorStopped
	}

	return msg, nil
}

func (a *BaseActor[T]) ReceiveFromOutboxWithTimeout(ctx context.Context, timeout time.Duration) (T, error) {
	var zero T

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		msg T
		err error
	}

	done := make(chan result, 1)

	go func() {
		msg, err := a.ReceiveFromOutbox(ctx)
		done <- result{msg, err}
	}()

	select {
	case res := <-done:
		return res.msg, res.err
	case <-timeoutCtx.Done():
		return zero, ErrActorTimeout
	}
}

func (a *BaseActor[T]) RegisterHandler(msgType string, handler Handler[T]) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.handlers[msgType]; exists {
		return fmt.Errorf("%w: type %s", ErrDuplicateHandler, msgType)
	}

	wrappedHandler := handler
	for i := len(a.middleware) - 1; i >= 0; i-- {
		wrappedHandler = a.middleware[i](wrappedHandler)
	}

	a.handlers[msgType] = wrappedHandler
	return nil
}

func (a *BaseActor[T]) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return nil // 已经启动，不需要再次启动
	}

	a.ctx, a.cancelFunc = context.WithCancel(ctx)

	// 使用自定义 inbox 或创建新的
	if a.customInbox != nil {
		a.inbox = a.customInbox
	} else {
		a.inbox = streams.NewStream[T](a.ctx, a.inboxCap)
	}

	// 使用自定义 outbox 或创建新的
	if a.customOutbox != nil {
		a.outbox = a.customOutbox
	} else {
		a.outbox = streams.NewStream[T](a.ctx, a.outboxCap)
	}

	a.started = true
	a.stopped = false
	a.mu.Unlock()

	a.wg.Add(1)
	go a.processLoop()

	return nil
}

func (a *BaseActor[T]) processLoop() {
	defer a.wg.Done()

	for {
		msg, finished, err := a.inbox.Recv()
		if finished {
			// 流已关闭，退出处理循环
			return
		}

		if err != nil {
			a.handleError(fmt.Errorf("error receiving event: %w", err))
			continue
		}

		msgType := msg.Type()
		a.mu.RLock()
		handler, exists := a.handlers[msgType]
		a.mu.RUnlock()

		if !exists {
			a.handleError(fmt.Errorf("%w: %s", ErrHandlerNotFound, msgType))
			continue
		}

		// 顺序处理模式
		func(m T, h Handler[T]) {
			defer func() {
				if r := recover(); r != nil {
					a.handleError(fmt.Errorf("panic in event handler: %v", r))
				}
			}()

			actorCtx := ActorContext[T]{
				Actor:   a,
				Context: a.ctx,
			}

			err := h(actorCtx, m)
			if err != nil {
				a.handleError(err)
			}
		}(msg, handler)
	}
}

func (a *BaseActor[T]) handleError(err error) {
	if a.errorHandler != nil {
		a.errorHandler(err)
	}
}

func (a *BaseActor[T]) Stop() error {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return nil // 已经停止，不需要再次停止
	}

	if !a.started {
		a.stopped = true
		a.mu.Unlock()
		return nil
	}

	a.stopped = true
	a.started = false
	a.mu.Unlock()

	// 取消上下文，这会间接触发 Stream 的关闭
	// Stream 在创建时注册了对上下文的监听，当上下文取消时会自动调用 CloseSend()
	// 对于自定义和非自定义的 Stream 都是一样的处理机制
	a.cancelFunc()

	// 等待所有处理中的消息完成
	waitCh := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(waitCh)
	}()

	// 使用配置的优雅关闭等待时间
	select {
	case <-waitCh:
		// 正常完成
	case <-time.After(a.gracePeriod):
		// 超时，强制退出
		logrus.Infof("Actor %s shutdown timed out after %v, some events may not be processed\n", a.id, a.gracePeriod)
	}

	return nil
}

func (a *BaseActor[T]) IsStopped() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.stopped
}
