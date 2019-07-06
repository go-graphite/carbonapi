package limiter

import (
	"context"
)

type ServerLimiter interface{
	Capacity() int
	Enter(ctx context.Context, s string) error
	Leave(ctx context.Context, s string)
}