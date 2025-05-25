package streams

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrContextAlreadyDone = errors.New("stream context already done")
	ErrChannelClosed      = errors.New("send to closed channel")
)

type Stream[T any] struct {
	ch     chan T
	err    chan error
	closed chan struct{}
	ctx    context.Context
	cancel context.CancelFunc

	once sync.Once
}

func NewStream[T any](ctx context.Context, maxSize int) *Stream[T] {
	ctx, cancel := context.WithCancel(ctx)
	s := &Stream[T]{
		ch:     make(chan T, maxSize),
		err:    make(chan error, 1), // 保持错误通道容量为1
		closed: make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}

	// 监听上下文取消，自动关闭流
	go func() {
		<-ctx.Done()
		s.CloseSend()
	}()

	return s
}

func (s *Stream[T]) Send(item T) error {
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
	select {
	case <-s.ctx.Done():
		// 上下文已取消，无法发送
	case <-s.closed:
		// 流已关闭，无法发送
	case s.err <- err:
		// 成功发送错误
	}
}

func (s *Stream[T]) Recv() (item T, finished bool, err error) {
	var zero T
	select {
	case <-s.ctx.Done():
		return zero, true, ErrContextAlreadyDone
	case item, ok := <-s.ch:
		if !ok {
			// 数据通道已关闭，检查是否有错误
			select {
			case err := <-s.err:
				return zero, true, err
			default:
				return zero, true, nil
			}
		}
		return item, false, nil
	case err := <-s.err:
		return zero, true, err
	}
}

func (s *Stream[T]) CloseSend() {
	s.once.Do(func() {
		s.cancel() // 取消上下文
		close(s.closed)
		close(s.ch)
		close(s.err)
	})
}

func (s *Stream[T]) Closed() bool {
	select {
	case <-s.closed:
		return true
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}
