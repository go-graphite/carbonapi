package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	pickle "github.com/lomik/og-rek"
	"github.com/satori/go.uuid"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/date"
	"github.com/go-graphite/carbonapi/expr"
	"github.com/go-graphite/carbonapi/expr/functions/cairo/png"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/util"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"github.com/go-graphite/carbonzipper/intervalset"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/lomik/zapwriter"
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

	r.HandleFunc("/functions", functionsHandler)
	r.HandleFunc("/functions/", functionsHandler)

	r.HandleFunc("/", usageHandler)
	return r
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

type renderResponse struct {
	data  *types.MetricData
	error error
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
	template := r.FormValue("template")
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
				accessLogDetails.ZipperRequests++

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
				accessLogDetails.ZipperRequests++

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
				rch := make(chan renderResponse, len(glob.Matches))
				var leaves int
				for _, m := range glob.Matches {
					if !m.IsLeaf {
						continue
					}
					leaves++

					apiMetrics.RenderRequests.Add(1)
					config.limiter.enter()
					accessLogDetails.ZipperRequests++

					go func(path string, from, until int32) {
						if r, err := config.zipper.Render(ctx, path, from, until); err == nil {
							rch <- renderResponse{r[0], nil}
						} else {
							rch <- renderResponse{nil, err}
						}
						config.limiter.leave()
					}(m.Path, mfetch.From, mfetch.Until)
				}

				errors := make([]error, 0)
				for i := 0; i < leaves; i++ {
					if r := <-rch; r.error == nil {
						size += r.data.Size()
						metricMap[mfetch] = append(metricMap[mfetch], r.data)
					} else {
						errors = append(errors, r.error)
					}
				}
				if len(errors) != 0 {
					logger.Error("render error occurred while fetching data",
						zap.Any("errors", errors),
					)
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
		body = png.MarshalPNGRequest(r, results, template)
	case svgFormat:
		body = png.MarshalSVGRequest(r, results, template)
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

func infoHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	uuid := uuid.NewV4()
	// TODO: Migrate to context.WithTimeout
	// ctx, _ := context.WithTimeout(context.TODO(), config.ZipperTimeout)
	ctx := util.SetUUID(r.Context(), uuid.String())
	username, _, _ := r.BasicAuth()
	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)
	format := r.FormValue("format")

	if format == "" {
		format = jsonFormat
	}

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
		err = fmt.Errorf("not implemented yet")
	default:
		err = fmt.Errorf("unknown format %v", format)
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

func functionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement helper for specific functions
	t0 := time.Now()
	username, _, _ := r.BasicAuth()

	srcIP, srcPort := splitRemoteAddr(r.RemoteAddr)

	accessLogger := zapwriter.Logger("access")
	var accessLogDetails = carbonapipb.AccessLogDetails{
		Handler:  "functions",
		Username: username,
		Url:      r.URL.RequestURI(),
		PeerIp:   srcIP,
		PeerPort: srcPort,
		Host:     r.Host,
		Referer:  r.Referer(),
		Uri:      r.RequestURI,
	}

	logAsError := false
	defer func() {
		deferredAccessLogging(accessLogger, &accessLogDetails, t0, logAsError)
	}()

	apiMetrics.Requests.Add(1)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest)+": "+err.Error(), http.StatusBadRequest)
		accessLogDetails.HttpCode = http.StatusBadRequest
		accessLogDetails.Reason = err.Error()
		logAsError = true
		return
	}

	grouped := false
	nativeOnly := false
	groupedStr := r.FormValue("grouped")
	prettyStr := r.FormValue("pretty")
	nativeOnlyStr := r.FormValue("nativeOnly")
	var marshaler func(interface{}) ([]byte, error)

	if groupedStr == "1" {
		grouped = true
	}

	if prettyStr == "1" {
		marshaler = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "\t")
		}
	} else {
		marshaler = json.Marshal
	}

	if nativeOnlyStr == "1" {
		nativeOnly = true
	}

	path := strings.Split(r.URL.EscapedPath(), "/")
	function := ""
	if len(path) >= 3 {
		function = path[2]
	}

	var b []byte
	if !nativeOnly {
		metadata.FunctionMD.RLock()
		if function != "" {
			b, err = marshaler(metadata.FunctionMD.Descriptions[function])
		} else if grouped {
			b, err = marshaler(metadata.FunctionMD.DescriptionsGrouped)
		} else {
			b, err = marshaler(metadata.FunctionMD.Descriptions)
		}
		metadata.FunctionMD.RUnlock()
	} else {
		metadata.FunctionMD.RLock()
		if function != "" {
			if !metadata.FunctionMD.Descriptions[function].Proxied {
				b, err = marshaler(metadata.FunctionMD.Descriptions[function])
			} else {
				err = fmt.Errorf("%v is proxied to graphite-web and nativeOnly was specified", function)
			}
		} else if grouped {
			descGrouped := make(map[string]map[string]types.FunctionDescription)
			for groupName, description := range metadata.FunctionMD.DescriptionsGrouped {
				desc := make(map[string]types.FunctionDescription)
				for f, d := range description {
					if d.Proxied {
						continue
					}
					desc[f] = d
				}
				if len(desc) > 0 {
					descGrouped[groupName] = desc
				}
			}
			b, err = marshaler(descGrouped)
		} else {
			desc := make(map[string]types.FunctionDescription)
			for f, d := range metadata.FunctionMD.Descriptions {
				if d.Proxied {
					continue
				}
				desc[f] = d
			}
			b, err = marshaler(desc)
		}
		metadata.FunctionMD.RUnlock()
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

	accessLogger.Info("request served", zap.Any("data", accessLogDetails))
}

var usageMsg = []byte(`
supported requests:
	/render/?target=
	/metrics/find/?query=
	/info/?target=
	/functions/
`)

func usageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(usageMsg)
}
