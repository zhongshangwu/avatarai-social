package streams

import "context"

type Reciver[T any] interface {
	Recv(ctx context.Context) (T, error)
}
