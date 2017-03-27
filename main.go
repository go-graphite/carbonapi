package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dgryski/carbonapi/expr"
	pb "github.com/dgryski/carbonzipper/carbonzipperpb3"
	"github.com/dgryski/carbonzipper/mstats"

	"github.com/bradfitz/gomemcache/memcache"
	ecache "github.com/dgryski/go-expirecache"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	"github.com/gorilla/handlers"
	"github.com/peterbourgon/g2g"

	"github.com/lomik/zapwriter"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// Metrics contains exported counters and values for graphite
var Metrics = struct {
	Requests         *expvar.Int
	RequestCacheHits *expvar.Int

	RenderRequests *expvar.Int

	MemcacheTimeouts *expvar.Int

	CacheSize  expvar.Func
	CacheItems expvar.Func
}{
	Requests:         expvar.NewInt("requests"),
	RequestCacheHits: expvar.NewInt("request_cache_hits"),

	RenderRequests: expvar.NewInt("render_requests"),

	MemcacheTimeouts: expvar.NewInt("memcache_timeouts"),
}

// BuildVersion is provided to be overridden at build time. Eg. go build -ldflags -X 'main.BuildVersion=...'
var BuildVersion = "(development build)"

// for testing
var timeNow = time.Now

func writeResponse(w http.ResponseWriter, b []byte, format string, jsonp string) {

	switch format {
	case "json":
		if jsonp != "" {
			w.Header().Set("Content-Type", contentTypeJavaScript)
			w.Write([]byte(jsonp))
			w.Write([]byte{'('})
			w.Write(b)
			w.Write([]byte{')'})
		} else {
			w.Header().Set("Content-Type", contentTypeJSON)
			w.Write(b)
		}
	case "protobuf":
		w.Header().Set("Content-Type", contentTypeProtobuf)
		w.Write(b)
	case "raw":
		w.Header().Set("Content-Type", contentTypeRaw)
		w.Write(b)
	case "pickle":
		w.Header().Set("Content-Type", contentTypePickle)
		w.Write(b)
	case "csv":
		w.Header().Set("Content-Type", contentTypeCSV)
		w.Write(b)
	case "png":
		w.Header().Set("Content-Type", contentTypePNG)
		w.Write(b)
	case "svg":
		w.Header().Set("Content-Type", contentTypeSVG)
		w.Write(b)
	}
}

const (
	contentTypeJSON       = "application/json"
	contentTypeProtobuf   = "application/x-protobuf"
	contentTypeJavaScript = "text/javascript"
	contentTypeRaw        = "text/plain"
	contentTypePickle     = "application/pickle"
	contentTypePNG        = "image/png"
	contentTypeCSV        = "text/csv"
	contentTypeSVG        = "image/svg+xml"
)

type renderStats struct {
	zipperRequests int
}

func buildParseErrorString(target, e string, err error) string {
	msg := fmt.Sprintf("%s\n\n%-20s: %s\n", http.StatusText(http.StatusBadRequest), "Target", target)
	if err != nil {
		msg += fmt.Sprintf("%-20s: %s\n", "Error", err.Error())
	}
	if e != "" {
		msg += fmt.Sprintf("%-20s: %s\n%-20s: %s\n",
			"Parsed so far", target[0:len(target)-len(e)],
			"Could not parse", e)
	}
	return msg
}

var errBadTime = errors.New("bad time")

// parseTime parses a time and returns hours and minutes
func parseTime(s string) (hour, minute int, err error) {

	switch s {
	case "midnight":
		return 0, 0, nil
	case "noon":
		return 12, 0, nil
	case "teatime":
		return 16, 0, nil
	}

	parts := strings.Split(s, ":")

	if len(parts) != 2 {
		return 0, 0, errBadTime
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errBadTime
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errBadTime
	}

	return hour, minute, nil
}

var timeFormats = []string{"20060102", "01/02/06"}

// dateParamToEpoch turns a passed string parameter into a unix epoch
func dateParamToEpoch(s string, qtz string, d int64) int32 {

	if s == "" {
		// return the default if nothing was passed
		return int32(d)
	}

	// relative timestamp
	if s[0] == '-' {
		offset, err := expr.IntervalString(s, -1)
		if err != nil {
			return int32(d)
		}

		return int32(timeNow().Add(time.Duration(offset) * time.Second).Unix())
	}

	switch s {
	case "now":
		return int32(timeNow().Unix())
	case "midnight", "noon", "teatime":
		yy, mm, dd := timeNow().Date()
		hh, min, _ := parseTime(s) // error ignored, we know it's valid
		dt := time.Date(yy, mm, dd, hh, min, 0, 0, Config.DefaultTimeZone)
		return int32(dt.Unix())
	}

	sint, err := strconv.Atoi(s)
	// need to check that len(s) > 8 to avoid turning 20060102 into seconds
	if err == nil && len(s) > 8 {
		return int32(sint) // We got a timestamp so returning it
	}

	s = strings.Replace(s, "_", " ", 1) // Go can't parse _ in date strings

	var ts, ds string
	split := strings.Fields(s)

	switch {
	case len(split) == 1:
		ds = s
	case len(split) == 2:
		ts, ds = split[0], split[1]
	case len(split) > 2:
		return int32(d)
	}

	var tz = Config.DefaultTimeZone
	if qtz != "" {
		if z, err := time.LoadLocation(qtz); err != nil {
			tz = z
		}
	}

	var t time.Time
dateStringSwitch:
	switch ds {
	case "today":
		t = timeNow()
		// nothing
	case "yesterday":
		t = timeNow().AddDate(0, 0, -1)
	case "tomorrow":
		t = timeNow().AddDate(0, 0, 1)
	default:
		for _, format := range timeFormats {
			t, err = time.ParseInLocation(format, ds, tz)
			if err == nil {
				break dateStringSwitch
			}
		}

		return int32(d)
	}

	var hour, minute int
	if ts != "" {
		hour, minute, _ = parseTime(ts)
		// defaults to hour=0, minute=0 on error, which is midnight, which is fine for now
	}

	yy, mm, dd := t.Date()
	t = time.Date(yy, mm, dd, hour, minute, 0, 0, Config.DefaultTimeZone)

	return int32(t.Unix())
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	logger := zapwriter.Logger("render").With(zap.String("uuid", uuid.String()))
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := context.WithValue(context.Background(), "carbonapi_uuid", uuid.String())
	ctx = context.WithValue(ctx, "carbonapi_handler", "render")
	ctx = context.WithValue(ctx, "carbonapi_request_url", r.URL.RequestURI())
	ctx = context.WithValue(ctx, "carbonapi_referer", r.Referer())
	username, _, _ := r.BasicAuth()
	ctx = context.WithValue(ctx, "carbonapi_username", username)

	zipperRequests := 0

	Metrics.Requests.Add(1)
	accessLogger := zapwriter.Logger("access").With(
		zap.String("url", r.URL.RequestURI()),
		zap.String("carbonapi_uuid", uuid.String()),
	)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.Duration("runtime", time.Since(t0)),
			zap.Int("http_code", http.StatusBadRequest),
		)
		return
	}

	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")
	format := r.FormValue("format")
	useCache := !expr.TruthyBool(r.FormValue("noCache"))

	var jsonp string

	if format == "json" {
		// TODO(dgryski): check jsonp only has valid characters
		jsonp = r.FormValue("jsonp")
	}

	if format == "" && (expr.TruthyBool(r.FormValue("rawData")) || expr.TruthyBool(r.FormValue("rawdata"))) {
		format = "raw"
	}

	if format == "" {
		format = "png"
	}

	cacheTimeout := int32(60)

	if tstr := r.FormValue("cacheTimeout"); tstr != "" {
		t, err := strconv.Atoi(tstr)
		if err != nil {
			logger.Error("failed to parse cacheTimeout",
				zap.String("cache_string", tstr),
				zap.Error(err),
			)
		} else {
			cacheTimeout = int32(t)
		}
	}

	// make sure the cache key doesn't say noCache, because it will never hit
	r.Form.Del("noCache")

	// jsonp callback names are frequently autogenerated and hurt our cache
	r.Form.Del("jsonp")

	// Strip some cache-busters.  If you don't want to cache, use noCache=1
	r.Form.Del("_salt")
	r.Form.Del("_ts")
	r.Form.Del("_t") // Used by jquery.graphite.js

	cacheKey := r.Form.Encode()

	// normalize from and until values
	qtz := r.FormValue("tz")
	from32 := dateParamToEpoch(from, qtz, timeNow().Add(-24*time.Hour).Unix())
	until32 := dateParamToEpoch(until, qtz, timeNow().Unix())

	accessLogger = accessLogger.With(
		zap.String("format", format),
		zap.Bool("use_cache", useCache),
		zap.Strings("targets", targets),
		zap.String("from_raw", from),
		zap.String("until_raw", until),
		zap.Int32("from", from32),
		zap.Int32("until", until32),
		zap.String("tz", qtz),
		zap.Int32("cache_timeout", cacheTimeout),
	)

	if response, ok := Config.queryCache.get(cacheKey); useCache && ok {
		Metrics.RequestCacheHits.Add(1)
		writeResponse(w, response, format, jsonp)
		accessLogger.Info("request served",
			zap.Bool("from_cache", true),
			zap.Int("http_code", http.StatusOK),
			zap.Duration("runtime", time.Since(t0)),
		)
		return
	}

	if from32 == until32 {
		http.Error(w, "Invalid empty time range", http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.String("reason", "Invalid empty time range"),
			zap.Duration("runtime", time.Since(t0)),
			zap.Int("http_code", http.StatusBadRequest),
		)
		return
	}

	var results []*expr.MetricData
	var errors []string
	metricMap := make(map[expr.MetricRequest][]*expr.MetricData)

	for _, target := range targets {

		exp, e, err := expr.ParseExpr(target)

		if err != nil || e != "" {
			msg := buildParseErrorString(target, e, err)
			http.Error(w, msg, http.StatusBadRequest)
			accessLogger.Error("request failed",
				zap.String("reason", msg),
				zap.Duration("runtime", time.Since(t0)),
				zap.Int("http_code", http.StatusBadRequest),
			)
			return
		}

		for _, m := range exp.Metrics() {

			mfetch := m
			mfetch.From += from32
			mfetch.Until += until32

			if _, ok := metricMap[mfetch]; ok {
				// already fetched this metric for this request
				continue
			}

			// For each metric returned in the Find response, query Render
			Metrics.RenderRequests.Add(1)
			Config.Limiter.enter()
			zipperRequests++

			r, err := Config.zipper.Render(ctx, m.Metric, mfetch.From, mfetch.Until)
			if err != nil {
				logger.Error("render error",
					zap.String("metric", m.Metric),
					zap.Error(err),
				)
				Config.Limiter.leave()
				continue
			} else {
				metricMap[mfetch] = r
			}
			Config.Limiter.leave()

			expr.SortMetrics(metricMap[mfetch], mfetch)

		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic during eval:",
						zap.String("cache_key", cacheKey),
						zap.Stack("stack"),
					)
				}
			}()
			exprs, err := expr.EvalExpr(exp, from32, until32, metricMap)
			if err != nil && err != expr.ErrSeriesDoesNotExist {
				errors = append(errors, target+": "+err.Error())
				return
			}
			results = append(results, exprs...)
		}()
	}

	if len(errors) > 0 {
		errors = append([]string{"Encountered the following errors:"}, errors...)
		http.Error(w, strings.Join(errors, "\n"), http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.String("reason", "encoundered multiple errors"),
			zap.Strings("errors", errors),
			zap.Duration("runtime", time.Since(t0)),
			zap.Int("http_code", http.StatusBadRequest),
		)
		return
	}

	var body []byte

	switch format {
	case "json":
		if maxDataPoints, _ := strconv.Atoi(r.FormValue("maxDataPoints")); maxDataPoints != 0 {
			expr.ConsolidateJSON(maxDataPoints, results)
		}

		body = expr.MarshalJSON(results)
	case "protobuf":
		body, err = expr.MarshalProtobuf(results)
		if err != nil {
			logger.Info("request failed",
				zap.Int("http_code", http.StatusInternalServerError),
				zap.String("reason", err.Error()),
				zap.Duration("runtime", time.Since(t0)),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "raw":
		body = expr.MarshalRaw(results)
	case "csv":
		body = expr.MarshalCSV(results)
	case "pickle":
		body = expr.MarshalPickle(results)
	case "png":
		body = expr.MarshalPNG(r, results)
	case "svg":
		body = expr.MarshalSVG(r, results)
	}

	writeResponse(w, body, format, jsonp)

	if len(results) != 0 {
		Config.queryCache.set(cacheKey, body, cacheTimeout)
	}

	accessLogger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
		zap.Int("zipper_requests", zipperRequests),
	)
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	logger := zapwriter.Logger("find").With(zap.String("uuid", uuid.String()))
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := context.WithValue(r.Context(), "carbonapi_uuid", uuid.String())
	ctx = context.WithValue(ctx, "carbonapi_request_url", r.URL.RequestURI())
	ctx = context.WithValue(ctx, "carbonapi_referer", r.Referer())
	username, _, _ := r.BasicAuth()
	ctx = context.WithValue(ctx, "carbonapi_username", username)

	format := r.FormValue("format")
	jsonp := r.FormValue("jsonp")

	query := r.FormValue("query")

	if query == "" {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		logger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.String("uuid", uuid.String()),
			zap.Int("http_code", http.StatusBadRequest),
			zap.String("reason", "missing parameter `query`"),
			zap.Duration("runtime", time.Since(t0)),
		)
		return
	}

	if format == "" {
		format = "treejson"
	}

	globs, err := Config.zipper.Find(ctx, query)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		logger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime", time.Since(t0)),
		)
		return
	}

	var b []byte
	switch format {
	case "treejson", "json":
		b, err = findTreejson(globs)
		format = "json"
	case "completer":
		b, err = findCompleter(globs)
		format = "json"
	case "raw":
		b, err = findList(globs)
		format = "raw"
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		logger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime", time.Since(t0)),
		)
		return
	}

	writeResponse(w, b, format, jsonp)

	logger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Int("http_code", http.StatusOK),
		zap.Duration("runtime", time.Since(t0)),
	)
}

type completer struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsLeaf string `json:"is_leaf"`
}

func findCompleter(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var complete = make([]completer, 0)

	for _, g := range globs.GetMatches() {
		c := completer{
			Path: g.GetPath(),
		}

		if g.GetIsLeaf() {
			c.IsLeaf = "1"
		} else {
			c.IsLeaf = "0"
		}

		i := strings.LastIndex(c.Path, ".")

		if i != -1 {
			c.Name = c.Path[i+1:]
		}

		complete = append(complete, c)
	}

	err := json.NewEncoder(&b).Encode(struct {
		Metrics []completer `json:"metrics"`
	}{
		Metrics: complete},
	)
	return b.Bytes(), err
}

func findList(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	for _, g := range globs.GetMatches() {

		var dot string
		// make sure non-leaves end in one dot
		if !g.GetIsLeaf() && !strings.HasSuffix(g.GetPath(), ".") {
			dot = "."
		}

		fmt.Fprintln(&b, g.GetPath()+dot)
	}

	return b.Bytes(), nil
}

type treejson struct {
	AllowChildren int            `json:"allowChildren"`
	Expandable    int            `json:"expandable"`
	Leaf          int            `json:"leaf"`
	ID            string         `json:"id"`
	Text          string         `json:"text"`
	Context       map[string]int `json:"context"` // unused
}

var treejsonContext = make(map[string]int)

func findTreejson(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var tree = make([]treejson, 0)

	seen := make(map[string]struct{})

	basepath := globs.GetName()

	if i := strings.LastIndex(basepath, "."); i != -1 {
		basepath = basepath[:i+1]
	}

	for _, g := range globs.GetMatches() {

		name := g.GetPath()

		if i := strings.LastIndex(name, "."); i != -1 {
			name = name[i+1:]
		}

		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		t := treejson{
			ID:      basepath + name,
			Context: treejsonContext,
			Text:    name,
		}

		if g.GetIsLeaf() {
			t.Leaf = 1
		} else {
			t.AllowChildren = 1
			t.Expandable = 1
		}

		tree = append(tree, t)
	}

	err := json.NewEncoder(&b).Encode(tree)
	return b.Bytes(), err
}

func passthroughHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	logger := zapwriter.Logger("passthrough").With(zap.String("uuid", uuid.String()))
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := context.WithValue(r.Context(), "carbonapi_uuid", uuid.String())
	ctx = context.WithValue(ctx, "carbonapi_handler", "passthrough")
	ctx = context.WithValue(ctx, "carbonapi_request_url", r.URL.RequestURI())
	ctx = context.WithValue(ctx, "carbonapi_referer", r.Referer())
	username, _, _ := r.BasicAuth()
	ctx = context.WithValue(ctx, "carbonapi_username", username)
	var data []byte
	var err error

	if data, err = Config.zipper.Passthrough(ctx, r.URL.RequestURI()); err != nil {
		logger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.Duration("runtime", time.Since(t0)),
			zap.Int("http_code", http.StatusBadRequest),
		)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Write(data)
	logger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
	)
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	logger := zapwriter.Logger("lbcheck")

	w.Write([]byte("Ok\n"))

	logger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
	)
}

var usageMsg = []byte(`
supported requests:
	/render/?target=
	/metrics/find/?query=
	/info/?target=
`)

func usageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(usageMsg)
}

var DefaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

type cacheConfig struct {
	Type             string
	Size             int      `yaml:"size_mb"`
	MemcachedServers []string `yaml:"memcachedServers"`
}

type graphiteConfig struct {
	Host     string
	Interval time.Duration
	Prefix   string
}

var Config = struct {
	Logger          []zapwriter.Config
	ZipperUrl       string `yaml:"zipper"`
	Listen          string
	Concurency      int
	Cache           cacheConfig
	Cpus            int
	TimezoneString  string `yaml:"tz"`
	Graphite        graphiteConfig
	IdleConnections int
	PidFile         string

	queryCache bytesCache

	DefaultTimeZone *time.Location

	// Zipper is API entry to carbonzipper
	zipper zipper

	// Limiter limits concurrent zipper requests
	Limiter limiter
}{
	ZipperUrl:  "http://localhost:8080",
	Listen:     "[::]:8081",
	Concurency: 20,
	Cache: cacheConfig{
		Type: "mem",
	},
	TimezoneString: "",
	Graphite: graphiteConfig{
		Host:     "",
		Interval: 60 * time.Second,
		Prefix:   "carbon.api",
	},
	Cpus:            0,
	IdleConnections: 10,
	PidFile:         "",

	DefaultTimeZone: time.Local,
	Logger:          []zapwriter.Config{DefaultLoggerConfig},
}

func main() {
	err := zapwriter.ApplyConfig([]zapwriter.Config{DefaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")

	}
	logger := zapwriter.Logger("main")

	configPath := flag.String("config", "carbonapi.yaml", "Path to the `config file`.")

	flag.Parse()

	if *configPath == "" {
		logger.Fatal("Can't run without a config file")
	}

	bytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		logger.Fatal("error reading config file",
			zap.String("config_path", *configPath),
			zap.Error(err),
		)
	}

	err = yaml.Unmarshal(bytes, &Config)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.String("config_path", *configPath),
			zap.Error(err),
		)
	}

	err = zapwriter.ApplyConfig(Config.Logger)
	if err != nil {
		logger.Fatal("failed to initialize logger with requested configuration",
			zap.Any("configuration", Config.Logger),
			zap.Error(err),
		)

	}

	expvar.NewString("BuildVersion").Set(BuildVersion)

	Config.Limiter = newLimiter(Config.Concurency)

	if _, err := url.Parse(Config.ZipperUrl); err != nil {
		logger.Fatal("unable to parze zipper", zap.Error(err))
	}

	Config.zipper = zipper{
		z: Config.ZipperUrl,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: Config.IdleConnections,
			},
		},
	}

	switch Config.Cache.Type {
	case "memcache":
		if len(Config.Cache.MemcachedServers) == 0 {
			logger.Fatal("memcache cache requested but no memcache servers provided")
		}

		logger.Info("memcached configured",
			zap.Strings("servers", Config.Cache.MemcachedServers),
		)
		Config.queryCache = &memcachedCache{client: memcache.New(Config.Cache.MemcachedServers...)}
	case "mem":
		qcache := &expireCache{ec: ecache.New(uint64(Config.Cache.Size * 1024 * 1024))}
		Config.queryCache = qcache
		go Config.queryCache.(*expireCache).ec.ApproximateCleaner(10 * time.Second)

		Metrics.CacheSize = expvar.Func(func() interface{} {
			return qcache.ec.Size()
		})
		expvar.Publish("cache_size", Metrics.CacheSize)

		Metrics.CacheItems = expvar.Func(func() interface{} {
			return qcache.ec.Items()
		})
		expvar.Publish("cache_items", Metrics.CacheItems)

	case "null":
		Config.queryCache = &nullCache{}
	}

	if Config.TimezoneString != "" {
		fields := strings.Split(Config.TimezoneString, ",")
		if len(fields) != 2 {
			logger.Fatal("unexpected amount of fields in tz",
				zap.Int("fields_got", len(fields)),
				zap.Int("fields_expected", 2),
			)
		}

		var err error
		offs, err := strconv.Atoi(fields[1])
		if err != nil {
			logger.Fatal("unable to parse seconds",
				zap.String("field[1]", fields[1]),
				zap.Error(err),
			)
		}

		Config.DefaultTimeZone = time.FixedZone(fields[0], offs)
		logger.Info("using fixed timezone",
			zap.String("timezone", Config.DefaultTimeZone.String()),
			zap.Int("offset", offs),
		)
	}

	if Config.Cpus != 0 {
		runtime.GOMAXPROCS(Config.Cpus)
	}

	var host string
	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || Config.Graphite.Host != "" {
		switch {
		case envhost != ":" && Config.Graphite.Host != "":
			host = Config.Graphite.Host
		case envhost != ":":
			host = envhost
		case Config.Graphite.Host != "":
			host = Config.Graphite.Host
		}
	}

	logger.Info("starting carbonapi",
		zap.String("build_version", BuildVersion),
		zap.String("listen_address", Config.Listen),
		zap.Any("zipper", Config.ZipperUrl),
		zap.Int("GOMAXPROCS", Config.Cpus),
		zap.Any("graphite_configuration", Config.Graphite),
		zap.Any("cache", Config.Cache),
	)

	if host != "" {
		// register our metrics with graphite
		graphite := g2g.NewGraphite(host, Config.Graphite.Interval, 10*time.Second)

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		graphite.Register(fmt.Sprintf("%s.%s.requests", Config.Graphite.Prefix, hostname), Metrics.Requests)
		graphite.Register(fmt.Sprintf("%s.%s.request_cache_hits", Config.Graphite.Prefix, hostname), Metrics.RequestCacheHits)

		graphite.Register(fmt.Sprintf("%s.%s.render_requests", Config.Graphite.Prefix, hostname), Metrics.RenderRequests)

		graphite.Register(fmt.Sprintf("%s.%s.memcache_timeouts", Config.Graphite.Prefix, hostname), Metrics.MemcacheTimeouts)

		if Metrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("%s.%s.cache_size", Config.Graphite.Prefix, hostname), Metrics.CacheSize)
			graphite.Register(fmt.Sprintf("%s.%s.cache_items", Config.Graphite.Prefix, hostname), Metrics.CacheItems)
		}

		go mstats.Start(Config.Graphite.Interval)

		graphite.Register(fmt.Sprintf("%s.%s.alloc", Config.Graphite.Prefix, hostname), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.%s.total_alloc", Config.Graphite.Prefix, hostname), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.%s.num_gc", Config.Graphite.Prefix, hostname), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.%s.pause_ns", Config.Graphite.Prefix, hostname), &mstats.PauseNS)

	}

	if Config.PidFile != "" {
		pidfile.SetPidfilePath(Config.PidFile)
		err := pidfile.Write()
		if err != nil {
			logger.Fatal("error during pidfile.Write()",
				zap.Error(err),
			)
		}
	}

	r := http.DefaultServeMux
	r.HandleFunc("/render/", renderHandler)
	r.HandleFunc("/render", renderHandler)

	r.HandleFunc("/metrics/find/", findHandler)
	r.HandleFunc("/metrics/find", findHandler)

	r.HandleFunc("/info/", passthroughHandler)
	r.HandleFunc("/info", passthroughHandler)

	r.HandleFunc("/lb_check", lbcheckHandler)
	r.HandleFunc("/", usageHandler)

	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	err = gracehttp.Serve(&http.Server{
		Addr:    Config.Listen,
		Handler: handler,
	})

	if err != nil {
		logger.Fatal("gracehttp failed",
			zap.Error(err),
		)
	}
}
