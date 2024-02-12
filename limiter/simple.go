package limiter

import "context"

type SimpleLimiter chan struct{}

func (l SimpleLimiter) Enter(ctx context.Context) error {
	if l == nil {
		return nil
	}

	select {
	case l <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrTimeout
	}
}

func (l SimpleLimiter) Leave() {
	if l != nil {
		<-l
	}
}

func NewSimpleLimiter(l int) SimpleLimiter {
	return make(chan struct{}, l)
}
