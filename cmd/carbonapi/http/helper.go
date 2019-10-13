package http

import (
	"fmt"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-graphite/carbonapi/carbonapipb"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type responseFormat int

// for testing
var timeNow = time.Now

type requestInterval struct {
	from  int64
	until int64
}

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

func writeResponse(w http.ResponseWriter, b []byte, format responseFormat, jsonp string) {

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
	case protoV2Format, protoV3Format:
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

func bucketRequestTimes(req *http.Request, t time.Duration) {
	logger := zapwriter.Logger("slow")

	ms := t.Nanoseconds() / int64(time.Millisecond)

	bucket := int(ms / 100)

	if bucket < config.Config.Buckets {
		atomic.AddInt64(&TimeBuckets[bucket], 1)
	} else {
		referer := req.Header.Get("Referer")
		// Too big? Increment overflow bucket and log
		atomic.AddInt64(&TimeBuckets[config.Config.Buckets], 1)
		logger.Warn("Slow Request",
			zap.Duration("time", t),
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
		accessLogDetails.HTTPCode = http.StatusOK
		accessLogger.Info("request served", zap.Any("data", *accessLogDetails))
	}
}
