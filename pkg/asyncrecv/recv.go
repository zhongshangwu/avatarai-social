package aysncrecv

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type Reciver[T any] interface {
	Recv() (T, error)
}

func AsyncRecv[T any](ctx context.Context, r Reciver[T], bufferSize int, logPrefix string) *streams.Stream[T] {
	s := streams.NewStream[T](ctx, bufferSize)
	go func() {
		defer s.CloseSend()

		for {
			item, err := r.Recv()
			if err != nil {
				logrus.WithError(err).Error(logPrefix + "[AsyncRecv] recv error")
				s.SendError(err)
				return
			}
			err = s.Send(item)
			if err != nil {
				logrus.WithError(err).Error(logPrefix + "[AsyncRecv] send error")
				return
			}
		}
	}()
	return s
}
