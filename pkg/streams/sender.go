package streams

import (
	"context"
)

type Sender[T any] interface {
	Send(ctx context.Context, msg T) error
}
