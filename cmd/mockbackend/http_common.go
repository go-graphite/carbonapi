package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Code           int      `yaml:"code"`
	ReplyDelayMS   int      `yaml:"replyDelayMS"`
	PathExpression string   `yaml:"pathExpression"`
	Data           []Metric `yaml:"data"`
	Tags           []string `yaml:"tags"`
}

type Metric struct {
	MetricName string    `yaml:"metricName"`
	Step       int       `yaml:"step"`
	StartTime  int       `yaml:"startTime"`
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

type responseFormat int

const (
	jsonFormat responseFormat = iota
	pickleFormat
	protoV2Format
	protoV3Format
)

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

var knownFormats = map[string]responseFormat{
	"json":            jsonFormat,
	"pickle":          pickleFormat,
	"protobuf":        protoV2Format,
	"protobuf3":       protoV2Format,
	"carbonapi_v2_pb": protoV2Format,
	"carbonapi_v3_pb": protoV3Format,
}

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

func copyResponse(src Response) Response {
	dst := Response{
		PathExpression: src.PathExpression,
		ReplyDelayMS:   src.ReplyDelayMS,
		Data:           make([]Metric, len(src.Data)),
	}

	for i := range src.Data {
		dst.Data[i] = Metric{
			MetricName: src.Data[i].MetricName,
			Values:     make([]float64, len(src.Data[i].Values)),
			StartTime:  src.Data[i].StartTime,
			Step:       src.Data[i].Step,
		}

		copy(dst.Data[i].Values, src.Data[i].Values)
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
