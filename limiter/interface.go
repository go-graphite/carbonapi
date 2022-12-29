package limiter

import (
	"context"
	"errors"
)

var ErrTimeout = errors.New("timeout exceeded")

type ServerLimiter interface {
	Capacity() int
	Enter(ctx context.Context, s string) error
	Leave(ctx context.Context, s string)
}
