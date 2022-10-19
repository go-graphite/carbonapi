package http

import (
	"fmt"

	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	"github.com/msaf1980/go-metrics"
	"go.uber.org/zap"
)

var ApiMetrics = struct {
	RequestCacheHits        metrics.Counter
	RequestCacheMisses      metrics.Counter
	BackendCacheHits        metrics.Counter
	BackendCacheMisses      metrics.Counter
	RequestsCacheOverheadNS metrics.Counter
	RequestsH               metrics.Histogram
	Requests200             metrics.Counter
	Requests400             metrics.Counter
	Requests403             metrics.Counter
	Requestsxxx             metrics.Counter // failback other 4xx statuses
	Requests500             metrics.Counter
	Requests503             metrics.Counter
	Requests5xx             metrics.Counter // failback other 5xx statuses

	RenderRequests metrics.Counter

	FindRequests metrics.Counter

	MemcacheTimeouts metrics.UGauge

	CacheSize  metrics.UGauge
	CacheItems metrics.Gauge
}{
	RenderRequests:          metrics.NewCounter(),
	RequestCacheHits:        metrics.NewCounter(),
	RequestCacheMisses:      metrics.NewCounter(),
	BackendCacheHits:        metrics.NewCounter(),
	BackendCacheMisses:      metrics.NewCounter(),
	RequestsCacheOverheadNS: metrics.NewCounter(),

	Requests200: metrics.NewCounter(),
	Requests400: metrics.NewCounter(),
	Requests403: metrics.NewCounter(),
	Requestsxxx: metrics.NewCounter(),
	Requests500: metrics.NewCounter(),
	Requests503: metrics.NewCounter(),
	Requests5xx: metrics.NewCounter(),

	FindRequests: metrics.NewCounter(),
}

var ZipperMetrics = struct {
	FindRequests metrics.Counter
	FindTimeouts metrics.Counter
	FindErrors   metrics.Counter

	SearchRequests metrics.Counter

	RenderRequests metrics.Counter
	RenderTimeouts metrics.Counter
	RenderErrors   metrics.Counter

	InfoRequests metrics.Counter
	InfoTimeouts metrics.Counter
	InfoErrors   metrics.Counter

	Timeouts metrics.Counter

	CacheMisses metrics.Counter
	CacheHits   metrics.Counter
}{
	FindRequests: metrics.NewCounter(),
	FindTimeouts: metrics.NewCounter(),
	FindErrors:   metrics.NewCounter(),

	SearchRequests: metrics.NewCounter(),

	RenderRequests: metrics.NewCounter(),
	RenderTimeouts: metrics.NewCounter(),
	RenderErrors:   metrics.NewCounter(),

	InfoRequests: metrics.NewCounter(),
	InfoTimeouts: metrics.NewCounter(),
	InfoErrors:   metrics.NewCounter(),

	Timeouts: metrics.NewCounter(),

	CacheHits:   metrics.NewCounter(),
	CacheMisses: metrics.NewCounter(),
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

func SetupMetrics(logger *zap.Logger) {
	switch config.Config.ResponseCacheConfig.Type {
	case "memcache":
		mcache := config.Config.ResponseCache.(*cache.MemcachedCache)

		ApiMetrics.MemcacheTimeouts = metrics.NewFunctionalUGauge(mcache.Timeouts)
	case "mem":
		qcache := config.Config.ResponseCache.(*cache.ExpireCache)

		ApiMetrics.CacheSize = metrics.NewFunctionalUGauge(qcache.Size)
		ApiMetrics.CacheItems = metrics.NewFunctionalGauge(func() int64 {
			return int64(qcache.Items())
		})
	default:
	}

	ApiMetrics.RequestsH = initRequestsHistogram()
}

func initRequestsHistogram() metrics.Histogram {
	if config.Config.Upstreams.SumBuckets {
		if len(config.Config.Upstreams.BucketsWidth) > 0 {
			labels := make([]string, len(config.Config.Upstreams.BucketsWidth)+1)

			for i := 0; i <= len(config.Config.Upstreams.BucketsWidth); i++ {
				if i >= len(config.Config.Upstreams.BucketsLabels) || config.Config.Upstreams.BucketsLabels[i] == "" {
					if i < len(config.Config.Upstreams.BucketsWidth) {
						labels[i] = fmt.Sprintf("_to_%dms", config.Config.Upstreams.BucketsWidth[i])
					} else {
						labels[i] = "_to_inf"
					}
				} else {
					labels[i] = config.Config.Upstreams.BucketsLabels[i]
				}
			}
			return metrics.NewVSumHistogram(config.Config.Upstreams.BucketsWidth, labels).
				SetNameTotal("")
		} else {
			labels := make([]string, config.Config.Upstreams.Buckets+1)

			for i := 0; i <= config.Config.Upstreams.Buckets; i++ {
				labels[i] = fmt.Sprintf("_to_%dms", (i+1)*100)
			}
			return metrics.NewFixedSumHistogram(100, int64(config.Config.Upstreams.Buckets)*100, 100).
				SetLabels(labels).
				SetNameTotal("")
		}
	} else if len(config.Config.Upstreams.BucketsWidth) > 0 {
		labels := make([]string, len(config.Config.Upstreams.BucketsWidth)+1)

		for i := 0; i <= len(config.Config.Upstreams.BucketsWidth); i++ {
			if i >= len(config.Config.Upstreams.BucketsLabels) || config.Config.Upstreams.BucketsLabels[i] == "" {
				if i == 0 {
					labels[i] = fmt.Sprintf("_in_0ms_to_%dms", config.Config.Upstreams.BucketsWidth[0])
				} else if i < len(config.Config.Upstreams.BucketsWidth) {
					labels[i] = fmt.Sprintf("_in_%dms_to_%dms", config.Config.Upstreams.BucketsWidth[i-1], config.Config.Upstreams.BucketsWidth[i])
				} else {
					labels[i] = fmt.Sprintf("_in_%dms_to_inf", config.Config.Upstreams.BucketsWidth[i-1])
				}
			} else {
				labels[i] = config.Config.Upstreams.BucketsLabels[i]
			}
		}
		return metrics.NewVSumHistogram(config.Config.Upstreams.BucketsWidth, labels).SetNameTotal("")
	} else {
		labels := make([]string, config.Config.Upstreams.Buckets+1)

		for i := 0; i <= config.Config.Upstreams.Buckets; i++ {
			labels[i] = fmt.Sprintf("_in_%dms_to_%dms", i*100, (i+1)*100)
		}
		return metrics.NewFixedSumHistogram(100, int64(config.Config.Upstreams.Buckets)*100, 100).
			SetLabels(labels).
			SetNameTotal("")
	}
}
