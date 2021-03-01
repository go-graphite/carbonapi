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

	"github.com/ansel1/merry"

	"github.com/dgryski/httputil"
	"github.com/facebookgo/pidfile"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"

	"github.com/go-graphite/carbonapi/intervalset"
	"github.com/go-graphite/carbonapi/mstats"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper"
	zipperConfig "github.com/go-graphite/carbonapi/zipper/config"
	"github.com/go-graphite/carbonapi/zipper/types"

	pickle "github.com/lomik/og-rek"
	"github.com/peterbourgon/g2g"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

var defaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

// GraphiteConfig contains configuration bits to send internal stats to Graphite
type GraphiteConfig struct {
	Pattern  string
	Host     string
	Interval time.Duration
	Prefix   string
}

// config contains necessary information for global
var config = struct {
	Backends   []string         `mapstructure:"backends"`
	Backendsv2 types.BackendsV2 `mapstructure:"backendsv2"`
	MaxProcs   int              `mapstructure:"maxProcs"`
	Graphite   GraphiteConfig   `mapstructure:"graphite"`
	GRPCListen string           `mapstructure:"grpcListen"`
	Listen     string           `mapstructure:"listen"`
	Buckets    int              `mapstructure:"buckets"`

	Timeouts          types.Timeouts `mapstructure:"timeouts"`
	KeepAliveInterval time.Duration  `mapstructure:"keepAliveInterval"`

	CarbonSearch   types.CarbonSearch   `mapstructure:"carbonsearch"`
	CarbonSearchV2 types.CarbonSearchV2 `mapstructure:"carbonsearchv2"`

	MaxIdleConnsPerHost int `mapstructure:"maxIdleConnsPerHost"`

	ConcurrencyLimitPerServer  int                `mapstructure:"concurrencyLimit"`
	ExpireDelaySec             int32              `mapstructure:"expireDelaySec"`
	Logger                     []zapwriter.Config `mapstructure:"logger"`
	GraphiteWeb09Compatibility bool               `mapstructure:"graphite09compat"`

	zipper *zipper.Zipper
}{
	MaxProcs: 1,
	Graphite: GraphiteConfig{
		Interval: 60 * time.Second,
		Prefix:   "carbon.zipper",
		Pattern:  "{prefix}.{fqdn}",
	},
	GRPCListen: ":8081",
	Listen:     ":8080",
	Buckets:    10,

	Timeouts: types.Timeouts{
		Render:  10000 * time.Second,
		Find:    10 * time.Second,
		Connect: 200 * time.Millisecond,
	},
	KeepAliveInterval: 30 * time.Second,

	MaxIdleConnsPerHost: 100,

	ExpireDelaySec: 10 * 60, // 10 minutes

	Logger: []zapwriter.Config{defaultLoggerConfig},
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

// set during startup, read-only after that
var searchConfigured = false

const (
	contentTypeJSON          = "application/json"
	contentTypeProtobuf      = "application/x-protobuf"
	contentTypePickle        = "application/pickle"
	contentTypeCarbonAPIv3PB = "application/x-carbonapi-v3-pb"
)

const (
	formatTypeEmpty         = ""
	formatTypePickle        = "pickle"
	formatTypeJSON          = "json"
	formatTypeProtobuf      = "protobuf"
	formatTypeProtobuf3     = "protobuf3"
	formatTypeV2            = "v2"
	formatTypeCarbonAPIV2PB = "carbonapi_v2_pb"
)

func findHandler(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	ctx := req.Context()
	ctx = util.SetUUID(ctx, uuid.String())
	logger := zapwriter.Logger("find").With(
		zap.String("handler", "find"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
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
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
	)

	metrics, stats, err := config.zipper.FindProtoV2(ctx, []string{originalQuery})
	sendStats(stats)
	if err != nil {
		accessLogger.Error("find failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		http.Error(w, "error fetching the data", http.StatusInternalServerError)
		return
	}

	// There should be exactly one match at this moment
	err = merry.Wrap(EncodeFindResponse(format, originalQuery, w, metrics[0].Matches))
	if err != nil {
		http.Error(w, "error marshaling data", http.StatusInternalServerError)
		accessLogger.Error("find failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", "error marshaling data"),
			zap.Duration("runtime_seconds", time.Since(t0)),
			zap.Error(err),
		)
		return
	}
	accessLogger.Info("request served",
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func EncodeFindResponse(format, query string, w http.ResponseWriter, metrics []protov2.GlobMatch) error {
	var err error
	var b []byte
	switch format {
	case formatTypeProtobuf, formatTypeProtobuf3:
		w.Header().Set("Content-Type", contentTypeProtobuf)
		var result protov2.GlobResponse
		result.Name = query
		result.Matches = metrics
		b, err = result.Marshal()
		/* #nosec */
		_, _ = w.Write(b)
	case formatTypeJSON:
		w.Header().Set("Content-Type", contentTypeJSON)
		jEnc := json.NewEncoder(w)
		err = jEnc.Encode(metrics)
	case formatTypeEmpty, formatTypePickle:
		w.Header().Set("Content-Type", contentTypePickle)

		var result []map[string]interface{}

		now := int32(time.Now().Unix() + 60)
		for _, metric := range metrics {
			// Tell graphite-web that we have everything
			var mm map[string]interface{}
			if config.GraphiteWeb09Compatibility {
				// graphite-web 0.9.x
				mm = map[string]interface{}{
					// graphite-web 0.9.x
					"metric_path": metric.Path,
					"isLeaf":      metric.IsLeaf,
				}
			} else {
				// graphite-web 1.0
				interval := &intervalset.IntervalSet{Start: 0, End: now}
				mm = map[string]interface{}{
					"is_leaf":   metric.IsLeaf,
					"path":      metric.Path,
					"intervals": interval,
				}
			}
			result = append(result, mm)
		}

		pEnc := pickle.NewEncoder(w)
		err = pEnc.Encode(result)
	}
	return err
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
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
	)

	logger.Debug("got render request",
		zap.String("request", req.URL.RequestURI()),
	)

	Metrics.RenderRequests.Add(1)

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "render"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
	)

	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse arguments", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "failed to parse arguments"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}
	targets := req.Form["target"]
	format := req.FormValue("format")
	accessLogger = accessLogger.With(
		zap.String("format", format),
		zap.Strings("targets", targets),
	)

	from, err := strconv.Atoi(req.FormValue("from"))
	if err != nil {
		http.Error(w, "from is not a integer", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "from is not a integer"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}
	until, err := strconv.Atoi(req.FormValue("until"))
	if err != nil {
		http.Error(w, "until is not a integer", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "until is not a integer"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	if len(targets) == 0 {
		http.Error(w, "empty target", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", "empty target"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	metrics, stats, err := config.zipper.FetchProtoV2(ctx, targets, int32(from), int32(until))
	sendStats(stats)
	if err != nil {
		http.Error(w, "error fetching the data", http.StatusInternalServerError)
		accessLogger.Error("request failed",
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.String("reason", err.Error()),
			zap.Int("http_code", http.StatusInternalServerError),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}

	var b []byte
	switch format {
	case formatTypeProtobuf, formatTypeProtobuf3:
		w.Header().Set("Content-Type", contentTypeProtobuf)
		b, err = metrics.Marshal()

		memoryUsage += len(b)
		/* #nosec */
		_, _ = w.Write(b)
	case formatTypeJSON:
		presponse := createRenderResponse(metrics, nil)
		w.Header().Set("Content-Type", contentTypeJSON)
		e := json.NewEncoder(w)
		err = e.Encode(presponse)
	case formatTypeEmpty, formatTypePickle:
		presponse := createRenderResponse(metrics, pickle.None{})
		w.Header().Set("Content-Type", contentTypePickle)
		e := pickle.NewEncoder(w)
		err = e.Encode(presponse)
	}

	if err != nil {
		http.Error(w, "error marshaling data", http.StatusInternalServerError)
		accessLogger.Error("render failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", "error marshaling data"),
			zap.Duration("runtime_seconds", time.Since(t0)),
			zap.Int("memory_usage_bytes", memoryUsage),
			zap.Error(err),
		)
		return
	}

	accessLogger.Info("request served",
		zap.Int("memory_usage_bytes", memoryUsage),
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func createRenderResponse(metrics *protov2.MultiFetchResponse, missing interface{}) []map[string]interface{} {

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
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
	)

	logger.Debug("request",
		zap.String("request", req.URL.RequestURI()),
	)

	Metrics.InfoRequests.Add(1)

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "info"),
		zap.String("carbonzipper_uuid", uuid.String()),
		zap.String("carbonapi_uuid", util.GetUUID(ctx)),
	)
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse arguments", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.String("reason", "failed to parse arguments"),
			zap.Int("http_code", http.StatusBadRequest),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		return
	}
	targets := req.Form["target"]
	format := req.FormValue("format")

	accessLogger = accessLogger.With(
		zap.Strings("targets", targets),
		zap.String("format", format),
	)

	if len(targets) == 0 {
		accessLogger.Error("info failed",
			zap.Int("http_code", http.StatusBadRequest),
			zap.String("reason", "empty target"),
			zap.Duration("runtime_seconds", time.Since(t0)),
		)
		http.Error(w, "info: empty target", http.StatusBadRequest)
		return
	}

	haveNonFatalErrors := false
	var b []byte
	if format == formatTypeV2 || format == formatTypeCarbonAPIV2PB || format == formatTypeProtobuf || format == formatTypeProtobuf3 {
		var result *protov2.ZipperInfoResponse
		var stats *types.Stats
		result, stats, err = config.zipper.InfoProtoV2(ctx, targets)
		sendStats(stats)
		if err != nil && err != types.ErrNonFatalErrors {
			accessLogger.Error("info failed",
				zap.Int("http_code", http.StatusInternalServerError),
				zap.String("reason", err.Error()),
				zap.Duration("runtime_seconds", time.Since(t0)),
			)
			http.Error(w, "info: error processing request", http.StatusInternalServerError)
			return
		}
		//                                err := sender.SendCluster(fg.Cluster, fg.Server)
		if err == types.ErrNonFatalErrors {
			haveNonFatalErrors = true
		}

		w.Header().Set("Content-Type", contentTypeProtobuf)
		b, err = result.Marshal()
		_, _ = w.Write(b)
	} else {
		var result *protov3.ZipperInfoResponse
		var stats *types.Stats
		result, stats, err = config.zipper.InfoProtoV3(ctx, &protov3.MultiGlobRequest{Metrics: targets})
		sendStats(stats)
		if err != nil && err != types.ErrNonFatalErrors {
			accessLogger.Error("info failed",
				zap.Int("http_code", http.StatusInternalServerError),
				zap.String("reason", err.Error()),
				zap.Duration("runtime_seconds", time.Since(t0)),
			)
			http.Error(w, "info: error processing request", http.StatusInternalServerError)
			return
		}

		if err == types.ErrNonFatalErrors {
			haveNonFatalErrors = true
		}

		switch format {
		case "v3", "carbonapi_v3_pb":
			w.Header().Set("Content-Type", contentTypeCarbonAPIv3PB)
			b, err = result.Marshal()
			/* #nosec */
			_, _ = w.Write(b)
		case "", "json":
			w.Header().Set("Content-Type", contentTypeJSON)
			jEnc := json.NewEncoder(w)
			err = jEnc.Encode(result)
		}
	}

	if err != nil {
		http.Error(w, "error marshaling data", http.StatusInternalServerError)
		accessLogger.Error("info failed",
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", "error marshaling data"),
			zap.Duration("runtime_seconds", time.Since(t0)),
			zap.Error(err),
		)
		return
	}
	accessLogger.Info("request served",
		zap.Bool("have_non_fatal_errors", haveNonFatalErrors),
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

	/* #nosec */
	_, _ = fmt.Fprintf(w, "Ok\n")
	accessLogger.Info("lb request served",
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime_seconds", time.Since(t0)),
	)
}

func main() {
	err := zapwriter.ApplyConfig([]zapwriter.Config{defaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")

	}
	logger := zapwriter.Logger("main")

	configFile := flag.String("config", "", "config file (yaml)")
	pidFile := flag.String("pid", "", "pidfile (default: empty, don't create pidfile)")
	envPrefix := flag.String("envprefix", "CARBONZIPPER_", "Prefix for environment variables override")
	if *envPrefix == "" {
		logger.Fatal("empty prefix is not suppoerted due to possible collisions with OS environment variables")
	}

	flag.Parse()

	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewString("BuildVersion").Set(BuildVersion)

	if *configFile == "" {
		logger.Fatal("missing config file option")
	}

	cfg, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatal("unable to load config file:",
			zap.Error(err),
		)
	}

	if strings.HasSuffix(*configFile, ".toml") {
		logger.Info("will parse config as toml",
			zap.String("config_file", *configFile),
		)
		viper.SetConfigType("TOML")
	} else {
		logger.Info("will parse config as yaml",
			zap.String("config_file", *configFile),
		)
		viper.SetConfigType("YAML")
	}
	err = viper.ReadConfig(bytes.NewBuffer(cfg))
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.String("config_path", *configFile),
			zap.Error(err),
		)
	}

	if *envPrefix != "" {
		viper.SetEnvPrefix(*envPrefix)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.String("config_path", *configFile),
			zap.Error(err),
		)
	}

	if len(config.Backends) == 0 && len(config.Backendsv2.Backends) == 0 {
		logger.Fatal("no Backends loaded -- exiting")
	}

	err = zapwriter.ApplyConfig(config.Logger)
	if err != nil {
		logger.Fatal("Failed to apply config",
			zap.Any("config", config.Logger),
			zap.Error(err),
		)
	}

	// Should print nicer stack traces in case of unexpected panic.
	defer func() {
		if r := recover(); r != nil {
			logger.Fatal("Recovered from unhandled panic",
				zap.Stack("stacktrace"),
			)
		}
	}()

	searchConfigured = (len(config.CarbonSearch.Prefix) > 0 && len(config.CarbonSearch.Backend) > 0) || (len(config.CarbonSearchV2.Prefix) > 0 && len(config.CarbonSearchV2.Backends) > 0)

	logger = zapwriter.Logger("main")
	logger.Info("starting carbonzipper",
		zap.String("build_version", BuildVersion),
		zap.Bool("carbonsearch_configured", searchConfigured),
		zap.Any("config", config),
	)

	runtime.GOMAXPROCS(config.MaxProcs)

	// +1 to track every over the number of buckets we track
	timeBuckets = make([]int64, config.Buckets+1)

	httputil.PublishTrackedConnections("httptrack")
	expvar.Publish("requestBuckets", expvar.Func(renderTimeBuckets))

	// export config via expvars
	expvar.Publish("config", expvar.Func(func() interface{} { return config }))

	/* Configure zipper */
	// set up caches
	zipperConfig := &zipperConfig.Config{
		ConcurrencyLimitPerServer: config.ConcurrencyLimitPerServer,
		MaxIdleConnsPerHost:       config.MaxIdleConnsPerHost,
		Backends:                  config.Backends,
		BackendsV2:                config.Backendsv2,
		ExpireDelaySec:            config.ExpireDelaySec,

		CarbonSearch:      config.CarbonSearch,
		CarbonSearchV2:    config.CarbonSearchV2,
		Timeouts:          config.Timeouts,
		KeepAliveInterval: config.KeepAliveInterval,
	}

	/*
		TODO(civil): Restore those metrics
		Metrics.CacheSize = expvar.Func(func() interface{} { return zipperConfig.PathCache.ECSize() })
		expvar.Publish("cacheSize", Metrics.CacheSize)

		Metrics.CacheItems = expvar.Func(func() interface{} { return zipperConfig.PathCache.ECItems() })
		expvar.Publish("cacheItems", Metrics.CacheItems)

		Metrics.SearchCacheSize = expvar.Func(func() interface{} { return zipperConfig.SearchCache.ECSize() })
		expvar.Publish("searchCacheSize", Metrics.SearchCacheSize)

		Metrics.SearchCacheItems = expvar.Func(func() interface{} { return zipperConfig.SearchCache.ECItems() })
		expvar.Publish("searchCacheItems", Metrics.SearchCacheItems)
	*/

	config.zipper, err = zipper.NewZipper(sendStats, zipperConfig, zapwriter.Logger("zipper"))
	if err != nil {
		logger.Fatal("failed to create zipper instance",
			zap.Error(err),
		)
	}

	http.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(util.ParseCtx(findHandler, util.HeaderUUIDAPI), bucketRequestTimes)))
	http.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(util.ParseCtx(renderHandler, util.HeaderUUIDAPI), bucketRequestTimes)))
	http.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(util.ParseCtx(infoHandler, util.HeaderUUIDAPI), bucketRequestTimes)))
	http.HandleFunc("/lb_check", lbCheckHandler)

	// nothing in the config? check the environment
	if config.Graphite.Host == "" {
		if host := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); host != ":" {
			config.Graphite.Host = host
		}
	}

	if config.Graphite.Pattern == "" {
		config.Graphite.Pattern = "{prefix}.{fqdn}"
	}

	if config.Graphite.Prefix == "" {
		config.Graphite.Prefix = "carbon.zipper"
	}

	// only register g2g if we have a graphite host
	if config.Graphite.Host != "" {
		// register our metrics with graphite
		graphite := g2g.NewGraphite(config.Graphite.Host, config.Graphite.Interval, 10*time.Second)

		/* #nosec */
		hostname, _ := os.Hostname()
		hostname = strings.ReplaceAll(hostname, ".", "_")

		prefix := config.Graphite.Prefix

		pattern := config.Graphite.Pattern
		pattern = strings.ReplaceAll(pattern, "{prefix}", prefix)
		pattern = strings.ReplaceAll(pattern, "{fqdn}", hostname)

		graphite.Register(fmt.Sprintf("%s.find_requests", pattern), Metrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.find_errors", pattern), Metrics.FindErrors)

		graphite.Register(fmt.Sprintf("%s.render_requests", pattern), Metrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.render_errors", pattern), Metrics.RenderErrors)

		graphite.Register(fmt.Sprintf("%s.info_requests", pattern), Metrics.InfoRequests)
		graphite.Register(fmt.Sprintf("%s.info_errors", pattern), Metrics.InfoErrors)

		graphite.Register(fmt.Sprintf("%s.timeouts", pattern), Metrics.Timeouts)

		for i := 0; i <= config.Buckets; i++ {
			graphite.Register(fmt.Sprintf("%s.requests_in_%dms_to_%dms", pattern, i*100, (i+1)*100), bucketEntry(i))
		}

		/* TODO(civil): Find a way to return that data
		graphite.Register(fmt.Sprintf("%s.cache_size", pattern), Metrics.CacheSize)
		graphite.Register(fmt.Sprintf("%s.cache_items", pattern), Metrics.CacheItems)

		graphite.Register(fmt.Sprintf("%s.search_cache_size", pattern), Metrics.SearchCacheSize)
		graphite.Register(fmt.Sprintf("%s.search_cache_items", pattern), Metrics.SearchCacheItems)
		*/

		graphite.Register(fmt.Sprintf("%s.cache_hits", pattern), Metrics.CacheHits)
		graphite.Register(fmt.Sprintf("%s.cache_misses", pattern), Metrics.CacheMisses)

		graphite.Register(fmt.Sprintf("%s.search_cache_hits", pattern), Metrics.SearchCacheHits)
		graphite.Register(fmt.Sprintf("%s.search_cache_misses", pattern), Metrics.SearchCacheMisses)

		go mstats.Start(config.Graphite.Interval)

		graphite.Register(fmt.Sprintf("%s.alloc", pattern), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.total_alloc", pattern), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.num_gc", pattern), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.pause_ns", pattern), &mstats.PauseNS)
	}

	if *pidFile != "" {
		pidfile.SetPidfilePath(*pidFile)
		err = pidfile.Write()
		if err != nil {
			logger.Fatal("error during pidfile.Write()",
				zap.Error(err),
			)
		}
	}

	if len(config.GRPCListen) > 0 {
		srv, err := NewGRPCServer(config.GRPCListen)
		if err != nil {
			logger.Fatal("failed to start gRPC server",
				zap.Error(err),
			)
		}
		go srv.serve()
	}

	srv := &http.Server{
		Addr:    config.Listen,
		Handler: nil,
	}
	err = srv.ListenAndServe()

	if err != nil {
		logger.Fatal("error during starting web server",
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

	if bucket < config.Buckets {
		atomic.AddInt64(&timeBuckets[bucket], 1)
	} else {
		// Too big? Increment overflow bucket and log
		atomic.AddInt64(&timeBuckets[config.Buckets], 1)
		logger.Warn("Slow Request",
			zap.Duration("time", t),
			zap.String("url", req.URL.String()),
		)
	}
}

func sendStats(stats *types.Stats) {
	if stats == nil {
		return
	}
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
