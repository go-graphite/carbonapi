package limiter

import (
	"context"
	"errors"
)

// ServerLimiter provides interface to limit amount of requests
type ServerLimiter struct {
	m   map[string]chan struct{}
	cap int
}

// NewServerLimiter creates a limiter for specific servers list.
func NewServerLimiter(servers []string, l int) *ServerLimiter {
	if l == 0 {
		return &ServerLimiter{}
	}

	sl := make(map[string]chan struct{})

	for _, s := range servers {
		sl[s] = make(chan struct{}, l)
	}

	return &ServerLimiter{
		m:   sl,
		cap: l,
	}
}

func (sl ServerLimiter) Capacity() int {
	return sl.cap
}

// Enter claims one of free slots or blocks until there is one.
func (sl ServerLimiter) Enter(ctx context.Context, s string) error {
	if sl.m == nil {
		return nil
	}

	select {
	case sl.m[s] <- struct{}{}:
		return nil
	case <-ctx.Done():
		return errors.New("timeout exceeded")
	}
}

// Frees a slot in limiter
func (sl ServerLimiter) Leave(ctx context.Context, s string) {
	if sl.m == nil {
		return
	}

	<-sl.m[s]
}
