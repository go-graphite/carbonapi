package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/intervalset"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	pickle "github.com/lomik/og-rek"
	"gopkg.in/yaml.v2"
)

type responseFormat int

func (r responseFormat) String() string {
	switch r {
	case jsonFormat:
		return "json"
	case pickleFormat:
		return "pickle"
	case protoV2Format:
		return "carbonapi_v2_pb"
	default:
		return "unknown"
	}
}

const (
	jsonFormat responseFormat = iota
	pickleFormat
	protoV2Format
	protoV3Format
)

type Metric struct {
	MetricName string    `yaml:"metricName"`
	Values     []float64 `yaml:"values"`
}

type metricForJson struct {
	MetricName string
	Values     []string
}

func (m *Metric) MarshalJSON() ([]byte, error) {
	m2 := metricForJson{
		MetricName: m.MetricName,
		Values:     make([]string, len(m.Values)),
	}

	for i, v := range m.Values {
		m2.Values[i] = fmt.Sprintf("%v", v)
	}

	return json.Marshal(m2)
}

type Response struct {
	PathExpression string   `yaml:"pathExpression"`
	Data           []Metric `yaml:"data"`
}

type MultiListenerConfig struct {
	Listeners []Config `yaml:"listeners"`
}

type Config struct {
	Address        string              `yaml:"address"`
	Code           int                 `yaml:"httpCode"`
	ShuffleResults bool                `yaml:"shuffleResults"`
	EmptyBody      bool                `yaml:"emptyBody"`
	Expressions    map[string]Response `yaml:"expressions"`
}

func copyResponse(src Response) Response {
	dst := Response{
		PathExpression: src.PathExpression,
		Data:           make([]Metric, len(src.Data)),
	}

	for i := range src.Data {
		dst.Data[i] = Metric{
			MetricName: src.Data[i].MetricName,
			Values:     make([]float64, len(src.Data[i].Values)),
		}

		for j := range src.Data[i].Values {
			dst.Data[i].Values[j] = src.Data[i].Values[j]
		}
	}

	return dst
}

func copyMap(src map[string]Response) map[string]Response {
	dst := make(map[string]Response)

	for k, v := range src {
		dst[k] = copyResponse(v)
	}

	return dst
}

var cfg = MultiListenerConfig{}

var knownFormats = map[string]responseFormat{
	"json":            jsonFormat,
	"pickle":          pickleFormat,
	"protobuf":        protoV2Format,
	"protobuf3":       protoV2Format,
	"carbonapi_v2_pb": protoV2Format,
	"carbonapi_v3_pb": protoV3Format,
}

func getFormat(req *http.Request) (responseFormat, error) {
	format := req.FormValue("format")
	if format == "" {
		format = "json"
	}

	formatCode, ok := knownFormats[format]
	if !ok {
		return formatCode, fmt.Errorf("unknown format")
	}

	return formatCode, nil
}

type listener struct {
	Config
	logger *zap.Logger
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

	query := req.Form["query"]

	if format == protoV3Format {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error("failed to read request body",
				zap.Error(err),
			)
			http.Error(wr, "Bad request (unsupported format)",
				http.StatusBadRequest)
		}

		var pv3Request protov3.MultiGlobRequest
		_ = pv3Request.Unmarshal(body)

		query = pv3Request.Metrics
	}

	logger.Info("request details",
		zap.Strings("query", query),
	)

	multiGlobs := protov3.MultiGlobResponse{
		Metrics: []protov3.GlobResponse{},
	}

	if query[0] != "*" {
		for m := range cfg.Config.Expressions {
			globMatches := []protov3.GlobMatch{}

			for _, metric := range cfg.Expressions[m].Data {
				globMatches = append(globMatches, protov3.GlobMatch{
					Path:   metric.MetricName,
					IsLeaf: true,
				})
			}
			multiGlobs.Metrics = append(multiGlobs.Metrics,
				protov3.GlobResponse{
					Name:    cfg.Expressions[m].PathExpression,
					Matches: globMatches,
				})
		}
	} else {
		returnMap := make(map[string]struct{})
		for m := range cfg.Config.Expressions {
			for _, metric := range cfg.Expressions[m].Data {
				returnMap[metric.MetricName] = struct{}{}
			}
		}

		globMatches := []protov3.GlobMatch{}
		for k := range returnMap {
			metricName := strings.Split(k, ".")

			globMatches = append(globMatches, protov3.GlobMatch{
				Path:   metricName[0],
				IsLeaf: len(metricName) == 1,
			})
		}
		multiGlobs.Metrics = append(multiGlobs.Metrics,
			protov3.GlobResponse{
				Name:    "*",
				Matches: globMatches,
			})
	}

	if cfg.Config.ShuffleResults {
		rand.Shuffle(len(multiGlobs.Metrics), func(i, j int) {
			multiGlobs.Metrics[i], multiGlobs.Metrics[j] = multiGlobs.Metrics[j], multiGlobs.Metrics[i]
		})
	}

	logger.Info("will return", zap.Any("response", multiGlobs))

	var b []byte
	switch format {
	case protoV2Format:
		response := protov2.GlobResponse{
			Name:    query[0],
			Matches: make([]protov2.GlobMatch, 0),
		}
		for _, metric := range multiGlobs.Metrics {
			if metric.Name == query[0] {
				for _, m := range metric.Matches {
					response.Matches = append(response.Matches,
						protov2.GlobMatch{
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
		pEnc := pickle.NewEncoder(p)
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
		wr.WriteHeader(http.StatusBadRequest)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}

	targets := req.Form["target"]
	logger.Info("request details",
		zap.Strings("target", targets),
	)

	multiv2 := protov2.MultiFetchResponse{
		Metrics: []protov2.FetchResponse{},
	}

	multiv3 := protov3.MultiFetchResponse{
		Metrics: []protov3.FetchResponse{},
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
			fr2 := protov2.FetchResponse{
				Name:      m.MetricName,
				StartTime: 1,
				StopTime:  int32(1 + len(protov2Values)),
				StepTime:  1,
				Values:    protov2Values,
				IsAbsent:  isAbsent,
			}

			fr3 := protov3.FetchResponse{
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
		pEnc := pickle.NewEncoder(&buf)
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

func main() {
	config := flag.String("config", "average.yaml", "yaml where it would be possible to get data")
	flag.Parse()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	if *config == "" {
		logger.Fatal("failed to get config, it should be non-null")
	}

	d, err := ioutil.ReadFile(*config)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	err = yaml.Unmarshal(d, &cfg)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
		return
	}

	wg := sync.WaitGroup{}
	for _, c := range cfg.Listeners {
		logger := logger.With(zap.String("listener", c.Address))
		listener := listener{
			Config: c,
			logger: logger,
		}

		if listener.Address == "" {
			listener.Address = ":9070"
		}

		if listener.Code == 0 {
			listener.Code = http.StatusOK
		}

		logger.Info("started",
			zap.String("listener", listener.Address),
			zap.Any("config", c),
		)

		mux := http.NewServeMux()
		mux.HandleFunc("/render", listener.renderHandler)
		mux.HandleFunc("/render/", listener.renderHandler)
		mux.HandleFunc("/metrics/find", listener.findHandler)
		mux.HandleFunc("/metrics/find/", listener.findHandler)

		wg.Add(1)
		go func() {
			err = http.ListenAndServe(listener.Address, mux)
			fmt.Println(err)
			wg.Done()
		}()
	}
	logger.Info("all listeners started")
	wg.Wait()
}
