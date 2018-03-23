package cache

import (
	"context"
	"sync/atomic"

	"github.com/dgryski/go-expirecache"
)

const (
	QueryIsPending uint64 = 1 << iota
	DataIsAvailable
)

type QueryItem struct {
	Data          atomic.Value
	Flags         uint64 // DataIsAvailable or QueryIsPending
	QueryFinished chan struct{}

	parent *QueryCache
}

func (q *QueryItem) GetStatus() uint64 {
	s := atomic.LoadUint64(&q.Flags)
	return s
}

func (q *QueryItem) FetchOrLock(ctx context.Context) (interface{}, bool) {
	d := q.Data.Load()
	if d != nil {
		return d, true
	}

	ok := atomic.CompareAndSwapUint64(&q.Flags, 0, QueryIsPending)
	if ok {
		// We are the leader now and will be fetching the data
		return nil, false
	}

	select {
	case <-ctx.Done():
		return nil, false
	case <-q.QueryFinished:
		break
	}

	return q.Data.Load(), true
}

func (q *QueryItem) StoreAbort() {
	d := q.Data.Load()
	if d != nil {
		return
	}
	oldChan := q.QueryFinished
	q.QueryFinished = make(chan struct{})
	close(oldChan)
	atomic.StoreUint64(&q.Flags, 0)
}

func (q *QueryItem) StoreAndUnlock(data interface{}, size uint64) {
	q.Data.Store(data)
	atomic.StoreUint64(&q.Flags, DataIsAvailable)
	close(q.QueryFinished)
	atomic.AddUint64(&q.parent.totalSize, size)
}

type QueryCache struct {
	ec *expirecache.Cache

	objectCount uint64
	totalSize   uint64
	expireTime  int32
}

func NewQueryCache(queryCacheSizeMB uint64, expireTime int32) *QueryCache {
	return &QueryCache{
		ec:         expirecache.New(queryCacheSizeMB),
		expireTime: expireTime,
	}
}

// TODO: Make size and expire configurable
func (q *QueryCache) GetQueryItem(k string) *QueryItem {
	objectCount := atomic.AddUint64(&q.objectCount, 1)
	size := atomic.AddUint64(&q.totalSize, 1)
	emptyQueryItem := &QueryItem{
		QueryFinished: make(chan struct{}),
		parent:        q,
	}
	return q.ec.GetOrSet(k, emptyQueryItem, size/objectCount, q.expireTime).(*QueryItem)
}
