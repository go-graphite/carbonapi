package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-graphite/protocol/carbonapi_v2_pb"
	"github.com/go-graphite/protocol/carbonapi_v3_pb"
	ogórek "github.com/lomik/og-rek"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
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
		body, err := io.ReadAll(req.Body)
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

	httpCode := http.StatusOK
	for _, target := range targets {
		if response, ok := cfg.Expressions[target]; ok {
			if response.ReplyDelayMS > 0 {
				delay := time.Duration(response.ReplyDelayMS) * time.Millisecond
				logger.Info("will add extra delay",
					zap.Duration("delay", delay),
				)
				time.Sleep(delay)
			}
			if response.Code > 0 && response.Code != http.StatusOK {
				httpCode = response.Code
			}
			if httpCode == http.StatusOK {
				for _, m := range response.Data {
					step := m.Step
					if step == 0 {
						step = 1
					}
					startTime := m.StartTime
					if startTime == 0 {
						startTime = step
					}
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
						StartTime: int32(startTime),
						StopTime:  int32(startTime + step*len(protov2Values)),
						StepTime:  int32(step),
						Values:    protov2Values,
						IsAbsent:  isAbsent,
					}

					fr3 := carbonapi_v3_pb.FetchResponse{
						Name:                    m.MetricName,
						PathExpression:          target,
						ConsolidationFunc:       "avg",
						StartTime:               int64(startTime),
						StopTime:                int64(startTime + step*len(m.Values)),
						StepTime:                int64(step),
						XFilesFactor:            0,
						HighPrecisionTimestamps: false,
						Values:                  m.Values,
						RequestStartTime:        1,
						RequestStopTime:         int64(startTime + step*len(m.Values)),
					}

					multiv2.Metrics = append(multiv2.Metrics, fr2)
					multiv3.Metrics = append(multiv3.Metrics, fr3)
				}
			}
		}
	}

	if httpCode == http.StatusOK {
		if len(multiv2.Metrics) == 0 {
			wr.WriteHeader(http.StatusNotFound)
			_, _ = wr.Write([]byte("Not found"))
			return
		}

		if cfg.Listener.ShuffleResults {
			rand.Shuffle(len(multiv2.Metrics), func(i, j int) {
				multiv2.Metrics[i], multiv2.Metrics[j] = multiv2.Metrics[j], multiv2.Metrics[i]
			})
			rand.Shuffle(len(multiv3.Metrics), func(i, j int) {
				multiv3.Metrics[i], multiv3.Metrics[j] = multiv3.Metrics[j], multiv3.Metrics[i]
			})
		}

		contentType, d := cfg.marshalResponse(wr, logger, format, multiv3, multiv2)
		if d == nil {
			return
		}
		wr.Header().Set("Content-Type", contentType)
		_, _ = wr.Write(d)
	} else {
		wr.WriteHeader(httpCode)
		_, _ = wr.Write([]byte(http.StatusText(httpCode)))
	}
}

func (cfg *listener) marshalResponse(wr http.ResponseWriter, logger *zap.Logger, format responseFormat, multiv3 carbonapi_v3_pb.MultiFetchResponse, multiv2 carbonapi_v2_pb.MultiFetchResponse) (string, []byte) {
	var d []byte
	var contentType string
	var err error
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
			return "", nil
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
			return "", nil
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
			return "", nil
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
			return "", nil
		}
	default:
		logger.Error("format is not supported",
			zap.Any("format", format),
		)
		return "", nil
	}
	return contentType, d
}
