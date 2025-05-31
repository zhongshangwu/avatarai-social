package streams

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrContextAlreadyDone     = errors.New("stream context already done")
	ErrChannelClosed          = errors.New("send to closed channel")
	ErrSendContextTimeout     = errors.New("stream send timeout")
	ErrSendContextCanceled    = errors.New("stream send canceled")
	ErrSendContextAlreadyDone = errors.New("stream send already done")
)

type Stream[T any] struct {
	ch     chan T
	err    chan error
	closed chan struct{}
	ctx    context.Context

	once sync.Once
}

func NewStream[T any](ctx context.Context, maxSize int) *Stream[T] {
	s := &Stream[T]{
		ch:     make(chan T, maxSize),
		err:    make(chan error, 1), // 保持错误通道容量为1
		closed: make(chan struct{}),
		ctx:    ctx,
	}

	// // 监听上下文取消，自动关闭流
	// go func() {
	// 	<-ctx.Done()
	// 	s.CloseSend()
	// }()

	return s
}

func (s *Stream[T]) Send(ctx context.Context, item T) error {
	select {
	case <-s.ctx.Done():
		return ErrContextAlreadyDone
	case <-s.closed:
		return ErrChannelClosed
	case <-ctx.Done():
		// 检查具体原因
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return ErrSendContextTimeout
		case context.Canceled:
			return ErrSendContextCanceled
		default:
			return ErrSendContextAlreadyDone
		}
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
		return item, !ok, nil
	case err := <-s.err:
		return zero, true, err
	}
}

func (s *Stream[T]) CloseSend() {
	s.once.Do(func() {
		// 关闭 closed 通道，标记流已关闭
		close(s.closed)

		// 关闭数据通道，防止新的数据发送
		close(s.ch)

		// 最后关闭错误通道
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
