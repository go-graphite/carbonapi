package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	ecache "github.com/dgryski/go-expirecache"
)

var (
	ErrTimeout  = errors.New("cache: timeout")
	ErrNotFound = errors.New("cache: not found")
)

type BytesCache interface {
	Get(k string) ([]byte, error)
	Set(k string, v []byte, expire int32)
}

type NullCache struct{}

func (NullCache) Get(string) ([]byte, error) { return nil, ErrNotFound }
func (NullCache) Set(string, []byte, int32)  {}

type ExpireCache struct {
	ec *ecache.Cache
}

func (ec ExpireCache) Get(k string) ([]byte, error) {
	v, ok := ec.ec.Get(k)

	if !ok {
		return nil, ErrNotFound
	}

	return v.([]byte), nil
}

func (ec ExpireCache) Set(k string, v []byte, expire int32) {
	ec.ec.Set(k, v, uint64(len(v)), expire)
}

type MemcachedCache struct {
	client *memcache.Client
}

func (m *MemcachedCache) Get(k string) ([]byte, error) {
	key := sha1.Sum([]byte(k))
	hk := hex.EncodeToString(key[:])
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
		return nil, ErrTimeout
	case <-done:
	}

	if err != nil {
		// translate to internal cache miss error
		if err == memcache.ErrCacheMiss {
			err = ErrNotFound
		}
		return nil, err
	}

	return item.Value, nil
}

func (m *MemcachedCache) Set(k string, v []byte, expire int32) {
	key := sha1.Sum([]byte(k))
	hk := hex.EncodeToString(key[:])
	go m.client.Set(&memcache.Item{Key: hk, Value: v, Expiration: expire})
}
