package main

import (
	"bytes"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	cu "github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonzipper/zipper"
	pb3 "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/pathcache"
	"github.com/go-graphite/carbonzipper/mstats"
	"github.com/go-graphite/carbonzipper/util"
	"github.com/dgryski/httputil"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	pickle "github.com/kisielk/og-rek"
	"github.com/peterbourgon/g2g"

	"github.com/lomik/zapwriter"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

var DefaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

// Config contains configuration values
var Config = struct {
	Backends    []string
	MaxProcs    int
	IntervalSec int
	Port        int
	Buckets     int

	TimeoutMs                int
	TimeoutMsAfterAllStarted int

	SearchBackend string
	SearchPrefix  string

	GraphiteHost         string
	InternalMetricPrefix string

	MemcachedServers []string

	MaxIdleConnsPerHost int

	ConcurrencyLimitPerServer int
	ExpireDelaySec            int32
	Logger                    []zapwriter.Config

	zipper *zipper.Zipper
}{
	MaxProcs:    1,
	IntervalSec: 60,
	Port:        8080,
	Buckets:     10,

	TimeoutMs:                10000,
	TimeoutMsAfterAllStarted: 2000,

	MaxIdleConnsPerHost: 100,

	ExpireDelaySec: 10 * 60, // 10 minutes

	Logger: []zapwriter.Config{DefaultLoggerConfig},
}

// Metrics contains grouped expvars for /debug/vars and graphite
var Metrics = struct {
	FindRequests *expvar.Int
	FindErrors   *expvar.Int

	SearchRequests *expvar.Int

	RenderRequests *expvar.Int
	RenderErrors   *expvar.Int

	InfoRequests *expvar.Int
	InfoErrors   *expvar.Int

	Timeouts *expvar.Int

	MemcacheTimeouts expvar.Func

	CacheSize         expvar.Func
	CacheItems        expvar.Func
	CacheMisses       *expvar.Int
	CacheHits         *expvar.Int
	SearchCacheSize   expvar.Func
	SearchCacheItems  expvar.Func
	SearchCacheMisses *expvar.Int
	SearchCacheHits   *expvar.Int
}{
	FindRequests: expvar.NewInt("find_requests"),
	FindErrors:   expvar.NewInt("find_errors"),

	SearchRequests: expvar.NewInt("search_requests"),

	RenderRequests: expvar.NewInt("render_requests"),
	RenderErrors:   expvar.NewInt("render_errors"),

	InfoRequests: expvar.NewInt("info_requests"),
	InfoErrors:   expvar.NewInt("info_errors"),

	Timeouts: expvar.NewInt("timeouts"),

	CacheHits:         expvar.NewInt("cache_hits"),
	CacheMisses:       expvar.NewInt("cache_misses"),
	SearchCacheHits:   expvar.NewInt("search_cache_hits"),
	SearchCacheMisses: expvar.NewInt("search_cache_misses"),
}

// BuildVersion is defined at build and reported at startup and as expvar
var BuildVersion = "(development version)"

type serverResponse struct {
	server   string
	response []byte
}

// set during startup, read-only after that
var searchConfigured = false

const (
	contentTypeJSON     = "application/json"
	contentTypeProtobuf = "application/x-protobuf"
	contentTypePickle   = "application/pickle"
)

func findHandler(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	ctx := req.Context()
	ctx = util.SetUUID(ctx, uuid.String())
	logger := zapwriter.Logger("find").With(
		zap.String("handler", "find"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)
	logger.Debug("got find request",
		zap.String("request", req.URL.RequestURI()),
	)

	originalQuery := req.FormValue("query")
	format := req.FormValue("format")

	Metrics.FindRequests.Add(1)

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "find"),
		zap.String("format", format),
		zap.String("target", originalQuery),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)


	metrics, stats, err := Config.zipper.Find(ctx, logger, originalQuery)
	Config.zipper.SendStats(stats)
	if err != nil {
		accessLogger.Error("find failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		http.Error(w, "error fetching the data", http.StatusInternalServerError)
		return
	}

	encodeFindResponse(format, originalQuery, w, metrics)
	accessLogger.Info("request served",
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func encodeFindResponse(format, query string, w http.ResponseWriter, metrics []*pb3.GlobMatch) {
	switch format {
	case "protobuf3":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		var result pb3.GlobResponse
		result.Name = query
		result.Matches = metrics
		b, _ := result.Marshal()
		w.Write(b)
	case "protobuf":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		var result pb3.GlobResponse
		result.Name = query
		result.Matches = metrics
		b, _ := result.Marshal()
		w.Write(b)
	case "json":
		w.Header().Set("Content-Type", contentTypeJSON)
		jEnc := json.NewEncoder(w)
		jEnc.Encode(metrics)
	case "", "pickle":
		w.Header().Set("Content-Type", contentTypePickle)

		var result []map[string]interface{}

		for _, metric := range metrics {
			mm := map[string]interface{}{
				"metric_path": metric.Path,
				"isLeaf":      metric.IsLeaf,
			}
			result = append(result, mm)
		}

		pEnc := pickle.NewEncoder(w)
		pEnc.Encode(result)
	}
}

func renderHandler(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	memoryUsage := 0
	uuid := uuid.NewV4()
	ctx := req.Context()

	ctx = util.SetUUID(ctx, uuid.String())
	logger := zapwriter.Logger("render").With(
		zap.Int("memory_usage_bytes", memoryUsage),
		zap.String("handler", "render"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)

	logger.Debug("got render request",
		zap.String("request", req.URL.RequestURI()),
	)

	Metrics.RenderRequests.Add(1)

	req.ParseForm()
	target := req.FormValue("target")
	format := req.FormValue("format")

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "render"),
		zap.String("format", format),
		zap.String("target", target),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)

	from, err := strconv.Atoi(req.FormValue("from"))
	if err != nil {
		http.Error(w, "empty target", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "invalid from"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}
	until, err := strconv.Atoi(req.FormValue("until"))
	if err != nil {
		http.Error(w, "empty target", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "invalid from"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	if target == "" {
		http.Error(w, "empty target", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "empty target"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	metrics, stats, err := Config.zipper.Render(ctx, logger, target, int32(from), int32(until))
	Config.zipper.SendStats(stats)
	if err != nil {
		http.Error(w, "error fetching the data", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", err.Error()),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	switch format {
	case "protobuf3":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		b, err := metrics.Marshal()
		if err != nil {
			logger.Error("error marshaling data",
				zap.Int("memory_usage_bytes", memoryUsage),
				zap.Error(err),
			)
		}
		memoryUsage += len(b)
		w.Write(b)

	case "protobuf":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		b, err := metrics.Marshal()
		if err != nil {
			logger.Error("error marshaling data",
				zap.Int("memory_usage_bytes", memoryUsage),
				zap.Error(err),
			)
		}
		memoryUsage += len(b)
		w.Write(b)
	case "json":
		presponse := createRenderResponse(metrics, nil)
		w.Header().Set("Content-Type", contentTypeJSON)
		e := json.NewEncoder(w)
		e.Encode(presponse)

	case "", "pickle":
		presponse := createRenderResponse(metrics, pickle.None{})
		w.Header().Set("Content-Type", contentTypePickle)
		e := pickle.NewEncoder(w)
		e.Encode(presponse)
	}
	accessLogger.Info("request served",
		zap.Int("memory_usage_bytes", memoryUsage),
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func createRenderResponse(metrics *pb3.MultiFetchResponse, missing interface{}) []map[string]interface{} {

	var response []map[string]interface{}

	for _, metric := range metrics.GetMetrics() {

		var pvalues []interface{}
		for i, v := range metric.Values {
			if metric.IsAbsent[i] {
				pvalues = append(pvalues, missing)
			} else {
				pvalues = append(pvalues, v)
			}
		}

		// create the response
		presponse := map[string]interface{}{
			"start":  metric.StartTime,
			"step":   metric.StepTime,
			"end":    metric.StopTime,
			"name":   metric.Name,
			"values": pvalues,
		}
		response = append(response, presponse)
	}

	return response
}

func infoHandler(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	ctx := req.Context()
	ctx = util.SetUUID(ctx, uuid.String())
	logger := zapwriter.Logger("info").With(
		zap.String("handler", "info"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)

	logger.Debug("request",
		zap.String("request", req.URL.RequestURI()),
	)

	Metrics.InfoRequests.Add(1)

	req.ParseForm()
	target := req.FormValue("target")
	format := req.FormValue("format")

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "info"),
		zap.String("target", target),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", cu.GetUUID(ctx)),
	)

	if target == "" {
		accessLogger.Error("info failed",
			zap.Int("http_code", http.StatusBadRequest),
			zap.String("reason", "empty target"),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		http.Error(w, "info: empty target", http.StatusBadRequest)
		return
	}

	infos, stats, err := Config.zipper.Info(ctx, logger, target)
	Config.zipper.SendStats(stats)
	if err != nil {
		accessLogger.Error("info failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		http.Error(w, "info: error processing request", http.StatusInternalServerError)
		return
	}

	switch format {
	case "protobuf3":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		var result pb3.ZipperInfoResponse
		result.Responses = make([]*pb3.ServerInfoResponse, len(infos))
		for s, i := range infos {
			var r pb3.ServerInfoResponse
			r.Server = s
			r.Info = &i
			result.Responses = append(result.Responses, &r)
		}
		b, _ := result.Marshal()
		w.Write(b)
	case "protobuf":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		var result pb3.ZipperInfoResponse
		result.Responses = make([]*pb3.ServerInfoResponse, len(infos))
		for s, i := range infos {
			var r pb3.ServerInfoResponse
			r.Server = s
			r.Info = &i
			result.Responses = append(result.Responses, &r)
		}
		b, _ := result.Marshal()
		w.Write(b)
	case "", "json":
		w.Header().Set("Content-Type", contentTypeJSON)
		jEnc := json.NewEncoder(w)
		jEnc.Encode(infos)
	}
	accessLogger.Info("request served",
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func lbCheckHandler(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	logger := zapwriter.Logger("loadbalancer").With(zap.String("handler", "loadbalancer"))
	accessLogger := zapwriter.Logger("access").With(zap.String("handler", "loadbalancer"))
	logger.Debug("loadbalacner",
		zap.String("request", req.URL.RequestURI()),
	)

	fmt.Fprintf(w, "Ok\n")
	accessLogger.Info("lb request served",
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func stripCommentHeader(cfg []byte) []byte {

	// strip out the comment header block that begins with '#' characters
	// as soon as we see a line that starts with something _other_ than '#', we're done

	var idx int
	for cfg[0] == '#' {
		idx = bytes.Index(cfg, []byte("\n"))
		if idx == -1 || idx+1 == len(cfg) {
			return nil
		}
		cfg = cfg[idx+1:]
	}

	return cfg
}

func main() {
	err := zapwriter.ApplyConfig([]zapwriter.Config{DefaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")

	}
	logger := zapwriter.Logger("main")

	configFile := flag.String("c", "", "config file (json)")
	port := flag.Int("p", 0, "port to listen on")
	maxprocs := flag.Int("maxprocs", 0, "GOMAXPROCS")
	interval := flag.Duration("i", 0, "interval to report internal statistics to graphite")
	pidFile := flag.String("pid", "", "pidfile (default: empty, don't create pidfile)")

	flag.Parse()

	expvar.NewString("BuildVersion").Set(BuildVersion)

	if *configFile == "" {
		logger.Fatal("missing config file option")
	}

	cfgjs, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatal("unable to load config file:",
			zap.Error(err),
		)
	}

	cfgjs = stripCommentHeader(cfgjs)

	if cfgjs == nil {
		logger.Fatal("error removing header comment from ",
			zap.String("config_file", *configFile),
		)
	}

	err = json.Unmarshal(cfgjs, &Config)
	if err != nil {
		logger.Fatal("error parsing config file: ",
			zap.Error(err),
		)
	}

	if len(Config.Backends) == 0 {
		logger.Fatal("no Backends loaded -- exiting")
	}

	err = zapwriter.ApplyConfig(Config.Logger)
	if err != nil {
		logger.Fatal("Failed to apply config",
			zap.Any("config", Config.Logger),
			zap.Error(err),
		)
	}

	// command line overrides config file

	if *port != 0 {
		Config.Port = *port
	}

	if *maxprocs != 0 {
		Config.MaxProcs = *maxprocs
	}

	if *interval == 0 {
		*interval = time.Duration(Config.IntervalSec) * time.Second
	}

	searchConfigured = len(Config.SearchPrefix) > 0 && len(Config.SearchBackend) > 0

	portStr := fmt.Sprintf(":%d", Config.Port)

	logger = zapwriter.Logger("main")
	logger.Info("starting carbonzipper",
		zap.String("build_version", BuildVersion),
		zap.Int("GOMAXPROCS", Config.MaxProcs),
		zap.Duration("stats interval", *interval),
		zap.Int("concurency_limit_per_server", Config.ConcurrencyLimitPerServer),
		zap.String("graphite_host", Config.GraphiteHost),
		zap.String("listen_port", portStr),
	)

	runtime.GOMAXPROCS(Config.MaxProcs)

	// +1 to track every over the number of buckets we track
	timeBuckets = make([]int64, Config.Buckets+1)

	httputil.PublishTrackedConnections("httptrack")
	expvar.Publish("requestBuckets", expvar.Func(renderTimeBuckets))

	// export config via expvars
	expvar.Publish("Config", expvar.Func(func() interface{} { return Config }))

	/* Configure zipper */
	// set up caches
	zipperConfig := &zipper.Config{
		PathCache: pathcache.NewPathCache(Config.MemcachedServers, Config.ExpireDelaySec),
		SearchCache: pathcache.NewPathCache(Config.MemcachedServers, Config.ExpireDelaySec),

		ConcurrencyLimitPerServer: Config.ConcurrencyLimitPerServer,
		MaxIdleConnsPerHost: Config.MaxIdleConnsPerHost,
		Backends:    Config.Backends,

		SearchBackend: Config.SearchBackend,
		SearchPrefix:  Config.SearchPrefix,
		TimeoutAfterAllStarted: time.Duration(Config.TimeoutMsAfterAllStarted) * time.Millisecond,
		Timeout: time.Duration(Config.TimeoutMs) * time.Millisecond,
	}

	Metrics.CacheSize = expvar.Func(func() interface{} { return zipperConfig.PathCache.ECSize() })
	expvar.Publish("cacheSize", Metrics.CacheSize)

	Metrics.CacheItems = expvar.Func(func() interface{} { return zipperConfig.PathCache.ECItems() })
	expvar.Publish("cacheItems", Metrics.CacheItems)

	Metrics.SearchCacheSize = expvar.Func(func() interface{} { return zipperConfig.SearchCache.ECSize() })
	expvar.Publish("searchCacheSize", Metrics.SearchCacheSize)

	Metrics.SearchCacheItems = expvar.Func(func() interface{} { return zipperConfig.SearchCache.ECItems() })
	expvar.Publish("searchCacheItems", Metrics.SearchCacheItems)

	Config.zipper = zipper.NewZipper(sendStats, zipperConfig)

	if len(Config.MemcachedServers) > 0 {
		logger.Info("memcached configured",
			zap.Strings("servers", Config.MemcachedServers),
		)

		Metrics.MemcacheTimeouts = expvar.Func(func() interface{} {
			return zipperConfig.PathCache.MCTimeouts() + zipperConfig.SearchCache.MCTimeouts()
		})
		expvar.Publish("memcacheTimeouts", Metrics.MemcacheTimeouts)
	}

	http.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(cu.ParseCtx(findHandler), bucketRequestTimes)))
	http.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(cu.ParseCtx(renderHandler), bucketRequestTimes)))
	http.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(cu.ParseCtx(infoHandler), bucketRequestTimes)))
	http.HandleFunc("/lb_check", lbCheckHandler)

	// nothing in the config? check the environment
	if Config.GraphiteHost == "" {
		if host := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); host != ":" {
			Config.GraphiteHost = host
		}
	}

	if Config.InternalMetricPrefix == "" {
		Config.InternalMetricPrefix = "carbon.zipper"
	}

	// only register g2g if we have a graphite host
	if Config.GraphiteHost != "" {
		// register our metrics with graphite
		graphite := g2g.NewGraphite(Config.GraphiteHost, *interval, 10*time.Second)

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		prefix := Config.InternalMetricPrefix

		graphite.Register(fmt.Sprintf("%s.%s.find_requests", prefix, hostname), Metrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.%s.find_errors", prefix, hostname), Metrics.FindErrors)

		graphite.Register(fmt.Sprintf("%s.%s.render_requests", prefix, hostname), Metrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.%s.render_errors", prefix, hostname), Metrics.RenderErrors)

		graphite.Register(fmt.Sprintf("%s.%s.info_requests", prefix, hostname), Metrics.InfoRequests)
		graphite.Register(fmt.Sprintf("%s.%s.info_errors", prefix, hostname), Metrics.InfoErrors)

		graphite.Register(fmt.Sprintf("%s.%s.timeouts", prefix, hostname), Metrics.Timeouts)

		for i := 0; i <= Config.Buckets; i++ {
			graphite.Register(fmt.Sprintf("%s.%s.requests_in_%dms_to_%dms", prefix, hostname, i*100, (i+1)*100), bucketEntry(i))
		}

		if Metrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("%s.%s.cache_size", prefix, hostname), Metrics.CacheSize)
			graphite.Register(fmt.Sprintf("%s.%s.cache_items", prefix, hostname), Metrics.CacheItems)

			graphite.Register(fmt.Sprintf("%s.%s.search_cache_size", prefix, hostname), Metrics.SearchCacheSize)
			graphite.Register(fmt.Sprintf("%s.%s.search_cache_items", prefix, hostname), Metrics.SearchCacheItems)
		}

		if Metrics.MemcacheTimeouts != nil {
			graphite.Register(fmt.Sprintf("%s.%s.memcache_timeouts", prefix, hostname), Metrics.MemcacheTimeouts)
		}

		graphite.Register(fmt.Sprintf("%s.%s.cache_hits", prefix, hostname), Metrics.CacheHits)
		graphite.Register(fmt.Sprintf("%s.%s.cache_misses", prefix, hostname), Metrics.CacheMisses)

		graphite.Register(fmt.Sprintf("%s.%s.search_cache_hits", prefix, hostname), Metrics.SearchCacheHits)
		graphite.Register(fmt.Sprintf("%s.%s.search_cache_misses", prefix, hostname), Metrics.SearchCacheMisses)

		go mstats.Start(*interval)

		graphite.Register(fmt.Sprintf("%s.%s.alloc", prefix, hostname), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.%s.total_alloc", prefix, hostname), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.%s.num_gc", prefix, hostname), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.%s.pause_ns", prefix, hostname), &mstats.PauseNS)
	}

	if *pidFile != "" {
		pidfile.SetPidfilePath(*pidFile)
		err = pidfile.Write()
		if err != nil {
			log.Fatalln("error during pidfile.Write():", err)
		}
	}

	err = gracehttp.Serve(&http.Server{
		Addr:    portStr,
		Handler: nil,
	})

	if err != nil {
		log.Fatal("error during gracehttp.Serve()",
			zap.Error(err),
		)
	}
}

var timeBuckets []int64

type bucketEntry int

func (b bucketEntry) String() string {
	return strconv.Itoa(int(atomic.LoadInt64(&timeBuckets[b])))
}

func renderTimeBuckets() interface{} {
	return timeBuckets
}

func bucketRequestTimes(req *http.Request, t time.Duration) {
	logger := zapwriter.Logger("slow")

	ms := t.Nanoseconds() / int64(time.Millisecond)

	bucket := int(ms / 100)

	if bucket < Config.Buckets {
		atomic.AddInt64(&timeBuckets[bucket], 1)
	} else {
		// Too big? Increment overflow bucket and log
		atomic.AddInt64(&timeBuckets[Config.Buckets], 1)
		logger.Warn("Slow Request",
			zap.Duration("time", t),
			zap.String("url", req.URL.String()),
		)
	}
}

func sendStats(stats *zipper.Stats) {
	Metrics.Timeouts.Add(stats.Timeouts)
	Metrics.FindErrors.Add(stats.FindErrors)
	Metrics.RenderErrors.Add(stats.RenderErrors)
	Metrics.InfoErrors.Add(stats.InfoErrors)
	Metrics.SearchRequests.Add(stats.SearchRequests)
	Metrics.SearchCacheHits.Add(stats.SearchCacheHits)
	Metrics.SearchCacheMisses.Add(stats.SearchCacheMisses)
	Metrics.CacheMisses.Add(stats.CacheMisses)
	Metrics.CacheHits.Add(stats.CacheHits)
}

