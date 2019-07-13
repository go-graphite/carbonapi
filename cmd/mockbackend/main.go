package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"

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

type Response struct {
	PathExpression string   `yaml:"pathExpression"`
	Data           []Metric `yaml:"data"`
}

type Config struct {
	Code 		   int      `yaml:"httpCode"`
	EmptyBody      bool     `yaml:"emptyBody"`
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

func copy(src map[string]Response) map[string]Response {
	dst := make(map[string]Response)

	for k, v := range src {
		dst[k] = copyResponse(v)
	}

	return dst
}

var cfg = Config{}

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
		return formatCode, fmt.Errorf("Unknown format")
	}

	return formatCode, nil
}

func renderHandler(wr http.ResponseWriter, req *http.Request) {
	if cfg.Code != http.StatusOK {
		wr.WriteHeader(cfg.Code)
		return
	}

	format, err := getFormat(req)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		wr.Write([]byte(err.Error()))
		return
	}

	targets := req.Form["target"]
	log.Printf("request for target=%+v\n", targets)

	multiv2 := protov2.MultiFetchResponse{
		Metrics: []protov2.FetchResponse{},
	}

	multiv3 := protov3.MultiFetchResponse{
		Metrics: []protov3.FetchResponse{},
	}

	newCfg := Config{
		Code: cfg.Code,
		EmptyBody: cfg.EmptyBody,
		Expressions: copy(cfg.Expressions),
	}

	for _, target := range targets{
		response, ok := newCfg.Expressions[target]
		if !ok {
			wr.WriteHeader(http.StatusNotFound)
			wr.Write([]byte(err.Error()))
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
			var m map[string]interface{}

			m = make(map[string]interface{})
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
		log.Printf("request will be served. format=pickle, data=%+v\n", response)
		pEnc := pickle.NewEncoder(&buf)
		err = pEnc.Encode(response)
		d = buf.Bytes()
	case protoV2Format:
		contentType = httpHeaders.ContentTypeCarbonAPIv2PB
		if cfg.EmptyBody {
			break
		}
		log.Printf("request will be served. format=protov2, data=%+v\n", multiv2)
		d, err = multiv2.Marshal()
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			wr.Write([]byte(err.Error()))
			return
		}
	case protoV3Format:
		contentType = httpHeaders.ContentTypeCarbonAPIv3PB
		if cfg.EmptyBody {
			break
		}
		log.Printf("request will be served. format=protov3, data=%+v\n", multiv3)
		d, err = multiv3.Marshal()
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			wr.Write([]byte(err.Error()))
			return
		}
	case jsonFormat:
		contentType = "application/json"
		if cfg.EmptyBody {
			break
		}
		log.Printf("request will be served. format=json, data=%+v\n", multiv2)
		d, err = json.Marshal(multiv2)
		if err != nil {
			wr.WriteHeader(http.StatusBadGateway)
			wr.Write([]byte(err.Error()))
			return
		}
	}
	wr.Header().Set("Content-Type", contentType)
	wr.Write(d)
}

func main() {
	config := flag.String("config", "average.yaml", "yaml where it would be possible to get data")
	address := flag.String("address", ":9070", "address to bind")
	flag.Parse()

	if *config == "" {
		fmt.Printf("failed to get config, it should be non-null")
		return
	}

	d, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = yaml.Unmarshal(d, &cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

		if cfg.Code == 0 {
			cfg.Code = http.StatusOK
		}


	log.Printf("started. config=%v\n", cfg)

	http.HandleFunc("/render", renderHandler)
	http.HandleFunc("/render/", renderHandler)

	err = http.ListenAndServe(*address, nil)
	fmt.Println(err)
}
