package types

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// Value contains timestamp/value after parsing
type Value struct {
	Timestamp float64
	Value     float64
}

// Result contains result as returned by Prometheus
type Result struct {
	Metric map[string]string `json:"metric"`
	Values []Value           `json:"values"`
}

// Data - all useful data (except status and errors) that's returned by prometheus
type Data struct {
	Result     []Result `json:"result"`
	ResultType string   `json:"resultType"`
}

// HTTPResponse - full HTTP response from prometheus
type HTTPResponse struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
	Data      Data   `json:"data"`
}

func (p *Value) UnmarshalJSON(data []byte) error {
	arr := make([]interface{}, 0)
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) != 2 {
		return fmt.Errorf("length mismatch, got %v, expected 2", len(arr))
	}

	var ok bool
	p.Timestamp, ok = arr[0].(float64)
	if !ok {
		return fmt.Errorf("type mismatch for element[0/1], expected 'float64', got '%T', str=%v", arr[0], string(data))
	}

	str, ok := arr[1].(string)
	if !ok {
		return fmt.Errorf("type mismatch for element[1/1], expected 'string', got '%T', str=%v", arr[1], string(data))
	}

	switch str {
	case "NaN":
		p.Value = math.NaN()
		return nil
	case "+Inf":
		p.Value = math.Inf(1)
		return nil
	case "-Inf":
		p.Value = math.Inf(-1)
		return nil
	default:
		p.Value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
	}

	return nil
}

type PrometheusTagResponse struct {
	Status    string   `json:"status"`
	ErrorType string   `json:"errorType"`
	Error     string   `json:"error"`
	Data      []string `json:"data"`
}

type PrometheusFindResponse struct {
	Status    string              `json:"status"`
	ErrorType string              `json:"errorType"`
	Error     string              `json:"error"`
	Data      []map[string]string `json:"data"`
}

// Tag handles prometheus-specific tags
type Tag struct {
	TagValue string
	OP       string
}
