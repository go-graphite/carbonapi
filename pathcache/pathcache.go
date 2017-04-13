package pathcache

import (
	"github.com/go-graphite/carbonzipper/cache"
	"github.com/go-graphite/carbonzipper/internal/ipb3"
	"github.com/dgryski/go-expirecache"

	"time"
)

type pathKV struct {
	k string
	v []string
}

type PathCache struct {
	ec *expirecache.Cache
	mc *cache.MemcachedCache
	q  chan pathKV

	expireDelaySec int32
}

func NewPathCache(servers []string, ExpireDelaySec int32) PathCache {

	p := PathCache{
		ec: expirecache.New(0),
		expireDelaySec: ExpireDelaySec,
	}

	go p.ec.ApproximateCleaner(10 * time.Second)

	if len(servers) > 0 {
		p.q = make(chan pathKV, 1000)
		p.mc = cache.NewMemcached("czip", servers...).(*cache.MemcachedCache)

		go func() {
			for kv := range p.q {
				var b, _ = (&ipb3.PathCacheEntry{Hosts: kv.v}).Marshal()
				p.mc.SyncSet(kv.k, b, p.expireDelaySec)
			}
		}()
	}

	return p
}

func (p *PathCache) ECItems() int {
	return p.ec.Items()
}

func (p *PathCache) ECSize() uint64 {
	return p.ec.Size()
}

func (p *PathCache) MCTimeouts() uint64 {
	return p.mc.Timeouts()
}


func (p *PathCache) Set(k string, v []string) {

	var size uint64
	for _, vv := range v {
		size += uint64(len(vv))
	}

	p.ec.Set(k, v, size, p.expireDelaySec)

	select {
	case p.q <- pathKV{k: k, v: v}:
	default:
	}
}

func (p *PathCache) Get(k string) ([]string, bool) {
	if v, ok := p.ec.Get(k); ok {
		return v.([]string), true
	}

	if p.mc == nil {
		return nil, false
	}

	// check second-level bytes cache
	b, err := p.mc.Get(k)
	if err != nil {
		return nil, false
	}

	var v ipb3.PathCacheEntry

	err = v.Unmarshal(b)

	if err != nil {
		return nil, false
	}

	return v.Hosts, true
}
