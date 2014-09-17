package main

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type bytesCache interface {
	get(k string) ([]byte, bool)
	set(k string, v []byte, expire int32)
}

type nullCache struct{}

func (_ nullCache) get(string) ([]byte, bool) { return nil, false }
func (_ nullCache) set(string, []byte, int32) {}

type cacheElement struct {
	validUntil time.Time
	data       []byte
}

type expireCache struct {
	sync.Mutex
	cache     map[string]cacheElement
	keys      []string
	totalSize uint64
	maxSize   uint64
}

func (ec *expireCache) get(k string) ([]byte, bool) {
	ec.Lock()
	v, ok := ec.cache[k]
	ec.Unlock()
	if !ok || v.validUntil.Before(timeNow()) {
		// Can't actually delete this element from the cache here since
		// we can't remove the key from ec.keys without a linear search.
		// It'll get removed during the next cleanup
		return nil, false
	}
	return v.data, ok
}

func (ec *expireCache) set(k string, v []byte, expire int32) {
	ec.Lock()
	oldv, ok := ec.cache[k]
	if !ok {
		ec.keys = append(ec.keys, k)
	} else {
		ec.totalSize -= uint64(len(oldv.data))
	}

	ec.totalSize += uint64(len(v))
	ec.cache[k] = cacheElement{validUntil: timeNow().Add(time.Duration(expire) * time.Second), data: v}

	for ec.maxSize > 0 && ec.totalSize > ec.maxSize {
		ec.randomEvict()
	}

	ec.Unlock()
}

func (ec *expireCache) randomEvict() {
	slot := rand.Intn(len(ec.keys))
	k := ec.keys[slot]

	ec.keys[slot] = ec.keys[len(ec.keys)-1]
	ec.keys = ec.keys[:len(ec.keys)-1]

	v := ec.cache[k]
	ec.totalSize -= uint64(len(v.data))

	delete(ec.cache, k)
}

func (ec *expireCache) cleaner() {

	var keys []string

	for {
		cleanerSleep(5 * time.Minute)

		now := timeNow()
		ec.Lock()

		// We could potentially be holding this lock for a long time,
		// but since we keep the cache expiration times small, we
		// expect only a small number of elements here to loop over

		for _, k := range ec.keys {
			v := ec.cache[k]
			if v.validUntil.Before(now) {
				ec.totalSize -= uint64(len(v.data))
				keys = append(keys, k)
			}
		}

		for _, k := range keys {
			delete(ec.cache, k)
		}

		keys = keys[:0]
		ec.Unlock()
		cleanerDone()
	}
}

var (
	cleanerSleep = time.Sleep
	cleanerDone  = func() {}
)

type memcachedCache struct {
	client *memcache.Client
}

func (m *memcachedCache) get(k string) ([]byte, bool) {
	hk := fmt.Sprintf("%x", md5.Sum([]byte(k)))
	done := make(chan bool, 1)

	var err error
	var item *memcache.Item

	go func() {
		item, err = m.client.Get(hk)
		done <- true
	}()

	timeout := time.After(50 * time.Millisecond)

	select {
	case <-timeout:
		Metrics.MemcacheTimeouts.Add(1)
		return nil, false
	case <-done:
	}

	if err != nil {
		return nil, false
	}

	return item.Value, true
}

func (m *memcachedCache) set(k string, v []byte, expire int32) {
	hk := fmt.Sprintf("%x", md5.Sum([]byte(k)))
	go m.client.Set(&memcache.Item{Key: hk, Value: v, Expiration: expire})
}
