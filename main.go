package main

import (
	"bytes"
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

	"github.com/go-graphite/carbonapi/expr"
	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonzipper/cache"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/mstats"

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

	FindRequests:        expvar.NewInt("find_requests"),
	FindCacheHits:       expvar.NewInt("find_cache_hits"),
	FindCacheMisses:     expvar.NewInt("find_cache_misses"),
	FindCacheOverheadNS: expvar.NewInt("find_cache_overhead_ns"),
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
		dt := time.Date(yy, mm, dd, hh, min, 0, 0, Config.defaultTimeZone)
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

	var tz = Config.defaultTimeZone
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
	t = time.Date(yy, mm, dd, hour, minute, 0, 0, Config.defaultTimeZone)

	return int32(t.Unix())
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()
	logger := zapwriter.Logger("render").With(
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("username", username),
	)

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "render"),
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("username", username),
		zap.String("url", r.URL.RequestURI()),
		zap.String("peer_ip", srcIP),
		zap.String("peer_port", srcPort),
		zap.String("host", r.Host),
		zap.String("referer", r.Referer()),
	)

	size := 0
	zipperRequests := 0

	Metrics.Requests.Add(1)

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

	cacheTimeout := Config.Cache.DefaultTimeoutSec

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

	if useCache {
		tc := time.Now()
		response, err := Config.queryCache.Get(cacheKey)
		td := time.Since(tc).Nanoseconds()
		Metrics.RenderCacheOverheadNS.Add(td)

		if err == nil {
			Metrics.RequestCacheHits.Add(1)
			writeResponse(w, response, format, jsonp)
			accessLogger.Info("request served",
				zap.Bool("from_cache", true),
				zap.Duration("runtime", time.Since(t0)),
				zap.Int("http_code", http.StatusOK),
				zap.Int("carbonzipper_response_size_bytes", 0),
				zap.Int("carbonapi_response_size_bytes", len(response)),
			)
			return
		}
		Metrics.RequestCacheMisses.Add(1)
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
	errors := make(map[string]string)
	metricMap := make(map[expr.MetricRequest][]*expr.MetricData)
	fatalError := false

	var metrics []string
	var targetIdx = 0
	for targetIdx < len(targets) {
		var target = targets[targetIdx]
		targetIdx++

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
				response, err := Config.findCache.Get(m.Metric)
				td := time.Since(tc).Nanoseconds()
				Metrics.FindCacheOverheadNS.Add(td)

				if err == nil {
					err := glob.Unmarshal(response)
					haveCacheData = err == nil
				}
			}

			if haveCacheData {
				Metrics.FindCacheHits.Add(1)
			} else {
				Metrics.FindCacheMisses.Add(1)
				var err error
				Metrics.FindRequests.Add(1)
				zipperRequests++
				glob, err = Config.zipper.Find(ctx, m.Metric)
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
					Config.findCache.Set(m.Metric, b, 5*60)
					td := time.Since(tc).Nanoseconds()
					Metrics.FindCacheOverheadNS.Add(td)
				}
			}

			var sendGlobs = Config.SendGlobsAsIs && len(glob.Matches) < Config.MaxBatchSize
			accessLogger = accessLogger.With(zap.Bool("send_globs", sendGlobs))

			if sendGlobs {
				// Request is "small enough" -- send the entire thing as a render request

				Metrics.RenderRequests.Add(1)
				Config.limiter.enter()
				zipperRequests++

				r, err := Config.zipper.Render(ctx, m.Metric, mfetch.From, mfetch.Until)
				if err != nil {
					errors[target] = err.Error()
					Config.limiter.leave()
					continue
				}
				Config.limiter.leave()
				metricMap[mfetch] = r
				for i := range r {
					size += r[i].Size()
				}

			} else {
				// Request is "too large"; send render requests individually
				// TODO(dgryski): group the render requests into batches
				rch := make(chan *expr.MetricData, len(glob.Matches))
				var leaves int
				for _, m := range glob.Matches {
					if !m.IsLeaf {
						continue
					}
					leaves++

					Metrics.RenderRequests.Add(1)
					Config.limiter.enter()
					zipperRequests++

					go func(path string, from, until int32) {
						if r, err := Config.zipper.Render(ctx, path, from, until); err == nil {
							rch <- r[0]
						} else {
							logger.Error("render error",
								zap.String("target", path),
								zap.Error(err),
							)
							rch <- nil
						}
						Config.limiter.leave()
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

		var rewritten bool
		var newTargets []string
		rewritten, newTargets, err = expr.RewriteExpr(exp, from32, until32, metricMap)
		if err != nil && err != expr.ErrSeriesDoesNotExist {
			errors[target] = err.Error()
			fatalError = true
			return
		} else if rewritten {
			targets = append(targets, newTargets...)
		} else {
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
					errors[target] = err.Error()
					fatalError = true
					return
				}
				results = append(results, exprs...)
			}()
		}
	}

	accessLogger = accessLogger.With(zap.Strings("metrics", metrics))

	if len(errors) > 0 && fatalError {
		httpErrors := make([]string, 0, len(errors))
		httpErrors = append(httpErrors, "Following errors have occured:")
		for _, e := range errors {
			httpErrors = append(httpErrors, e)
		}
		http.Error(w, strings.Join(httpErrors, "\n"), http.StatusBadRequest)
		accessLogger.Error("request failed",
			zap.String("reason", "encoundered multiple errors"),
			zap.Any("errors", errors),
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
		tc := time.Now()
		Config.queryCache.Set(cacheKey, body, cacheTimeout)
		td := time.Since(tc).Nanoseconds()
		Metrics.RenderCacheOverheadNS.Add(td)
	}

	gotErrors := false
	if len(errors) > 0 {
		gotErrors = true
	}

	accessLogger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
		zap.Bool("have_non_fatal_errors", gotErrors),
		zap.Any("errors", errors),
		zap.Int("zipper_requests", zipperRequests),
		zap.Int("zipper_response_size_bytes", size),
		zap.Int("carbonapi_response_size_bytes", len(body)),
	)
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()

	format := r.FormValue("format")
	jsonp := r.FormValue("jsonp")

	query := r.FormValue("query")
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "find"),
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("username", username),
		zap.String("url", r.URL.RequestURI()),
		zap.String("peer_ip", srcIP),
		zap.String("peer_port", srcPort),
		zap.String("host", r.Host),
		zap.String("referer", r.Referer()),
	)

	if query == "" {
		http.Error(w, "missing parameter `query`", http.StatusBadRequest)
		accessLogger.Info("request failed",
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
		accessLogger.Info("request failed",
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
		accessLogger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.Int("http_code", http.StatusInternalServerError),
			zap.String("reason", err.Error()),
			zap.Duration("runtime", time.Since(t0)),
		)
		return
	}

	writeResponse(w, b, format, jsonp)

	accessLogger.Info("request served",
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

func passthroughHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), Config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	accessLogger := zapwriter.Logger("access").With(
		zap.String("username", username),
		zap.String("handler", "passtrhough"),
		zap.String("carbonapi_uuid", uuid.String()),
		zap.String("peer_ip", srcIP),
		zap.String("peer_port", srcPort),
		zap.String("host", r.Host),
		zap.String("referer", r.Referer()),
	)
	var data []byte
	var err error

	if data, err = Config.zipper.Passthrough(ctx, r.URL.RequestURI()); err != nil {
		accessLogger.Info("request failed",
			zap.String("uri", r.RequestURI),
			zap.Duration("runtime", time.Since(t0)),
			zap.Int("http_code", http.StatusBadRequest),
		)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Write(data)
	accessLogger.Info("request served",
		zap.String("uri", r.RequestURI),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
	)
}

func lbcheckHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	accessLogger := zapwriter.Logger("access")

	w.Write([]byte("Ok\n"))

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	accessLogger.Info("request served",
		zap.String("handler", "lbcheck"),
		zap.String("uri", r.RequestURI),
		zap.String("peer_ip", srcIP),
		zap.String("peer_port", srcPort),
		zap.String("host", r.Host),
		zap.Duration("runtime", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
		zap.String("referer", r.Referer()),
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
	Type              string   `yaml:"type"`
	Size              int      `yaml:"size_mb"`
	MemcachedServers  []string `yaml:"memcachedServers"`
	DefaultTimeoutSec int32    `yaml:"defaultTimeoutSec"`
}

type graphiteConfig struct {
	Host     string
	Interval time.Duration
	Prefix   string
}

var Config = struct {
	Logger          []zapwriter.Config `yaml:"logger"`
	ZipperUrl       string             `yaml:"zipper"`
	Listen          string             `yaml:"listen"`
	Concurency      int                `yaml:"concurency"`
	Cache           cacheConfig        `yaml:"cache"`
	Cpus            int                `yaml:"cpus"`
	TimezoneString  string             `yaml:"tz"`
	Graphite        graphiteConfig     `yaml:"graphite"`
	IdleConnections int                `yaml:"idleConnections"`
	PidFile         string             `yaml:"pidFile"`
	SendGlobsAsIs   bool               `yaml:"sendGlobsAsIs"`
	MaxBatchSize    int                `yaml:"maxBatchSize"`

	queryCache cache.BytesCache
	findCache  cache.BytesCache

	defaultTimeZone *time.Location

	// Zipper is API entry to carbonzipper
	zipper zipper

	// Limiter limits concurrent zipper requests
	limiter limiter
}{
	ZipperUrl:     "http://localhost:8080",
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
	logger = zapwriter.Logger("main")

	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewString("BuildVersion").Set(BuildVersion)
	expvar.Publish("Config", expvar.Func(func() interface{} { return Config }))

	Config.limiter = newLimiter(Config.Concurency)

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
		Config.queryCache = cache.NewMemcached("capi", Config.Cache.MemcachedServers...)
		// find cache is only used if SendGlobsAsIs is false.
		if !Config.SendGlobsAsIs {
			Config.findCache = cache.NewExpireCache(0)
		}

		mcache := Config.queryCache.(*cache.MemcachedCache)

		Metrics.MemcacheTimeouts = expvar.Func(func() interface{} {
			return mcache.Timeouts()
		})
		expvar.Publish("memcache_timeouts", Metrics.MemcacheTimeouts)

	case "mem":
		Config.queryCache = cache.NewExpireCache(uint64(Config.Cache.Size * 1024 * 1024))

		// find cache is only used if SendGlobsAsIs is false.
		if !Config.SendGlobsAsIs {
			Config.findCache = cache.NewExpireCache(0)
		}

		qcache := Config.queryCache.(*cache.ExpireCache)

		Metrics.CacheSize = expvar.Func(func() interface{} {
			return qcache.Size()
		})
		expvar.Publish("cache_size", Metrics.CacheSize)

		Metrics.CacheItems = expvar.Func(func() interface{} {
			return qcache.Items()
		})
		expvar.Publish("cache_items", Metrics.CacheItems)

	case "null":
		// defaults
		Config.queryCache = cache.NullCache{}
		Config.findCache = cache.NullCache{}
	default:
		logger.Error("Unknown cache type",
			zap.String("cache_type", Config.Cache.Type),
			zap.Strings("known_cache_types", []string{"null", "mem", "memcache"}),
		)
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

		Config.defaultTimeZone = time.FixedZone(fields[0], offs)
		logger.Info("using fixed timezone",
			zap.String("timezone", Config.defaultTimeZone.String()),
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
		zap.Any("config", Config),
	)

	if host != "" {
		// register our metrics with graphite
		graphite := g2g.NewGraphite(host, Config.Graphite.Interval, 10*time.Second)

		hostname, _ := os.Hostname()
		hostname = strings.Replace(hostname, ".", "_", -1)

		graphite.Register(fmt.Sprintf("%s.%s.requests", Config.Graphite.Prefix, hostname), Metrics.Requests)
		graphite.Register(fmt.Sprintf("%s.%s.request_cache_hits", Config.Graphite.Prefix, hostname), Metrics.RequestCacheHits)
		graphite.Register(fmt.Sprintf("%s.%s.request_cache_misses", Config.Graphite.Prefix, hostname), Metrics.RequestCacheMisses)
		graphite.Register(fmt.Sprintf("%s.%s.request_cache_overhead_ns", Config.Graphite.Prefix, hostname), Metrics.RenderCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.%s.find_requests", Config.Graphite.Prefix, hostname), Metrics.FindRequests)
		graphite.Register(fmt.Sprintf("%s.%s.find_cache_hits", Config.Graphite.Prefix, hostname), Metrics.FindCacheHits)
		graphite.Register(fmt.Sprintf("%s.%s.find_cache_misses", Config.Graphite.Prefix, hostname), Metrics.FindCacheMisses)
		graphite.Register(fmt.Sprintf("%s.%s.find_cache_overhead_ns", Config.Graphite.Prefix, hostname), Metrics.FindCacheOverheadNS)

		graphite.Register(fmt.Sprintf("%s.%s.render_requests", Config.Graphite.Prefix, hostname), Metrics.RenderRequests)

		if Metrics.MemcacheTimeouts != nil {
			graphite.Register(fmt.Sprintf("%s.%s.memcache_timeouts", Config.Graphite.Prefix, hostname), Metrics.MemcacheTimeouts)
		}

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
