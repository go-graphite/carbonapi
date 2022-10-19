package main

import (
	"os"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/http"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-metrics/graphite"
	"go.uber.org/zap"
)

var (
	g *graphite.Graphite
)

func setupGraphiteMetrics(logger *zap.Logger) {
	var host string
	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || config.Config.Graphite.Host != "" {
		switch {
		case envhost != ":" && config.Config.Graphite.Host != "":
			host = config.Config.Graphite.Host
		case envhost != ":":
			host = envhost
		case config.Config.Graphite.Host != "":
			host = config.Config.Graphite.Host
		}
	}

	logger.Info("starting carbonapi",
		zap.String("build_version", BuildVersion),
		zap.Any("config", config.Config),
	)

	if host != "" {
		hostname, _ := os.Hostname()
		hostname = strings.ReplaceAll(hostname, ".", "_")

		prefix := config.Config.Graphite.Prefix

		pattern := config.Config.Graphite.Pattern
		pattern = strings.ReplaceAll(pattern, "{prefix}", prefix)
		pattern = strings.ReplaceAll(pattern, "{fqdn}", hostname)

		// register our metrics with graphite
		g = graphite.New(config.Config.Graphite.Interval, pattern, host, 10*time.Second)

		// StatsD client
		if config.Config.Graphite.Statsd != "" && config.Config.Upstreams.ExtendedStat {
			var err error
			config := &statsd.ClientConfig{
				Address:       config.Config.Graphite.Statsd,
				Prefix:        pattern,
				ResInterval:   5 * time.Minute,
				UseBuffered:   true,
				FlushInterval: 300 * time.Millisecond,
			}
			http.Gstatsd, err = statsd.NewClientWithConfig(config)
			if err != nil {
				logger.Error("statsd init", zap.Error(err))
			}
		}

		if http.Gstatsd == nil {
			http.Gstatsd = http.NullSender{}
		}

		metrics.Register("request_cache_hits", http.ApiMetrics.RequestCacheHits)
		metrics.Register("request_cache_misses", http.ApiMetrics.RequestCacheMisses)
		metrics.Register("request_cache_overhead_ns", http.ApiMetrics.RequestsCacheOverheadNS)
		metrics.Register("backend_cache_hits", http.ApiMetrics.BackendCacheHits)
		metrics.Register("backend_cache_misses", http.ApiMetrics.BackendCacheMisses)

		if config.Config.Upstreams.ExtendedStat {
			metrics.Register("requests_status_code.200", http.ApiMetrics.Requests200)
			metrics.Register("requests_status_code.400", http.ApiMetrics.Requests400)
			metrics.Register("requests_status_code.403", http.ApiMetrics.Requests403)
			metrics.Register("requests_status_code.4xx", http.ApiMetrics.Requestsxxx)
			metrics.Register("requests_status_code.500", http.ApiMetrics.Requests500)
			metrics.Register("requests_status_code.503", http.ApiMetrics.Requests503)
			metrics.Register("requests_status_code.5xx", http.ApiMetrics.Requests5xx)
		}

		// requests histogram
		metrics.Register("requests", http.ApiMetrics.RequestsH)

		metrics.Register("find_requests", http.ApiMetrics.FindRequests)
		metrics.Register("render_requests", http.ApiMetrics.RenderRequests)

		if http.ApiMetrics.MemcacheTimeouts != nil {
			metrics.Register("memcache_timeouts", http.ApiMetrics.MemcacheTimeouts)
		}

		if http.ApiMetrics.CacheSize != nil {
			metrics.Register("cache_size", http.ApiMetrics.CacheSize)
			metrics.Register("cache_items", http.ApiMetrics.CacheItems)
		}

		metrics.Register("zipper.find_requests", http.ZipperMetrics.FindRequests)
		metrics.Register("zipper.find_errors", http.ZipperMetrics.FindErrors)

		metrics.Register("zipper.render_requests", http.ZipperMetrics.RenderRequests)
		metrics.Register("zipper.render_errors", http.ZipperMetrics.RenderErrors)

		metrics.Register("zipper.info_requests", http.ZipperMetrics.InfoRequests)
		metrics.Register("zipper.info_errors", http.ZipperMetrics.InfoErrors)

		metrics.Register("zipper.timeouts", http.ZipperMetrics.Timeouts)

		metrics.Register("zipper.cache_hits", http.ZipperMetrics.CacheHits)
		metrics.Register("zipper.cache_misses", http.ZipperMetrics.CacheMisses)

		metrics.RegisterRuntimeMemStats(nil)
		go metrics.CaptureRuntimeMemStats(config.Config.Graphite.Interval)

		g.Start(nil)
	}
}
