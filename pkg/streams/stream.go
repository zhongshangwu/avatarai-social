package streams

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrContextAlreadyDone = errors.New("stream context already done")
	ErrChannelClosed      = errors.New("send to closed channel")
)

// Active → Closed (正常关闭)
// Active → ContextDone (上下文取消)
// Active → Error (错误状态)
type StreamState int32

const (
	StreamStateActive StreamState = iota
	StreamStateClosing
	StreamStateClosed
)

type StreamResult[T any] struct {
	Data      T     // 接收到的数据
	HasData   bool  // 是否有数据
	Completed bool  // 流是否已完成
	Error     error // 错误信息（仅在 Completed=true 时有意义）
}

type Stream[T any] struct {
	ch     chan T
	err    chan error
	closed chan struct{}
	ctx    context.Context

	state int32 // 使用 atomic 操作的状态
	once  sync.Once
}

func NewStream[T any](ctx context.Context, maxSize int) *Stream[T] {
	s := &Stream[T]{
		ch:     make(chan T, maxSize),
		err:    make(chan error, 1),
		closed: make(chan struct{}),
		ctx:    ctx,
		state:  int32(StreamStateActive),
	}

	// 监听上下文取消，自动关闭流
	go func() {
		<-ctx.Done()
		s.closeWithError(ErrContextAlreadyDone)
	}()

	return s
}

func (s *Stream[T]) Send(item T) error {
	if atomic.LoadInt32(&s.state) != int32(StreamStateActive) {
		return ErrChannelClosed
	}

	select {
	case <-s.ctx.Done():
		return ErrContextAlreadyDone
	case <-s.closed:
		return ErrChannelClosed
	case s.ch <- item:
		return nil
	}
}

func (s *Stream[T]) SendError(err error) {
	if err == nil {
		return
	}

	select {
	case <-s.ctx.Done():
		// 上下文已取消，无法发送
	case <-s.closed:
		// 流已关闭，无法发送
	case s.err <- err:
		// 成功发送错误，开始关闭流程
		s.initiateClose()
	default:
		// 错误通道已满，强制关闭
		s.closeWithError(err)
	}
}

func (s *Stream[T]) Recv() StreamResult[T] {
	var zero T
	select {
	case item, ok := <-s.ch:
		if ok {
			return StreamResult[T]{
				Data:      item,
				HasData:   true,
				Completed: false,
				Error:     nil,
			}
		}
		// 数据通道关闭，检查错误
		select {
		case err := <-s.err:
			return StreamResult[T]{
				Data:      zero,
				HasData:   false,
				Completed: true,
				Error:     err,
			}
		default:
			return StreamResult[T]{
				Data:      zero,
				HasData:   false,
				Completed: true,
				Error:     nil,
			}
		}
	case <-s.ctx.Done():
		return StreamResult[T]{
			Data:      zero,
			HasData:   false,
			Completed: true,
			Error:     ErrContextAlreadyDone,
		}
	case err := <-s.err:
		return StreamResult[T]{
			Data:      zero,
			HasData:   false,
			Completed: true,
			Error:     err,
		}
	}
}

// CloseSend 开始优雅关闭流程
func (s *Stream[T]) CloseSend() {
	s.initiateClose()
}

// ForceClose 强制关闭流，可能丢失数据
func (s *Stream[T]) ForceClose() {
	s.closeWithError(nil)
}

// initiateClose 开始关闭流程，不立即关闭数据通道
func (s *Stream[T]) initiateClose() {
	if !atomic.CompareAndSwapInt32(&s.state, int32(StreamStateActive), int32(StreamStateClosing)) {
		return // 已经在关闭过程中
	}

	go func() {
		// 发送关闭信号
		select {
		case <-s.closed:
			// 已经关闭
		default:
			close(s.closed)
		}

		// 等待一小段时间让接收方处理剩余数据
		// 然后关闭数据通道
		s.once.Do(func() {
			close(s.ch)
			atomic.StoreInt32(&s.state, int32(StreamStateClosed))
		})
	}()
}

// closeWithError 带错误的强制关闭
func (s *Stream[T]) closeWithError(err error) {
	if !atomic.CompareAndSwapInt32(&s.state, int32(StreamStateActive), int32(StreamStateClosed)) {
		return // 已经关闭
	}

	s.once.Do(func() {
		// 如果有错误，尝试发送
		if err != nil {
			select {
			case s.err <- err:
			default:
				// 错误通道满了，忽略
			}
		}

		// 关闭所有通道
		select {
		case <-s.closed:
		default:
			close(s.closed)
		}
		close(s.ch)
		close(s.err)
	})
}

func (s *Stream[T]) Closed() bool {
	state := atomic.LoadInt32(&s.state)
	return state != int32(StreamStateActive)
}

func (s *Stream[T]) State() StreamState {
	return StreamState(atomic.LoadInt32(&s.state))
}
