package pathcache

import (
	"github.com/dgryski/go-expirecache"
	"github.com/go-graphite/carbonapi/zipper/types"

	"time"
)

// PathCache provides general interface to cache find and search queries
type PathCache struct {
	ec *expirecache.Cache

	expireDelaySec int32
}

// NewPathCache initializes PathCache structure
func NewPathCache(ExpireDelaySec int32) PathCache {

	p := PathCache{
		ec:             expirecache.New(0),
		expireDelaySec: ExpireDelaySec,
	}

	go p.ec.ApproximateCleaner(10 * time.Second)

	return p
}

// ECItems returns amount of items in the cache
func (p *PathCache) ECItems() int {
	return p.ec.Items()
}

// ECSize returns size of the cache
func (p *PathCache) ECSize() uint64 {
	return p.ec.Size()
}

// Set allows to set a key (k) to value (v).
func (p *PathCache) Set(k string, v []types.ServerClient) {

	var size uint64
	for _, vv := range v {
		size += uint64(len(vv.Backends()))
	}

	p.ec.Set(k, v, size, p.expireDelaySec)
}

// Get returns an an element by key. If not successful - returns also false in second var.
func (p *PathCache) Get(k string) ([]types.ServerClient, bool) {
	if v, ok := p.ec.Get(k); ok {
		return v.([]types.ServerClient), true
	}

	return nil, false
}
