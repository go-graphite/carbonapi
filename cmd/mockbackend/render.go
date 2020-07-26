package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"

	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/protocol/carbonapi_v2_pb"
	"github.com/go-graphite/protocol/carbonapi_v3_pb"
	ogórek "github.com/lomik/og-rek"
	"go.uber.org/zap"
)

func (cfg *listener) renderHandler(wr http.ResponseWriter, req *http.Request) {
	hdrs := make(map[string][]string)

	for n, v := range req.Header {
		hdrs[n] = v
	}

	logger := cfg.logger.With(
		zap.String("function", "renderHandler"),
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.Any("headers", hdrs),
	)

	logger.Info("got request")
	if cfg.Code != http.StatusOK {
		wr.WriteHeader(cfg.Code)
		return
	}

	format, err := getFormat(req)
	if err != nil {
		logger.Error("bad request, failed to parse format")
		wr.WriteHeader(http.StatusBadRequest)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}

	targets := req.Form["target"]
	maxDataPoints := int64(0)

	if format == protoV3Format {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error("bad request, failed to read request body",
				zap.Error(err),
			)
			http.Error(wr, "bad request (failed to read request body): "+err.Error(), http.StatusBadRequest)
			return
		}

		var pv3Request carbonapi_v3_pb.MultiFetchRequest
		err = pv3Request.Unmarshal(body)

		if err != nil {
			logger.Error("bad request, failed to unmarshal request",
				zap.Error(err),
			)
			http.Error(wr, "bad request (failed to parse format): "+err.Error(), http.StatusBadRequest)
			return
		}

		targets = make([]string, len(pv3Request.Metrics))
		for i, r := range pv3Request.Metrics {
			targets[i] = r.PathExpression
		}
		maxDataPoints = pv3Request.Metrics[0].MaxDataPoints
	}

	logger.Info("request details",
		zap.Strings("target", targets),
		zap.String("format", format.String()),
		zap.Int64("maxDataPoints", maxDataPoints),
	)

	multiv2 := carbonapi_v2_pb.MultiFetchResponse{
		Metrics: []carbonapi_v2_pb.FetchResponse{},
	}

	multiv3 := carbonapi_v3_pb.MultiFetchResponse{
		Metrics: []carbonapi_v3_pb.FetchResponse{},
	}

	newCfg := Config{
		Code:        cfg.Code,
		EmptyBody:   cfg.EmptyBody,
		Expressions: copyMap(cfg.Expressions),
	}

	for _, target := range targets {
		response, ok := newCfg.Expressions[target]
		if !ok {
			wr.WriteHeader(http.StatusNotFound)
			_, _ = wr.Write([]byte("Not found"))
			return
		}
		for _, m := range response.Data {
			isAbsent := make([]bool, 0, len(m.Values))
			protov2Values := make([]float64, 0, len(m.Values))
			for i := range m.Values {
				if math.IsNaN(m.Values[i]) {
					isAbsent = append(isAbsent, true)
					protov2Values = append(protov2Values, 0.0)
				} else {
					isAbsent = append(isAbsent, false)
					protov2Values = append(protov2Values, m.Values[i])
				}
			}
			fr2 := carbonapi_v2_pb.FetchResponse{
				Name:      m.MetricName,
				StartTime: 1,
				StopTime:  int32(1 + len(protov2Values)),
				StepTime:  1,
				Values:    protov2Values,
				IsAbsent:  isAbsent,
			}

			fr3 := carbonapi_v3_pb.FetchResponse{
				Name:                    m.MetricName,
				PathExpression:          target,
				ConsolidationFunc:       "avg",
				StartTime:               1,
				StopTime:                int64(1 + len(m.Values)),
				StepTime:                1,
				XFilesFactor:            0,
				HighPrecisionTimestamps: false,
				Values:                  m.Values,
				RequestStartTime:        1,
				RequestStopTime:         int64(1 + len(m.Values)),
			}

			multiv2.Metrics = append(multiv2.Metrics, fr2)
			multiv3.Metrics = append(multiv3.Metrics, fr3)
		}
	}

	if cfg.Config.ShuffleResults {
		rand.Shuffle(len(multiv2.Metrics), func(i, j int) {
			multiv2.Metrics[i], multiv2.Metrics[j] = multiv2.Metrics[j], multiv2.Metrics[i]
		})
		rand.Shuffle(len(multiv3.Metrics), func(i, j int) {
			multiv3.Metrics[i], multiv3.Metrics[j] = multiv3.Metrics[j], multiv3.Metrics[i]
		})
	}

	var d []byte
	var contentType string
	switch format {
	case pickleFormat:
		contentType = httpHeaders.ContentTypePickle
		if cfg.EmptyBody {
			break
		}
		var response []map[string]interface{}

		for _, metric := range multiv3.GetMetrics() {
			m := make(map[string]interface{})
			m["start"] = metric.StartTime
			m["step"] = metric.StepTime
			m["end"] = metric.StopTime
			m["name"] = metric.Name
			m["pathExpression"] = metric.PathExpression
			m["xFilesFactor"] = 0.5
			m["consolidationFunc"] = "avg"

			mv := make([]interface{}, len(metric.Values))
			for i, p := range metric.Values {
				if math.IsNaN(p) {
					mv[i] = nil
				} else {
					mv[i] = p
				}
			}

			m["values"] = mv
			log.Printf("%+v\n\n", m)
			response = append(response, m)
		}

		var buf bytes.Buffer
		logger.Info("request will be served",
			zap.String("format", "pickle"),
			zap.Any("content", response),
		)
		pEnc := ogórek.NewEncoder(&buf)
		err = pEnc.Encode(response)
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			_, _ = wr.Write([]byte(err.Error()))
			return
		}
		d = buf.Bytes()
	case protoV2Format:
		contentType = httpHeaders.ContentTypeCarbonAPIv2PB
		if cfg.EmptyBody {
			break
		}
		logger.Info("request will be served",
			zap.String("format", "protov2"),
			zap.Any("content", multiv2),
		)
		d, err = multiv2.Marshal()
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			_, _ = wr.Write([]byte(err.Error()))
			return
		}
	case protoV3Format:
		contentType = httpHeaders.ContentTypeCarbonAPIv3PB
		if cfg.EmptyBody {
			break
		}
		logger.Info("request will be served",
			zap.String("format", "protov3"),
			zap.Any("content", multiv3),
		)
		d, err = multiv3.Marshal()
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			_, _ = wr.Write([]byte(err.Error()))
			return
		}
	case jsonFormat:
		contentType = "application/json"
		if cfg.EmptyBody {
			break
		}
		logger.Info("request will be served",
			zap.String("format", "json"),
			zap.Any("content", multiv2),
		)
		d, err = json.Marshal(multiv2)
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			_, _ = wr.Write([]byte(err.Error()))
			return
		}
	default:
		logger.Error("format is not supported",
			zap.Any("format", format),
		)
	}
	wr.Header().Set("Content-Type", contentType)
	_, _ = wr.Write(d)
}
