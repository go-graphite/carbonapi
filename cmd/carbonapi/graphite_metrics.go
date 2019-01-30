package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/http"
	"github.com/go-graphite/carbonapi/mstats"
	"github.com/peterbourgon/g2g"
	"go.uber.org/zap"
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
		// register our metrics with graphite
		graphite := g2g.NewGraphite(host, config.Config.Graphite.Interval, 10*time.Second)

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		prefix := config.Config.Graphite.Prefix

		pattern := config.Config.Graphite.Pattern
		pattern = strings.Replace(pattern, "{prefix}", prefix, -1)
		pattern = strings.Replace(pattern, "{fqdn}", hostname, -1)

		graphite.Register(fmt.Sprintf("%s.requests", pattern), http.ApiMetrics.Requests)
		graphite.Register(fmt.Sprintf("%s.request_cache_hits", pattern), http.ApiMetrics.RequestCacheHits)
		graphite.Register(fmt.Sprintf("%s.request_cache_misses", pattern), http.ApiMetrics.RequestCacheMisses)
		graphite.Register(fmt.Sprintf("%s.request_cache_overhead_ns", pattern), http.ApiMetrics.RenderCacheOverheadNS)

		for i := 0; i <= config.Config.Buckets; i++ {
			graphite.Register(fmt.Sprintf("%s.requests_in_%dms_to_%dms", pattern, i*100, (i+1)*100), http.BucketEntry(i))
		}

		graphite.Register(fmt.Sprintf("%s.find_requests", pattern), http.ApiMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.find_cache_hits", pattern), http.ApiMetrics.FindCacheHits)
		graphite.Register(fmt.Sprintf("%s.find_cache_misses", pattern), http.ApiMetrics.FindCacheMisses)
		graphite.Register(fmt.Sprintf("%s.find_cache_overhead_ns", pattern), http.ApiMetrics.FindCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.render_requests", pattern), http.ApiMetrics.RenderRequests)

		if http.ApiMetrics.MemcacheTimeouts != nil {
			graphite.Register(fmt.Sprintf("%s.memcache_timeouts", pattern), http.ApiMetrics.MemcacheTimeouts)
		}

		if http.ApiMetrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("%s.cache_size", pattern), http.ApiMetrics.CacheSize)
			graphite.Register(fmt.Sprintf("%s.cache_items", pattern), http.ApiMetrics.CacheItems)
		}

		graphite.Register(fmt.Sprintf("%s.zipper.find_requests", pattern), http.ZipperMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.find_errors", pattern), http.ZipperMetrics.FindErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.render_requests", pattern), http.ZipperMetrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.render_errors", pattern), http.ZipperMetrics.RenderErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.info_requests", pattern), http.ZipperMetrics.InfoRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.info_errors", pattern), http.ZipperMetrics.InfoErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.timeouts", pattern), http.ZipperMetrics.Timeouts)

		graphite.Register(fmt.Sprintf("%s.zipper.cache_hits", pattern), http.ZipperMetrics.CacheHits)
		graphite.Register(fmt.Sprintf("%s.zipper.cache_misses", pattern), http.ZipperMetrics.CacheMisses)

		go mstats.Start(config.Config.Graphite.Interval)

		graphite.Register(fmt.Sprintf("%s.alloc", pattern), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.total_alloc", pattern), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.num_gc", pattern), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.pause_ns", pattern), &mstats.PauseNS)

	}
}
