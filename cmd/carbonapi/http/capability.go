package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

func capabilityHandler(wr http.ResponseWriter, req *http.Request) {
	t0 := time.Now()

	accessLogger := zapwriter.Logger("access").With(
		zap.String("handler", "capability"),
		zap.String("url", req.URL.RequestURI()),
		zap.String("peer", req.RemoteAddr),
	)

	format := req.FormValue("format")

	accepts := req.Header["Accept"]
	for _, accept := range accepts {
		if accept == httpHeaders.ContentTypeCarbonAPIv3PB {
			format = "carbonapi_v3_pb"
			break
		}
	}

	if formatCode, ok := knownFormats[format]; ok {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			accessLogger.Error("find failed",
				zap.Duration("runtime_seconds", time.Since(t0)),
				zap.String("reason", err.Error()),
				zap.Int("http_code", http.StatusBadRequest),
			)
			http.Error(wr, "Bad request (unsupported format)",
				http.StatusBadRequest,
			)
		}

		var pv3Request pb.CapabilityRequest
		err = pv3Request.Unmarshal(body)
		if err != nil {
			accessLogger.Error("find failed",
				zap.Duration("runtime_seconds", time.Since(t0)),
				zap.String("reason", err.Error()),
				zap.Int("http_code", http.StatusBadRequest),
			)
			http.Error(wr, "Bad request (malformed body)",
				http.StatusBadRequest,
			)
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "(unknown)"
		}
		pvResponse := pb.CapabilityResponse{
			SupportedProtocols:        []string{"carbonapi_v3_pb", "carbonapi_v2_pb", "graphite-web-pickle", "graphite-web-pickle-1.1", "carbonapi_v2_json"},
			Name:                      hostname,
			HighPrecisionTimestamps:   false,
			SupportFilteringFunctions: false,
			LikeSplittedRequests:      false,
			SupportStreaming:          false,
		}

		var data []byte
		contentType := ""
		switch formatCode {
		case jsonFormat:
			contentType = httpHeaders.ContentTypeJSON
			data, err = json.Marshal(pvResponse)
		case protoV3Format:
			contentType = httpHeaders.ContentTypeCarbonAPIv3PB
			data, err = pvResponse.Marshal()
			if err != nil {
				accessLogger.Error("capability failed",
					zap.Duration("runtime_seconds", time.Since(t0)),
					zap.String("reason", err.Error()),
					zap.Int("http_code", http.StatusBadRequest),
				)
				http.Error(wr, "Bad request (unsupported format)",
					http.StatusBadRequest,
				)
			}
		}

		wr.Header().Set("Content-Type", contentType)
		_, _ = wr.Write(data)
	} else {
		accessLogger.Error("capability failed",
			zap.Duration("runtime_seconds", time.Since(t0)),
			zap.String("reason", "supported only for protoV3 format"),
			zap.Int("http_code", http.StatusBadRequest),
		)
		http.Error(wr, "Bad request (unsupported format)",
			http.StatusBadRequest,
		)
	}

	accessLogger.Info("capability success",
		zap.Duration("runtime_seconds", time.Since(t0)),
		zap.Int("http_code", http.StatusOK),
	)
	return
}
