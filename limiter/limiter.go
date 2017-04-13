package limiter

type ServerLimiter map[string]chan struct{}

func NewServerLimiter(servers []string, l int) ServerLimiter {
	sl := make(map[string]chan struct{})

	for _, s := range servers {
		sl[s] = make(chan struct{}, l)
	}

	return sl
}

func (sl ServerLimiter) Enter(s string) {
	if sl == nil {
		return
	}
	sl[s] <- struct{}{}
}

func (sl ServerLimiter) Leave(s string) {
	if sl == nil {
		return
	}
	<-sl[s]
}
