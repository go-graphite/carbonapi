package main

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

func NewLimiter(l int) limiter {
	return make(chan struct{}, l)
}
