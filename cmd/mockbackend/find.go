package main

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/go-graphite/protocol/carbonapi_v2_pb"
	"github.com/go-graphite/protocol/carbonapi_v3_pb"
	ogórek "github.com/lomik/og-rek"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/intervalset"
)

func (cfg *listener) findHandler(wr http.ResponseWriter, req *http.Request) {
	_ = req.ParseMultipartForm(16 * 1024 * 1024)
	hdrs := make(map[string][]string)

	for n, v := range req.Header {
		hdrs[n] = v
	}

	logger := cfg.logger.With(
		zap.String("function", "findHandler"),
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.Any("form", req.Form),
		zap.Any("headers", hdrs),
	)
	logger.Info("got request")

	if cfg.Code != http.StatusOK {
		wr.WriteHeader(cfg.Code)
		return
	}

	format, err := getFormat(req)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}

	var query []string

	if format == protoV3Format {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Error("failed to read request body",
				zap.Error(err),
			)
			http.Error(wr, "Bad request (unsupported format)",
				http.StatusBadRequest)
			return
		}

		var pv3Request carbonapi_v3_pb.MultiGlobRequest
		_ = pv3Request.Unmarshal(body)

		query = pv3Request.Metrics
	} else {
		query = req.Form["query"]
	}

	if len(query) == 0 {
		logger.Error("Bad request (no query)")
		http.Error(wr, "Bad request (no query)",
			http.StatusBadRequest)
		return
	}

	logger.Info("request details",
		zap.Strings("query", query),
	)

	multiGlobs := carbonapi_v3_pb.MultiGlobResponse{
		Metrics: []carbonapi_v3_pb.GlobResponse{},
	}

	if query[0] == "*" {
		returnMap := make(map[string]struct{})
		for m := range cfg.Listener.Expressions {
			response := cfg.Expressions[m]
			if response.ReplyDelayMS > 0 {
				delay := time.Duration(response.ReplyDelayMS) * time.Millisecond
				time.Sleep(delay)
			}
			for _, metric := range response.Data {
				returnMap[metric.MetricName] = struct{}{}
			}
		}

		globMatches := []carbonapi_v3_pb.GlobMatch{}
		for k := range returnMap {
			metricName := strings.Split(k, ".")

			globMatches = append(globMatches, carbonapi_v3_pb.GlobMatch{
				Path:   metricName[0],
				IsLeaf: len(metricName) == 1,
			})
		}
		multiGlobs.Metrics = append(multiGlobs.Metrics,
			carbonapi_v3_pb.GlobResponse{
				Name:    "*",
				Matches: globMatches,
			})
	} else {
		for _, m := range query {
			globMatches := []carbonapi_v3_pb.GlobMatch{}
			if response, ok := cfg.Expressions[m]; ok {
				if response.ReplyDelayMS > 0 {
					delay := time.Duration(response.ReplyDelayMS) * time.Millisecond
					time.Sleep(delay)
				}
				if response.Code != 0 && response.Code != http.StatusOK {
					// return first error
					http.Error(wr, http.StatusText(response.Code), response.Code)
					return
				}

				for _, metric := range cfg.Expressions[m].Data {
					globMatches = append(globMatches, carbonapi_v3_pb.GlobMatch{
						Path:   metric.MetricName,
						IsLeaf: true,
					})
				}
				multiGlobs.Metrics = append(multiGlobs.Metrics,
					carbonapi_v3_pb.GlobResponse{
						Name:    cfg.Expressions[m].PathExpression,
						Matches: globMatches,
					})
			}
		}
	}

	if cfg.Listener.ShuffleResults {
		rand.Shuffle(len(multiGlobs.Metrics), func(i, j int) {
			multiGlobs.Metrics[i], multiGlobs.Metrics[j] = multiGlobs.Metrics[j], multiGlobs.Metrics[i]
		})
	}

	logger.Info("will return", zap.Any("response", multiGlobs))

	var b []byte
	switch format {
	case protoV2Format:
		response := carbonapi_v2_pb.GlobResponse{
			Name:    query[0],
			Matches: make([]carbonapi_v2_pb.GlobMatch, 0),
		}
		for _, metric := range multiGlobs.Metrics {
			if metric.Name == query[0] {
				for _, m := range metric.Matches {
					response.Matches = append(response.Matches,
						carbonapi_v2_pb.GlobMatch{
							Path:   m.Path,
							IsLeaf: m.IsLeaf,
						})
				}
			}
		}
		b, err = response.Marshal()
		format = protoV2Format
	case protoV3Format:
		b, err = multiGlobs.Marshal()
		format = protoV3Format
	case pickleFormat:
		var result []map[string]interface{}
		now := int32(time.Now().Unix() + 60)
		for _, globs := range multiGlobs.Metrics {
			for _, metric := range globs.Matches {
				if strings.HasPrefix(metric.Path, "_tag") {
					continue
				}
				// Tell graphite-web that we have everything
				var mm map[string]interface{}
				// graphite-web 1.0
				interval := &intervalset.IntervalSet{Start: 0, End: now}
				mm = map[string]interface{}{
					"is_leaf":   metric.IsLeaf,
					"path":      metric.Path,
					"intervals": interval,
				}
				result = append(result, mm)
			}
		}

		p := bytes.NewBuffer(b)
		pEnc := ogórek.NewEncoder(p)
		err = merry.Wrap(pEnc.Encode(result))
		b = p.Bytes()
	}

	if err != nil {
		logger.Error("failed to marshal", zap.Error(err))
		http.Error(wr, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch format {
	case jsonFormat:
		wr.Header().Set("Content-Type", contentTypeJSON)
	case protoV3Format, protoV2Format:
		wr.Header().Set("Content-Type", contentTypeProtobuf)
	case pickleFormat:
		wr.Header().Set("Content-Type", contentTypePickle)
	}
	_, _ = wr.Write(b)
}
