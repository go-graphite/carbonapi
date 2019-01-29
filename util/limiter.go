package util

type SimpleLimiter chan struct{}

func (l SimpleLimiter) Enter() { l <- struct{}{} }
func (l SimpleLimiter) Leave() { <-l }

func NewSimpleLimiter(l int) SimpleLimiter {
	return make(chan struct{}, l)
}
