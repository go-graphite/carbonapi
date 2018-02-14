package main

import (
	"bytes"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"io/ioutil"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/date"
	"github.com/go-graphite/carbonapi/expr"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/png"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonzipper/cache"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/intervalset"
	"github.com/go-graphite/carbonzipper/mstats"
	"github.com/go-graphite/carbonzipper/pathcache"
	realZipper "github.com/go-graphite/carbonzipper/zipper"
	"github.com/gorilla/handlers"
	pickle "github.com/lomik/og-rek"
	"github.com/lomik/zapwriter"
	"github.com/peterbourgon/g2g"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	jsonFormat      = "json"
	treejsonFormat  = "treejson"
	pngFormat       = "png"
	csvFormat       = "csv"
	rawFormat       = "raw"
	svgFormat       = "svg"
	protobufFormat  = "protobuf"
	protobuf3Format = "protobuf3"
	pickleFormat    = "pickle"
)

func initHandlers() *http.ServeMux {
	r := http.DefaultServeMux
	r.HandleFunc("/render/", renderHandler)
	r.HandleFunc("/render", renderHandler)

	r.HandleFunc("/metrics/find/", findHandler)
	r.HandleFunc("/metrics/find", findHandler)

	r.HandleFunc("/info/", infoHandler)
	r.HandleFunc("/info", infoHandler)

	r.HandleFunc("/lb_check", lbcheckHandler)

	r.HandleFunc("/version", versionHandler)
	r.HandleFunc("/version/", versionHandler)

	r.HandleFunc("/", usageHandler)
	return r
}

var apiMetrics = struct {
	Requests              *expvar.Int
	RenderRequests        *expvar.Int
	RequestCacheHits      *expvar.Int
	RequestCacheMisses    *expvar.Int
	RenderCacheOverheadNS *expvar.Int

	FindRequests        *expvar.Int
	FindCacheHits       *expvar.Int
	FindCacheMisses     *expvar.Int
	FindCacheOverheadNS *expvar.Int

	MemcacheTimeouts expvar.Func

	CacheSize  expvar.Func
	CacheItems expvar.Func
}{
	Requests: expvar.NewInt("requests"),
	// TODO: request_cache -> render_cache
	RenderRequests:        expvar.NewInt("render_requests"),
	RequestCacheHits:      expvar.NewInt("request_cache_hits"),
	RequestCacheMisses:    expvar.NewInt("request_cache_misses"),
	RenderCacheOverheadNS: expvar.NewInt("render_cache_overhead_ns"),

	FindRequests: expvar.NewInt("find_requests"),

	FindCacheHits:       expvar.NewInt("find_cache_hits"),
	FindCacheMisses:     expvar.NewInt("find_cache_misses"),
	FindCacheOverheadNS: expvar.NewInt("find_cache_overhead_ns"),
}

var zipperMetrics = struct {
	FindRequests *expvar.Int
	FindErrors   *expvar.Int

	SearchRequests *expvar.Int

	RenderRequests *expvar.Int
	RenderErrors   *expvar.Int

	InfoRequests *expvar.Int
	InfoErrors   *expvar.Int

	Timeouts *expvar.Int

	CacheSize        expvar.Func
	CacheItems       expvar.Func
	SearchCacheSize  expvar.Func
	SearchCacheItems expvar.Func

	CacheMisses       *expvar.Int
	CacheHits         *expvar.Int
	SearchCacheMisses *expvar.Int
	SearchCacheHits   *expvar.Int
}{
	FindRequests: expvar.NewInt("zipper_find_requests"),
	FindErrors:   expvar.NewInt("zipper_find_errors"),

	SearchRequests: expvar.NewInt("zipper_search_requests"),

	RenderRequests: expvar.NewInt("zipper_render_requests"),
	RenderErrors:   expvar.NewInt("zipper_render_errors"),

	InfoRequests: expvar.NewInt("zipper_info_requests"),
	InfoErrors:   expvar.NewInt("zipper_info_errors"),

	Timeouts: expvar.NewInt("zipper_timeouts"),

	CacheHits:         expvar.NewInt("zipper_cache_hits"),
	CacheMisses:       expvar.NewInt("zipper_cache_misses"),
	SearchCacheHits:   expvar.NewInt("zipper_search_cache_hits"),
	SearchCacheMisses: expvar.NewInt("zipper_search_cache_misses"),
}

// BuildVersion is provided to be overridden at build time. Eg. go build -ldflags -X 'main.BuildVersion=...'
var BuildVersion = "(development build)"

// for testing
var timeNow = time.Now

func splitRemoteAddr(addr string) (string, string) {
	tmp := strings.Split(addr, ":")
	if len(tmp) < 1 {
		return "unknown", "unknown"
	}
	if len(tmp) == 1 {
		return tmp[0], ""
	}

	return tmp[0], tmp[1]
}

func writeResponse(w http.ResponseWriter, b []byte, format string, jsonp string) {

	switch format {
	case jsonFormat:
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
	case protobufFormat, protobuf3Format:
		w.Header().Set("Content-Type", contentTypeProtobuf)
		w.Write(b)
	case rawFormat:
		w.Header().Set("Content-Type", contentTypeRaw)
		w.Write(b)
	case pickleFormat:
		w.Header().Set("Content-Type", contentTypePickle)
		w.Write(b)
	case csvFormat:
		w.Header().Set("Content-Type", contentTypeCSV)
		w.Write(b)
	case pngFormat:
		w.Header().Set("Content-Type", contentTypePNG)
		w.Write(b)
	case svgFormat:
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

func deferredAccessLogging(accessLogger *zap.Logger, accessLogDetails *carbonapipb.AccessLogDetails, t time.Time, logAsError bool) {
	accessLogDetails.Runtime = time.Since(t).Seconds()
	if logAsError {
		accessLogger.Error("request failed", zap.Any("data", *accessLogDetails))
	} else {
		accessLogDetails.HttpCode = http.StatusOK
		accessLogger.Info("request served", zap.Any("data", *accessLogDetails))
	}
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()

	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()

	logger := zapwriter.Logger("render")
	logger.With(
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("username", username),
	)

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:       "render",
		Username:      username,
		CarbonapiUuid: uuid.String(),
		Url:           r.URL.RequestURI(),
		PeerIp:        srcIP,
		PeerPort:      srcPort,
		Host:          r.Host,
		Referer:       r.Referer(),
		Uri:           r.RequestURI,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	size := 0
	zipperRequests := 0
	apiMetrics.Requests.Add(1)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	targets := r.Form["target"]
	from := r.FormValue("from")
	until := r.FormValue("until")
	format := r.FormValue("format")
	useCache := !parser.TruthyBool(r.FormValue("noCache"))

	var jsonp string

	if format == jsonFormat {
		// TODO(dgryski): check jsonp only has valid characters
		jsonp = r.FormValue("jsonp")
	}

	if format == "" && (parser.TruthyBool(r.FormValue("rawData")) || parser.TruthyBool(r.FormValue("rawdata"))) {
		format = rawFormat
	}

	if format == "" {
		format = pngFormat
	}

	cacheTimeout := config.Cache.DefaultTimeoutSec

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
	from32 := date.DateParamToEpoch(from, qtz, timeNow().Add(-24*time.Hour).Unix(), config.defaultTimeZone)
	until32 := date.DateParamToEpoch(until, qtz, timeNow().Unix(), config.defaultTimeZone)

	accessLogDetails.UseCache = useCache
	accessLogDetails.FromRaw = from
	accessLogDetails.From = from32
	accessLogDetails.UntilRaw = until
	accessLogDetails.Until = until32
	accessLogDetails.Tz = qtz
	accessLogDetails.CacheTimeout = cacheTimeout
	accessLogDetails.Format = format
	accessLogDetails.Targets = targets
	if useCache {
		tc := time.Now()
		response, err := config.queryCache.Get(cacheKey)
		td := time.Since(tc).Nanoseconds()
		apiMetrics.RenderCacheOverheadNS.Add(td)

		accessLogDetails.CarbonzipperResponseSizeBytes = 0
		accessLogDetails.CarbonapiResponseSizeBytes = int64(len(response))

		if err == nil {
			apiMetrics.RequestCacheHits.Add(1)
			writeResponse(w, response, format, jsonp)
			accessLogDetails.FromCache = true
			return
		}
		apiMetrics.RequestCacheMisses.Add(1)
	}

	if from32 == until32 {
		http.Error(w, "Invalid empty time range", http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "invalid empty time range"
		logAsError = true
		return
	}

	var results []*types.MetricData
	errors := make(map[string]string)
	metricMap := make(map[parser.MetricRequest][]*types.MetricData)

	var metrics []string
	var targetIdx = 0
	for targetIdx < len(targets) {
		var target = targets[targetIdx]
		targetIdx++

		exp, e, err := parser.ParseExpr(target)

		if err != nil || e != "" {
			msg := buildParseErrorString(target, e, err)
			http.Error(w, msg, http.StatusBadRequest)
			accessLogDetails.Reason = msg
			accessLogDetails.HttpCode = http.StatusBadRequest
			logAsError = true
			return
		}

		for _, m := range exp.Metrics() {
			metrics = append(metrics, m.Metric)
			mfetch := m
			mfetch.From += from32
			mfetch.Until += until32

			if _, ok := metricMap[mfetch]; ok {
				// already fetched this metric for this request
				continue
			}

			var glob pb.GlobResponse
			var haveCacheData bool

			if useCache {
				tc := time.Now()
				response, err := config.findCache.Get(m.Metric)
				td := time.Since(tc).Nanoseconds()
				apiMetrics.FindCacheOverheadNS.Add(td)

				if err == nil {
					err := glob.Unmarshal(response)
					haveCacheData = err == nil
				}
			}

			if haveCacheData {
				apiMetrics.FindCacheHits.Add(1)
			} else {
				apiMetrics.FindCacheMisses.Add(1)
				var err error
				apiMetrics.FindRequests.Add(1)
				zipperRequests++

				glob, err = config.zipper.Find(ctx, m.Metric)
				if err != nil {
					logger.Error("find error",
						zap.String("metric", m.Metric),
						zap.Error(err),
					)
					continue
				}
				b, err := glob.Marshal()
				if err == nil {
					tc := time.Now()
					config.findCache.Set(m.Metric, b, 5*60)
					td := time.Since(tc).Nanoseconds()
					apiMetrics.FindCacheOverheadNS.Add(td)
				}
			}

			var sendGlobs = config.SendGlobsAsIs && len(glob.Matches) < config.MaxBatchSize
			accessLogDetails.SendGlobs = sendGlobs

			if sendGlobs {
				// Request is "small enough" -- send the entire thing as a render request

				apiMetrics.RenderRequests.Add(1)
				config.limiter.enter()
				zipperRequests++

				r, err := config.zipper.Render(ctx, m.Metric, mfetch.From, mfetch.Until)
				if err != nil {
					errors[target] = err.Error()
					config.limiter.leave()
					continue
				}
				config.limiter.leave()
				metricMap[mfetch] = r
				for i := range r {
					size += r[i].Size()
				}

			} else {
				// Request is "too large"; send render requests individually
				// TODO(dgryski): group the render requests into batches
				rch := make(chan *types.MetricData, len(glob.Matches))
				var leaves int
				for _, m := range glob.Matches {
					if !m.IsLeaf {
						continue
					}
					leaves++

					apiMetrics.RenderRequests.Add(1)
					config.limiter.enter()
					zipperRequests++

					go func(path string, from, until int32) {
						if r, err := config.zipper.Render(ctx, path, from, until); err == nil {
							rch <- r[0]
						} else {
							logger.Error("render error",
								zap.String("target", path),
								zap.Error(err),
							)
							rch <- nil
						}
						config.limiter.leave()
					}(m.Path, mfetch.From, mfetch.Until)
				}

				for i := 0; i < leaves; i++ {
					if r := <-rch; r != nil {
						size += r.Size()
						metricMap[mfetch] = append(metricMap[mfetch], r)
					}
				}
			}

			expr.SortMetrics(metricMap[mfetch], mfetch)
		}
		accessLogDetails.Metrics = metrics

		var rewritten bool
		var newTargets []string
		rewritten, newTargets, err = expr.RewriteExpr(exp, from32, until32, metricMap)
		if err != nil && err != parser.ErrSeriesDoesNotExist {
			errors[target] = err.Error()
			accessLogDetails.Reason = err.Error()
			logAsError = true
			return
		} else if rewritten {
			targets = append(targets, newTargets...)
		} else {
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("panic during eval:",
							zap.String("cache_key", cacheKey),
							zap.Any("reason", r),
							zap.Stack("stack"),
						)
					}
				}()
				exprs, err := expr.EvalExpr(exp, from32, until32, metricMap)
				if err != nil && err != parser.ErrSeriesDoesNotExist {
					errors[target] = err.Error()
					accessLogDetails.Reason = err.Error()
					logAsError = true
					return
				}
				results = append(results, exprs...)
			}()
		}
	}

	var body []byte

	switch format {
	case jsonFormat:
		if maxDataPoints, _ := strconv.Atoi(r.FormValue("maxDataPoints")); maxDataPoints != 0 {
			types.ConsolidateJSON(maxDataPoints, results)
		}

		body = types.MarshalJSON(results)
	case protobufFormat, protobuf3Format:
		body, err = types.MarshalProtobuf(results)
		if err != nil {
			logger.Info("request failed",
				zap.Int("http_code", http.StatusInternalServerError),
				zap.String("reason", err.Error()),
				zap.Duration("runtime", time.Since(t0)),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case rawFormat:
		body = types.MarshalRaw(results)
	case csvFormat:
		body = types.MarshalCSV(results)
	case pickleFormat:
		body = types.MarshalPickle(results)
	case pngFormat:
		body = png.MarshalPNGRequest(r, results)
	case svgFormat:
		body = png.MarshalSVGRequest(r, results)
	}

	writeResponse(w, body, format, jsonp)

	if len(results) != 0 {
		tc := time.Now()
		config.queryCache.Set(cacheKey, body, cacheTimeout)
		td := time.Since(tc).Nanoseconds()
		apiMetrics.RenderCacheOverheadNS.Add(td)
	}

	gotErrors := false
	if len(errors) > 0 {
		gotErrors = true
	}
	accessLogDetails.HaveNonFatalErrors = gotErrors
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()

	format := r.FormValue("format")
	jsonp := r.FormValue("jsonp")

	query := r.FormValue("query")
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:       "find",
		Username:      username,
		CarbonapiUuid: uuid.String(),
		Url:           r.URL.RequestURI(),
		PeerIp:        srcIP,
		PeerPort:      srcPort,
		Host:          r.Host,
		Referer:       r.Referer(),
		Uri:           r.RequestURI,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	if query == "" {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "missing parameter `query`"
		logAsError = true
		return
	}

	if format == "" {
		format = treejsonFormat
	}

	globs, err := config.zipper.Find(ctx, query)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	var b []byte
	switch format {
	case treejsonFormat, jsonFormat:
		b, err = findTreejson(globs)
		format = jsonFormat
	case "completer":
		b, err = findCompleter(globs)
		format = jsonFormat
	case rawFormat:
		b, err = findList(globs)
		format = rawFormat
	case protobufFormat, protobuf3Format:
		b, err = globs.Marshal()
		format = protobufFormat
	case "", pickleFormat:
		var result []map[string]interface{}

		now := int32(time.Now().Unix() + 60)
		for _, metric := range globs.Matches {
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

		p := bytes.NewBuffer(b)
		pEnc := pickle.NewEncoder(p)
		err = pEnc.Encode(result)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	writeResponse(w, b, format, jsonp)
}

type completer struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsLeaf string `json:"is_leaf"`
}

func findCompleter(globs pb.GlobResponse) ([]byte, error) {
	var b bytes.Buffer

	var complete = make([]completer, 0)

	for _, g := range globs.Matches {
		c := completer{
			Path: g.Path,
		}

		if g.IsLeaf {
			c.IsLeaf = "1"
		} else {
			c.IsLeaf = "0"
		}

		i := strings.LastIndex(c.Path, ".")

		if i != -1 {
			c.Name = c.Path[i+1:]
		} else {
			c.Name = g.Path
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

	for _, g := range globs.Matches {

		var dot string
		// make sure non-leaves end in one dot
		if !g.IsLeaf && !strings.HasSuffix(g.Path, ".") {
			dot = "."
		}

		fmt.Fprintln(&b, g.Path+dot)
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

	basepath := globs.Name

	if i := strings.LastIndex(basepath, "."); i != -1 {
		basepath = basepath[:i+1]
	} else {
		basepath = ""
	}

	for _, g := range globs.Matches {

		name := g.Path

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

		if g.IsLeaf {
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

func infoHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	format := r.FormValue("format")

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:       "info",
		Username:      username,
		CarbonapiUuid: uuid.String(),
		Url:           r.URL.RequestURI(),
		PeerIp:        srcIP,
		PeerPort:      srcPort,
		Host:          r.Host,
		Referer:       r.Referer(),
		Format:        format,
		Uri:           r.RequestURI,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	var data map[string]pb.InfoResponse
	var err error

	query := r.FormValue("target")
	if query == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = "no target specified"
		logAsError = true
		return
	}

	if data, err = config.zipper.Info(ctx, query); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	var b []byte
	switch format {
	case jsonFormat:
		b, err = json.Marshal(data)
	case protobufFormat, protobuf3Format:
		err = fmt.Errorf("Not implemented yet")
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		accessLogDetails.HttpCode = http.StatusInternalServerError
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	w.Write(b)
	accessLogDetails.Runtime = time.Since(t0).Seconds()
	accessLogDetails.HttpCode = http.StatusOK
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	accessLogger := zapwriter.Logger("access")

	w.Write([]byte("Ok\n"))

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:  "lbcheck",
		Url:      r.URL.RequestURI(),
		PeerIp:   srcIP,
		PeerPort: srcPort,
		Host:     r.Host,
		Referer:  r.Referer(),
		Runtime:  time.Since(t0).Seconds(),
		HttpCode: http.StatusOK,
		Uri:      r.RequestURI,
	}
	accessLogger.Info("request served", zap.Any("data", accessLogDetails))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	accessLogger := zapwriter.Logger("access")

	if config.GraphiteWeb09Compatibility {
		w.Write([]byte("0.9.15\n"))
	} else {
		w.Write([]byte("1.0.0\n"))
	}

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:  "version",
		Url:      r.URL.RequestURI(),
		PeerIp:   srcIP,
		PeerPort: srcPort,
		Host:     r.Host,
		Referer:  r.Referer(),
		Runtime:  time.Since(t0).Seconds(),
		HttpCode: http.StatusOK,
		Uri:      r.RequestURI,
	}
	accessLogger.Info("request served", zap.Any("data", accessLogDetails))
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

var defaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

type cacheConfig struct {
	Type              string   `yaml:"type"`
	Size              int      `yaml:"size_mb"`
	MemcachedServers  []string `yaml:"memcachedServers"`
	DefaultTimeoutSec int32    `yaml:"defaultTimeoutSec"`
}

type graphiteConfig struct {
	Pattern  string
	Host     string
	Interval time.Duration
	Prefix   string
}

var config = struct {
	ExtrapolateExperiment      bool               `yaml:"extrapolateExperiment"`
	Logger                     []zapwriter.Config `yaml:"logger"`
	Listen                     string             `yaml:"listen"`
	Concurency                 int                `yaml:"concurency"`
	Cache                      cacheConfig        `yaml:"cache"`
	Cpus                       int                `yaml:"cpus"`
	TimezoneString             string             `yaml:"tz"`
	UnicodeRangeTables         []string           `yaml:"unicodeRangeTables"`
	Graphite                   graphiteConfig     `yaml:"graphite"`
	IdleConnections            int                `yaml:"idleConnections"`
	PidFile                    string             `yaml:"pidFile"`
	SendGlobsAsIs              bool               `yaml:"sendGlobsAsIs"`
	MaxBatchSize               int                `yaml:"maxBatchSize"`
	Zipper                     string             `yaml:"zipper"`
	Upstreams                  realZipper.Config  `yaml:"upstreams"`
	ExpireDelaySec             int32              `yaml:"expireDelaySec"`
	GraphiteWeb09Compatibility bool               `yaml:"graphite09compat"`
	IgnoreClientTimeout        bool               `yaml:"ignoreClientTimeout"`

	queryCache cache.BytesCache
	findCache  cache.BytesCache

	defaultTimeZone *time.Location

	// Zipper is API entry to carbonzipper
	zipper CarbonZipper

	// Limiter limits concurrent zipper requests
	limiter limiter
}{
	ExtrapolateExperiment: false,
	Listen:        "[::]:8081",
	Concurency:    20,
	SendGlobsAsIs: false,
	MaxBatchSize:  100,
	Cache: cacheConfig{
		Type:              "mem",
		DefaultTimeoutSec: 60,
	},
	TimezoneString: "",
	Graphite: graphiteConfig{
		Pattern:  "{prefix}.{fqdn}",
		Host:     "",
		Interval: 60 * time.Second,
		Prefix:   "carbon.api",
	},
	Cpus:            0,
	IdleConnections: 10,
	PidFile:         "",

	queryCache: cache.NullCache{},
	findCache:  cache.NullCache{},

	defaultTimeZone: time.Local,
	Logger:          []zapwriter.Config{defaultLoggerConfig},

	Upstreams: realZipper.Config{
		Timeouts: realZipper.Timeouts{
			Global:       10000 * time.Second,
			AfterStarted: 2 * time.Second,
			Connect:      200 * time.Millisecond,
		},
		KeepAliveInterval: 30 * time.Second,

		MaxIdleConnsPerHost: 100,
	},
	ExpireDelaySec:             10 * 60,
	GraphiteWeb09Compatibility: false,
}

func zipperStats(stats *realZipper.Stats) {
	zipperMetrics.Timeouts.Add(stats.Timeouts)
	zipperMetrics.FindErrors.Add(stats.FindErrors)
	zipperMetrics.RenderErrors.Add(stats.RenderErrors)
	zipperMetrics.InfoErrors.Add(stats.InfoErrors)
	zipperMetrics.SearchRequests.Add(stats.SearchRequests)
	zipperMetrics.SearchCacheHits.Add(stats.SearchCacheHits)
	zipperMetrics.SearchCacheMisses.Add(stats.SearchCacheMisses)
	zipperMetrics.CacheMisses.Add(stats.CacheMisses)
	zipperMetrics.CacheHits.Add(stats.CacheHits)
}

func setUpConfig(logger *zap.Logger, zipper CarbonZipper) {
	config.Cache.MemcachedServers = viper.GetStringSlice("cache.memcachedServers")
	if n := viper.GetString("logger.logger"); n != "" {
		config.Logger[0].Logger = n
	}
	if n := viper.GetString("logger.file"); n != "" {
		config.Logger[0].File = n
	}
	if n := viper.GetString("logger.level"); n != "" {
		config.Logger[0].Level = n
	}
	if n := viper.GetString("logger.encoding"); n != "" {
		config.Logger[0].Encoding = n
	}
	if n := viper.GetString("logger.encodingtime"); n != "" {
		config.Logger[0].EncodingTime = n
	}
	if n := viper.GetString("logger.encodingduration"); n != "" {
		config.Logger[0].EncodingDuration = n
	}
	err := zapwriter.ApplyConfig(config.Logger)
	if err != nil {
		logger.Fatal("failed to initialize logger with requested configuration",
			zap.Any("configuration", config.Logger),
			zap.Error(err),
		)
	}

	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewString("BuildVersion").Set(BuildVersion)
	expvar.Publish("config", expvar.Func(func() interface{} { return config }))

	config.limiter = newLimiter(config.Concurency)
	config.zipper = zipper

	switch config.Cache.Type {
	case "memcache":
		if len(config.Cache.MemcachedServers) == 0 {
			logger.Fatal("memcache cache requested but no memcache servers provided")
		}

		logger.Info("memcached configured",
			zap.Strings("servers", config.Cache.MemcachedServers),
		)
		config.queryCache = cache.NewMemcached("capi", config.Cache.MemcachedServers...)
		// find cache is only used if SendGlobsAsIs is false.
		if !config.SendGlobsAsIs {
			config.findCache = cache.NewExpireCache(0)
		}

		mcache := config.queryCache.(*cache.MemcachedCache)

		apiMetrics.MemcacheTimeouts = expvar.Func(func() interface{} {
			return mcache.Timeouts()
		})
		expvar.Publish("memcache_timeouts", apiMetrics.MemcacheTimeouts)

	case "mem":
		config.queryCache = cache.NewExpireCache(uint64(config.Cache.Size * 1024 * 1024))

		// find cache is only used if SendGlobsAsIs is false.
		if !config.SendGlobsAsIs {
			config.findCache = cache.NewExpireCache(0)
		}

		qcache := config.queryCache.(*cache.ExpireCache)

		apiMetrics.CacheSize = expvar.Func(func() interface{} {
			return qcache.Size()
		})
		expvar.Publish("cache_size", apiMetrics.CacheSize)

		apiMetrics.CacheItems = expvar.Func(func() interface{} {
			return qcache.Items()
		})
		expvar.Publish("cache_items", apiMetrics.CacheItems)

	case "null":
		// defaults
		config.queryCache = cache.NullCache{}
		config.findCache = cache.NullCache{}
	default:
		logger.Error("unknown cache type",
			zap.String("cache_type", config.Cache.Type),
			zap.Strings("known_cache_types", []string{"null", "mem", "memcache"}),
		)
	}

	if config.TimezoneString != "" {
		fields := strings.Split(config.TimezoneString, ",")
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

		config.defaultTimeZone = time.FixedZone(fields[0], offs)
		logger.Info("using fixed timezone",
			zap.String("timezone", config.defaultTimeZone.String()),
			zap.Int("offset", offs),
		)
	}

	if len(config.UnicodeRangeTables) != 0 {
		for _, stringRange := range config.UnicodeRangeTables {
			parser.RangeTables = append(parser.RangeTables, unicode.Scripts[stringRange])
		}
	} else {
		parser.RangeTables = append(parser.RangeTables, unicode.Latin)
	}

	if config.Cpus != 0 {
		runtime.GOMAXPROCS(config.Cpus)
	}

	var host string
	if envhost := os.Getenv("GRAPHITEHOST") + ":" + os.Getenv("GRAPHITEPORT"); envhost != ":" || config.Graphite.Host != "" {
		switch {
		case envhost != ":" && config.Graphite.Host != "":
			host = config.Graphite.Host
		case envhost != ":":
			host = envhost
		case config.Graphite.Host != "":
			host = config.Graphite.Host
		}
	}

	logger.Info("starting carbonapi",
		zap.String("build_version", BuildVersion),
		zap.Any("config", config),
	)

	if host != "" {
		// register our metrics with graphite
		graphite := g2g.NewGraphite(host, config.Graphite.Interval, 10*time.Second)

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		prefix := config.Graphite.Prefix

		pattern := config.Graphite.Pattern
		pattern = strings.Replace(pattern, "{prefix}", prefix, -1)
		pattern = strings.Replace(pattern, "{fqdn}", hostname, -1)

		graphite.Register(fmt.Sprintf("%s.requests", pattern), apiMetrics.Requests)
		graphite.Register(fmt.Sprintf("%s.request_cache_hits", pattern), apiMetrics.RequestCacheHits)
		graphite.Register(fmt.Sprintf("%s.request_cache_misses", pattern), apiMetrics.RequestCacheMisses)
		graphite.Register(fmt.Sprintf("%s.request_cache_overhead_ns", pattern), apiMetrics.RenderCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.find_requests", pattern), apiMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.find_cache_hits", pattern), apiMetrics.FindCacheHits)
		graphite.Register(fmt.Sprintf("%s.find_cache_misses", pattern), apiMetrics.FindCacheMisses)
		graphite.Register(fmt.Sprintf("%s.find_cache_overhead_ns", pattern), apiMetrics.FindCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.render_requests", pattern), apiMetrics.RenderRequests)

		if apiMetrics.MemcacheTimeouts != nil {
			graphite.Register(fmt.Sprintf("%s.memcache_timeouts", pattern), apiMetrics.MemcacheTimeouts)
		}

		if apiMetrics.CacheSize != nil {
			graphite.Register(fmt.Sprintf("%s.cache_size", pattern), apiMetrics.CacheSize)
			graphite.Register(fmt.Sprintf("%s.cache_items", pattern), apiMetrics.CacheItems)
		}

		graphite.Register(fmt.Sprintf("%s.zipper.find_requests", pattern), zipperMetrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.find_errors", pattern), zipperMetrics.FindErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.render_requests", pattern), zipperMetrics.RenderRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.render_errors", pattern), zipperMetrics.RenderErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.info_requests", pattern), zipperMetrics.InfoRequests)
		graphite.Register(fmt.Sprintf("%s.zipper.info_errors", pattern), zipperMetrics.InfoErrors)

		graphite.Register(fmt.Sprintf("%s.zipper.timeouts", pattern), zipperMetrics.Timeouts)

		graphite.Register(fmt.Sprintf("%s.zipper.cache_size", pattern), zipperMetrics.CacheSize)
		graphite.Register(fmt.Sprintf("%s.zipper.cache_items", pattern), zipperMetrics.CacheItems)

		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_size", pattern), zipperMetrics.SearchCacheSize)
		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_items", pattern), zipperMetrics.SearchCacheItems)

		graphite.Register(fmt.Sprintf("%s.zipper.cache_hits", pattern), zipperMetrics.CacheHits)
		graphite.Register(fmt.Sprintf("%s.zipper.cache_misses", pattern), zipperMetrics.CacheMisses)

		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_hits", pattern), zipperMetrics.SearchCacheHits)
		graphite.Register(fmt.Sprintf("%s.zipper.search_cache_misses", pattern), zipperMetrics.SearchCacheMisses)

		go mstats.Start(config.Graphite.Interval)

		graphite.Register(fmt.Sprintf("%s.alloc", pattern), &mstats.Alloc)
		graphite.Register(fmt.Sprintf("%s.total_alloc", pattern), &mstats.TotalAlloc)
		graphite.Register(fmt.Sprintf("%s.num_gc", pattern), &mstats.NumGC)
		graphite.Register(fmt.Sprintf("%s.pause_ns", pattern), &mstats.PauseNS)

	}

	if config.PidFile != "" {
		pidfile.SetPidfilePath(config.PidFile)
		err := pidfile.Write()
		if err != nil {
			logger.Fatal("error during pidfile.Write()",
				zap.Error(err),
			)
		}
	}

	helper.ExtrapolatePoints = config.ExtrapolateExperiment
	if config.ExtrapolateExperiment {
		logger.Warn("extraploation experiment is enabled",
			zap.String("reason", "this feature is highly experimental and untested"),
		)
	}
}

func setUpViper(logger *zap.Logger) {
	configPath := flag.String("config", "", "Path to the `config file`.")
	flag.Parse()
	if *configPath != "" {
		b, err := ioutil.ReadFile(*configPath)
		if err != nil {
			logger.Fatal("error reading config file",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}

		if strings.HasSuffix(*configPath, ".toml") {
			logger.Info("will parse config as toml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("TOML")
		} else {
			logger.Info("will parse config as yaml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("YAML")
		}
		err = viper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("listen", "localhost:8081")
	viper.SetDefault("concurency", 20)
	viper.SetDefault("cache.type", "mem")
	viper.SetDefault("cache.size_mb", 0)
	viper.SetDefault("cache.defaultTimeoutSec", 60)
	viper.SetDefault("cache.memcachedServers", []string{})
	viper.SetDefault("cpus", 0)
	viper.SetDefault("tz", "")
	viper.SetDefault("sendGlobsAsIs", false)
	viper.SetDefault("maxBatchSize", 100)
	viper.SetDefault("graphite.host", "")
	viper.SetDefault("graphite.interval", "60s")
	viper.SetDefault("graphite.prefix", "carbon.api")
	viper.SetDefault("graphite.pattern", "{prefix}.{fqdn}")
	viper.SetDefault("idleConnections", 10)
	viper.SetDefault("pidFile", "")
	viper.SetDefault("upstreams.buckets", 10)
	viper.SetDefault("upstreams.timeouts.global", "10s")
	viper.SetDefault("upstreams.timeouts.afterStarted", "2s")
	viper.SetDefault("upstreams.timeouts.connect", "200ms")
	viper.SetDefault("upstreams.concurrencyLimit", 0)
	viper.SetDefault("upstreams.keepAliveInterval", "30s")
	viper.SetDefault("upstreams.maxIdleConnsPerHost", 100)
	viper.SetDefault("upstreams.backends", []string{"http://127.0.0.1:8080"})
	viper.SetDefault("upstreams.carbonsearch.backend", "")
	viper.SetDefault("upstreams.carbonsearch.prefix", "virt.v1.*")
	viper.SetDefault("upstreams.graphite09compat", false)
	viper.SetDefault("expireDelaySec", 10)
	viper.SetDefault("logger", map[string]string{})
	viper.AutomaticEnv()

	err := viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.Error(err),
		)
	}
}

func setUpConfigUpstreams(logger *zap.Logger) {
	if config.Zipper != "" {
		logger.Warn("found legacy 'zipper' option, will use it instead of any 'upstreams' specified. This will be removed in future versions!")

		config.Upstreams.Backends = []string{config.Zipper}
	}
	if len(config.Upstreams.Backends) == 0 {
		logger.Fatal("no backends specified for upstreams!")
	}

	// Setup in-memory path cache for carbonzipper requests
	config.Upstreams.PathCache = pathcache.NewPathCache(config.ExpireDelaySec)
	config.Upstreams.SearchCache = pathcache.NewPathCache(config.ExpireDelaySec)

	zipperMetrics.CacheSize = expvar.Func(func() interface{} { return config.Upstreams.PathCache.ECSize() })
	expvar.Publish("cacheSize", zipperMetrics.CacheSize)

	zipperMetrics.CacheItems = expvar.Func(func() interface{} { return config.Upstreams.PathCache.ECItems() })
	expvar.Publish("cacheItems", zipperMetrics.CacheItems)

	zipperMetrics.SearchCacheSize = expvar.Func(func() interface{} { return config.Upstreams.SearchCache.ECSize() })
	expvar.Publish("searchCacheSize", zipperMetrics.SearchCacheSize)

	zipperMetrics.SearchCacheItems = expvar.Func(func() interface{} { return config.Upstreams.SearchCache.ECItems() })
	expvar.Publish("searchCacheItems", zipperMetrics.SearchCacheItems)
}

func main() {
	err := zapwriter.ApplyConfig([]zapwriter.Config{defaultLoggerConfig})
	if err != nil {
		log.Fatal("Failed to initialize logger with default configuration")
	}
	logger := zapwriter.Logger("main")

	setUpViper(logger)
	setUpConfigUpstreams(logger)
	zipper := newZipper(zipperStats, &config.Upstreams, config.IgnoreClientTimeout, logger.With(zap.String("handler", "zipper")))
	setUpConfig(logger, zipper)

	r := initHandlers()
	handler := handlers.CompressHandler(r)
	handler = handlers.CORS()(handler)
	handler = handlers.ProxyHeaders(handler)

	err = gracehttp.Serve(&http.Server{
		Addr:    config.Listen,
		Handler: handler,
	})

	if err != nil {
		logger.Fatal("gracehttp failed",
			zap.Error(err),
		)
	}
}
