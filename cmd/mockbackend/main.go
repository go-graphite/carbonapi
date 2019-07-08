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
)

var knownFormats = map[string]responseFormat{
	"json":            jsonFormat,
	"pickle":          pickleFormat,
	"protobuf":        protoV2Format,
	"protobuf3":       protoV2Format,
	"carbonapi_v2_pb": protoV2Format,
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

	multiv2 := protov2.MultiFetchResponse{
		Metrics: []protov2.FetchResponse{},
	}

	newCfg := copy(&cfg)

	for _, m := range newCfg.Data {
		isAbsent := make([]bool, 0, len(m.Values))
		for i := range m.Values {
			if math.IsNaN(m.Values[i]) {
				isAbsent = append(isAbsent, true)
				m.Values[i] = 0.0
			} else {
				isAbsent = append(isAbsent, false)
			}
		}
		fr := protov2.FetchResponse{
			Name:      m.MetricName,
			StartTime: 1,
			StopTime:  int32(1 + len(m.Values)),
			StepTime:  1,
			Values:    m.Values,
			IsAbsent:  isAbsent,
		}
		multiv2.Metrics = append(multiv2.Metrics, fr)
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

		for _, metric := range multiv2.GetMetrics() {
			var m map[string]interface{}

			m = make(map[string]interface{})
			m["start"] = metric.StartTime
			m["step"] = metric.StepTime
			m["end"] = metric.StopTime
			m["name"] = metric.Name
			m["pathExpression"] = cfg.PathExpression
			m["xFilesFactor"] = 0.5
			m["consolidationFunc"] = "avg"

			mv := make([]interface{}, len(metric.Values))
			for i, p := range metric.Values {
				if metric.IsAbsent[i] {
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

type Metric struct {
	MetricName string    `yaml:"metricName"`
	Values     []float64 `yaml:"values"`
}

type Response struct {
	Code 		   int      `yaml:"httpCode"`
	EmptyBody      bool     `yaml:"emptyBody"`
	PathExpression string   `yaml:"pathExpression"`
	Data           []Metric `yaml:"data"`
}

func copy(src *Response) *Response {
	dst := &Response{
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

var cfg = Response{}

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
