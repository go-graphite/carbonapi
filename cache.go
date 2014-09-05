package main

import (
	"sync"
	"time"
)

type bytesCache interface {
	get(k string) ([]byte, bool)
	set(k string, v []byte)
}

type cacheElement struct {
	validUntil time.Time
	data       []byte
}

type expireCache struct {
	sync.Mutex
	cache map[string]cacheElement
}

func (ec *expireCache) get(k string) ([]byte, bool) {
	ec.Lock()
	v, ok := ec.cache[k]
	ec.Unlock()
	if !ok || v.validUntil.Before(time.Now()) {
		return nil, false
	}
	return v.data, ok
}

func (ec *expireCache) set(k string, v []byte) {
	ec.Lock()
	ec.cache[k] = cacheElement{validUntil: time.Now().Add(60 * time.Second), data: v}
	ec.Unlock()
}

func (ec *expireCache) cleaner() {

	var keys []string

	for {
		time.Sleep(5 * time.Minute)

		now := time.Now()
		ec.Lock()

		for k, v := range ec.cache {
			if v.validUntil.Before(now) {
				keys = append(keys, k)
			}
		}

		for _, k := range keys {
			delete(ec.cache, k)
		}

		keys = keys[:0]
		ec.Unlock()
	}
}
