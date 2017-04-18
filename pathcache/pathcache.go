package pathcache

import (
	"github.com/dgryski/go-expirecache"

	"time"
)

type pathKV struct {
	k string
	v []string
}

type PathCache struct {
	ec *expirecache.Cache

	expireDelaySec int32
}

func NewPathCache(ExpireDelaySec int32) PathCache {

	p := PathCache{
		ec:             expirecache.New(0),
		expireDelaySec: ExpireDelaySec,
	}

	go p.ec.ApproximateCleaner(10 * time.Second)

	return p
}

func (p *PathCache) ECItems() int {
	return p.ec.Items()
}

func (p *PathCache) ECSize() uint64 {
	return p.ec.Size()
}

func (p *PathCache) Set(k string, v []string) {

	var size uint64
	for _, vv := range v {
		size += uint64(len(vv))
	}

	p.ec.Set(k, v, size, p.expireDelaySec)
}

func (p *PathCache) Get(k string) ([]string, bool) {
	if v, ok := p.ec.Get(k); ok {
		return v.([]string), true
	}

	return nil, false
}
