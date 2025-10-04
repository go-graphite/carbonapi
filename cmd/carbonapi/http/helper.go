package http

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type responseFormat int

// for testing
var timeNow = time.Now

const (
	jsonFormat responseFormat = iota
	treejsonFormat
	pngFormat
	csvFormat
	rawFormat
	svgFormat
	protoV2Format
	protoV3Format
	pickleFormat
	completerFormat
)

const (
	ctxHeaderUUID = "X-CTX-CarbonAPI-UUID"
)

func (r responseFormat) String() string {
	switch r {
	case jsonFormat:
		return "json"
	case pickleFormat:
		return "pickle"
	case protoV2Format:
		return "protobuf3"
	case protoV3Format:
		return "carbonapi_v3_pb"
	case treejsonFormat:
		return "treejson"
	case pngFormat:
		return "png"
	case csvFormat:
		return "csv"
	case rawFormat:
		return "raw"
	case svgFormat:
		return "svg"
	case completerFormat:
		return "completer"
	default:
		return "unknown"
	}
}

func (r responseFormat) ValidExpandFormat() bool {
	switch r {
	case treejsonFormat:
		return true
	case jsonFormat:
		return true
	default:
		return false
	}
}

func (r responseFormat) ValidFindFormat() bool {
	switch r {
	case jsonFormat:
		return true
	case pickleFormat:
		return true
	case protoV2Format:
		return true
	case protoV3Format:
		return true
	case completerFormat:
		return true
	case csvFormat:
		return true
	case rawFormat:
		return true
	case treejsonFormat:
		return true
	default:
		return false
	}
}

func (r responseFormat) ValidRenderFormat() bool {
	switch r {
	case jsonFormat:
		return true
	case pickleFormat:
		return true
	case protoV2Format:
		return true
	case protoV3Format:
		return true
	case pngFormat:
		return true
	case svgFormat:
		return true
	case csvFormat:
		return true
	case rawFormat:
		return true
	default:
		return false
	}
}

var knownFormats = map[string]responseFormat{
	"json":            jsonFormat,
	"pickle":          pickleFormat,
	"treejson":        treejsonFormat,
	"protobuf":        protoV2Format,
	"protobuf3":       protoV2Format,
	"carbonapi_v2_pb": protoV2Format,
	"carbonapi_v3_pb": protoV3Format,
	"png":             pngFormat,
	"csv":             csvFormat,
	"raw":             rawFormat,
	"svg":             svgFormat,
	"completer":       completerFormat,
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

func getFormat(r *http.Request, defaultFormat responseFormat) (responseFormat, bool, string) {
	format := r.FormValue("format")

	if format == "" && (parser.TruthyBool(r.FormValue("rawData")) || parser.TruthyBool(r.FormValue("rawdata"))) {
		return rawFormat, true, format
	}

	if format == "" {
		return defaultFormat, true, format
	}

	f, ok := knownFormats[format]
	return f, ok, format
}

func writeResponse(w http.ResponseWriter, returnCode int, b []byte, format responseFormat, jsonp, carbonapiUUID string) {
	//TODO: Simplify that switch
	w.Header().Set(ctxHeaderUUID, carbonapiUUID)
	switch format {
	case jsonFormat:
		if jsonp != "" {
			w.Header().Set("Content-Type", contentTypeJavaScript)
			w.WriteHeader(returnCode)
			_, _ = w.Write([]byte(jsonp))
			_, _ = w.Write([]byte{'('})
			_, _ = w.Write(b)
			_, _ = w.Write([]byte{')'})
		} else {
			w.Header().Set("Content-Type", contentTypeJSON)
			w.WriteHeader(returnCode)
			_, _ = w.Write(b)
		}
	case protoV2Format, protoV3Format:
		w.Header().Set("Content-Type", contentTypeProtobuf)
		w.WriteHeader(returnCode)
		_, _ = w.Write(b)
	case rawFormat:
		w.Header().Set("Content-Type", contentTypeRaw)
		w.WriteHeader(returnCode)
		_, _ = w.Write(b)
	case pickleFormat:
		w.Header().Set("Content-Type", contentTypePickle)
		w.WriteHeader(returnCode)
		_, _ = w.Write(b)
	case csvFormat:
		w.Header().Set("Content-Type", contentTypeCSV)
		_, _ = w.Write(b)
	case pngFormat:
		w.Header().Set("Content-Type", contentTypePNG)
		w.WriteHeader(returnCode)
		_, _ = w.Write(b)
	case svgFormat:
		w.Header().Set("Content-Type", contentTypeSVG)
		w.WriteHeader(returnCode)
		_, _ = w.Write(b)
	}
}

func bucketRequestTimes(req *http.Request, t time.Duration) {
	ms := t.Nanoseconds() / int64(time.Millisecond)
	ApiMetrics.RequestsH.Add(ms)

	if t > config.Config.Upstreams.SlowLogThreshold {
		logger := zapwriter.Logger("slow")
		referer := req.Header.Get("Referer")
		logger.Warn("Slow Request",
			zap.Duration("time", t),
			zap.Duration("slowLogThreshold", config.Config.Upstreams.SlowLogThreshold),
			zap.String("url", req.URL.String()),
			zap.String("referer", referer),
		)
	}
}

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

func stripKey(key string, n int) string {
	if len(key) > n+3 {
		key = key[:n/2] + "..." + key[n/2+1:]
	}
	return key
}

// stripError for strip network errors (ip and other private info)
func stripError(err string) string {
	if strings.Contains(err, "connection refused") {
		return "connection refused"
	} else if strings.Contains(err, " lookup ") {
		return "lookup error"
	} else if strings.Contains(err, "broken pipe") {
		return "broken pipe"
	} else if strings.Contains(err, " connection reset ") {
		return "connection reset"
	}
	return html.EscapeString(err)
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

func deferredAccessLogging(accessLogger *zap.Logger, accessLogDetails *carbonapipb.AccessLogDetails, t time.Time, logAsError bool) {
	accessLogDetails.Runtime = time.Since(t).Seconds()
	if logAsError {
		accessLogger.Error("request failed", zap.Any("data", *accessLogDetails))
		if config.Config.Upstreams.ExtendedStat {
			switch accessLogDetails.HTTPCode {
			case 400:
				ApiMetrics.Requests400.Add(1)
			case 403:
				ApiMetrics.Requests403.Add(1)
			case 500:
				ApiMetrics.Requests500.Add(1)
			case 503:
				ApiMetrics.Requests503.Add(1)
			default:
				if accessLogDetails.HTTPCode > 500 {
					ApiMetrics.Requests5xx.Add(1)
				} else {
					ApiMetrics.Requestsxxx.Add(1)
				}
			}
		}
	} else {
		accessLogDetails.HTTPCode = http.StatusOK
		accessLogger.Info("request served", zap.Any("data", *accessLogDetails))
		ApiMetrics.Requests200.Add(1)
		Gstatsd.Timing("stat.all.response_size", accessLogDetails.CarbonapiResponseSizeBytes, 1.0)
	}
}

// durations slice is small, so no need ordered tree or other complex structure
func timestampTruncate(ts int64, duration time.Duration, durations []config.DurationTruncate) int64 {
	tm := time.Unix(ts, 0).UTC()
	for _, d := range durations {
		if duration > d.Duration || d.Duration == 0 {
			return tm.Truncate(d.Truncate).UTC().Unix()
		}
	}
	return ts
}

func setError(w http.ResponseWriter, accessLogDetails *carbonapipb.AccessLogDetails, msg string, status int, carbonapiUUID string) {
	w.Header().Set(ctxHeaderUUID, carbonapiUUID)
	if msg == "" {
		msg = http.StatusText(status)
	}
	accessLogDetails.Reason = msg
	accessLogDetails.HTTPCode = int32(status)
	msg = html.EscapeString(stripError(msg))
	http.Error(w, msg, status)
}

func joinErrors(errMap map[string]string, sep string, status int) (msg, reason string) {
	if len(errMap) == 0 {
		msg = http.StatusText(status)
	} else {
		var buf, rBuf strings.Builder
		buf.Grow(512)
		rBuf.Grow(512)

		// map is unsorted, can produce flapping ordered output, not critical
		for k, err := range errMap {
			if buf.Len() > 0 {
				buf.WriteString(sep)
				rBuf.WriteString(sep)
			}
			buf.WriteString(html.EscapeString(stripKey(k, 128)))
			rBuf.WriteString(k)
			buf.WriteString(": ")
			rBuf.WriteString(": ")
			buf.WriteString(html.EscapeString(stripError(err)))
			rBuf.WriteString(err)
		}

		msg = buf.String()
		reason = rBuf.String()
	}
	return
}

func setErrors(w http.ResponseWriter, accessLogDetails *carbonapipb.AccessLogDetails, errMamp map[string]string, status int, carbonapiUUID string) {
	w.Header().Set(ctxHeaderUUID, carbonapiUUID)
	var msg string
	if status != http.StatusOK {
		if len(errMamp) == 0 {
			msg = http.StatusText(status)
			accessLogDetails.Reason = msg
		} else {
			msg, accessLogDetails.Reason = joinErrors(errMamp, "\n", status)
		}
	}
	accessLogDetails.HTTPCode = int32(status)
	http.Error(w, msg, status)
}

func queryLengthLimitExceeded(query []string, maxLength uint64) bool {
	if maxLength > 0 {
		var queryLengthSum uint64 = 0
		for _, q := range query {
			queryLengthSum += uint64(len(q))
		}
		if queryLengthSum > maxLength {
			return true
		}
	}
	return false
}
