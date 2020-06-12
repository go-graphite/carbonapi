package http

import (
	"expvar"
	"strconv"
	"sync/atomic"

	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	"go.uber.org/zap"
)

var ApiMetrics = struct {
	Requests              *expvar.Int
	RenderRequests        *expvar.Int
	RequestCacheHits      *expvar.Int
	RequestCacheMisses    *expvar.Int
	BackendCacheHits      *expvar.Int
	BackendCacheMisses    *expvar.Int
	RenderCacheOverheadNS *expvar.Int
	RequestBuckets        expvar.Func

	FindRequests *expvar.Int

	MemcacheTimeouts expvar.Func

	CacheSize  expvar.Func
	CacheItems expvar.Func
}{
	Requests: expvar.NewInt("requests"),
	// TODO: request_cache -> render_cache
	RenderRequests:        expvar.NewInt("render_requests"),
	RequestCacheHits:      expvar.NewInt("request_cache_hits"),
	RequestCacheMisses:    expvar.NewInt("request_cache_misses"),
	BackendCacheHits:      expvar.NewInt("backend_cache_hits"),
	BackendCacheMisses:    expvar.NewInt("backend_cache_misses"),
	RenderCacheOverheadNS: expvar.NewInt("render_cache_overhead_ns"),

	FindRequests: expvar.NewInt("find_requests"),
}

var ZipperMetrics = struct {
	FindRequests *expvar.Int
	FindTimeouts *expvar.Int
	FindErrors   *expvar.Int

	SearchRequests *expvar.Int

	RenderRequests *expvar.Int
	RenderTimeouts *expvar.Int
	RenderErrors   *expvar.Int

	InfoRequests *expvar.Int
	InfoTimeouts *expvar.Int
	InfoErrors   *expvar.Int

	Timeouts *expvar.Int

	CacheSize   expvar.Func
	CacheItems  expvar.Func
	CacheMisses *expvar.Int
	CacheHits   *expvar.Int
}{
	FindRequests: expvar.NewInt("zipper_find_requests"),
	FindTimeouts: expvar.NewInt("zipper_find_timeouts"),
	FindErrors:   expvar.NewInt("zipper_find_errors"),

	SearchRequests: expvar.NewInt("zipper_search_requests"),

	RenderRequests: expvar.NewInt("zipper_render_requests"),
	RenderTimeouts: expvar.NewInt("zipper_render_timeouts"),
	RenderErrors:   expvar.NewInt("zipper_render_errors"),

	InfoRequests: expvar.NewInt("zipper_info_requests"),
	InfoTimeouts: expvar.NewInt("zipper_info_timeouts"),
	InfoErrors:   expvar.NewInt("zipper_info_errors"),

	Timeouts: expvar.NewInt("zipper_timeouts"),

	CacheHits:   expvar.NewInt("zipper_cache_hits"),
	CacheMisses: expvar.NewInt("zipper_cache_misses"),
}

func ZipperStats(stats *zipperTypes.Stats) {
	if stats == nil {
		return
	}
	ZipperMetrics.Timeouts.Add(stats.Timeouts)
	ZipperMetrics.FindRequests.Add(stats.FindRequests)
	ZipperMetrics.FindTimeouts.Add(stats.FindTimeouts)
	ZipperMetrics.FindErrors.Add(stats.FindErrors)
	ZipperMetrics.RenderRequests.Add(stats.RenderRequests)
	ZipperMetrics.RenderTimeouts.Add(stats.RenderTimeouts)
	ZipperMetrics.RenderErrors.Add(stats.RenderErrors)
	ZipperMetrics.InfoRequests.Add(stats.InfoRequests)
	ZipperMetrics.InfoTimeouts.Add(stats.InfoTimeouts)
	ZipperMetrics.InfoErrors.Add(stats.InfoErrors)
	ZipperMetrics.SearchRequests.Add(stats.SearchRequests)
	ZipperMetrics.CacheMisses.Add(stats.CacheMisses)
	ZipperMetrics.CacheHits.Add(stats.CacheHits)
}

type BucketEntry int

var TimeBuckets []int64

func (b BucketEntry) String() string {
	return strconv.Itoa(int(atomic.LoadInt64(&TimeBuckets[b])))
}

func RenderTimeBuckets() interface{} {
	return TimeBuckets
}

func SetupMetrics(logger *zap.Logger) {
	switch config.Config.ResponseCacheConfig.Type {
	case "memcache":
		mcache := config.Config.ResponseCache.(*cache.MemcachedCache)

		ApiMetrics.MemcacheTimeouts = expvar.Func(func() interface{} {
			return mcache.Timeouts()
		})
		expvar.Publish("memcache_timeouts", ApiMetrics.MemcacheTimeouts)

	case "mem":
		qcache := config.Config.ResponseCache.(*cache.ExpireCache)

		ApiMetrics.CacheSize = expvar.Func(func() interface{} {
			return qcache.Size()
		})
		expvar.Publish("cache_size", ApiMetrics.CacheSize)

		ApiMetrics.CacheItems = expvar.Func(func() interface{} {
			return qcache.Items()
		})
		expvar.Publish("cache_items", ApiMetrics.CacheItems)
	default:
	}

	// +1 to track every over the number of buckets we track
	TimeBuckets = make([]int64, config.Config.Buckets+1)
	expvar.Publish("requestBuckets", expvar.Func(RenderTimeBuckets))
}
