package limiter

import (
	"context"
)

// ServerLimiter provides interface to limit amount of requests
type NoopLimiter struct {
}

func (l NoopLimiter) Capacity() int {
	return 0
}

// Enter claims one of free slots or blocks until there is one.
func (l NoopLimiter) Enter(ctx context.Context, s string) error {
	return nil
}

// Frees a slot in limiter
func (l NoopLimiter) Leave(ctx context.Context, s string) {
}
