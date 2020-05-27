package types

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/tags"
	pbv2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	pickle "github.com/lomik/og-rek"
)

var (
	// ErrWildcardNotAllowed is an eval error returned when a wildcard/glob argument is found where a single series is required.
	ErrWildcardNotAllowed = errors.New("found wildcard where series expected")
	// ErrTooManyArguments is an eval error returned when too many arguments are provided.
	ErrTooManyArguments = errors.New("too many arguments")
)

// MetricData contains necessary data to represent parsed metric (ready to be send out or drawn)
type MetricData struct {
	pb.FetchResponse

	GraphOptions

	ValuesPerPoint    int
	aggregatedValues  []float64
	Tags              map[string]string
	AggregateFunction func([]float64) float64 `json:"-"`
}

// MarshalCSV marshals metric data to CSV
func MarshalCSV(results []*MetricData) []byte {

	var b []byte

	for _, r := range results {

		step := r.StepTime
		t := r.StartTime
		for _, v := range r.Values {
			b = append(b, "\""+r.Name+"\","+time.Unix(t, 0).Format("2006-01-02 15:04:05")+","...)
			if !math.IsNaN(v) {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
			b = append(b, '\n')
			t += step
		}
	}
	return b
}

// ConsolidateJSON consolidates values to maxDataPoints size
func ConsolidateJSON(maxDataPoints int, results []*MetricData) {
	if len(results) == 0 {
		return
	}
	startTime := results[0].StartTime
	endTime := results[0].StopTime
	for _, r := range results {
		t := r.StartTime
		if startTime > t {
			startTime = t
		}
		t = r.StopTime
		if endTime < t {
			endTime = t
		}
	}

	timeRange := endTime - startTime

	if timeRange <= 0 {
		return
	}

	for _, r := range results {
		numberOfDataPoints := math.Floor(float64(timeRange) / float64(r.StepTime))
		if numberOfDataPoints > float64(maxDataPoints) {
			valuesPerPoint := math.Ceil(numberOfDataPoints / float64(maxDataPoints))
			r.SetValuesPerPoint(int(valuesPerPoint))
		}
	}
}

// MarshalJSON marshals metric data to JSON
func MarshalJSON(results []*MetricData, timestampMultiplier int64, noNullPoints bool) []byte {
	var b []byte
	b = append(b, '[')

	var topComma bool
	for _, r := range results {
		if r == nil {
			continue
		}

		if topComma {
			b = append(b, ',')
		}
		topComma = true

		b = append(b, `{"target":`...)
		b = strconv.AppendQuoteToASCII(b, r.Name)
		b = append(b, `,"datapoints":[`...)

		var innerComma bool
		t := r.StartTime * timestampMultiplier
		for _, v := range r.AggregatedValues() {
			if noNullPoints && math.IsNaN(v) {
				t += r.AggregatedTimeStep() * timestampMultiplier
			} else {
				if innerComma {
					b = append(b, ',')
				}
				innerComma = true

				b = append(b, '[')

				if math.IsNaN(v) || math.IsInf(v, 1) || math.IsInf(v, -1) {
					b = append(b, "null"...)
				} else {
					b = strconv.AppendFloat(b, v, 'f', -1, 64)
				}

				b = append(b, ',')

				b = strconv.AppendInt(b, t, 10)

				b = append(b, ']')

				t += r.AggregatedTimeStep() * timestampMultiplier
			}
		}

		b = append(b, `],"tags":{`...)
		notFirstTag := false
		responseTags := make([]string, 0, len(r.Tags))
		for tag := range r.Tags {
			responseTags = append(responseTags, tag)
		}
		sort.Strings(responseTags)
		for _, tag := range responseTags {
			v := r.Tags[tag]
			if notFirstTag {
				b = append(b, ',')
			}
			b = strconv.AppendQuoteToASCII(b, tag)
			b = append(b, ':')
			b = strconv.AppendQuoteToASCII(b, v)
			notFirstTag = true
		}

		b = append(b, `}}`...)
	}

	b = append(b, ']')

	return b
}

// MarshalPickle marshals metric data to pickle format
func MarshalPickle(results []*MetricData) []byte {

	var p []map[string]interface{}

	for _, r := range results {
		values := make([]interface{}, len(r.Values))
		for i, v := range r.Values {
			if math.IsNaN(v) {
				values[i] = pickle.None{}
			} else {
				values[i] = v
			}

		}
		p = append(p, map[string]interface{}{
			"name":              r.Name,
			"pathExpression":    r.PathExpression,
			"consolidationFunc": r.ConsolidationFunc,
			"start":             r.StartTime,
			"end":               r.StopTime,
			"step":              r.StepTime,
			"xFilesFactor":      r.XFilesFactor,
			"values":            values,
		})
	}

	var buf bytes.Buffer

	penc := pickle.NewEncoder(&buf)
	_ = penc.Encode(p)

	return buf.Bytes()
}

// MarshalProtobufV3 marshals metric data to protobuf
func MarshalProtobufV2(results []*MetricData) ([]byte, error) {
	response := pbv2.MultiFetchResponse{}
	for _, metric := range results {
		fmv3 := (*metric).FetchResponse
		v := make([]float64, len(fmv3.Values))
		isAbsent := make([]bool, len(fmv3.Values))
		for i := range fmv3.Values {
			if math.IsNaN(fmv3.Values[i]) {
				v[i] = 0
				isAbsent[i] = true
			} else {
				v[i] = fmv3.Values[i]
			}
		}
		fm := pbv2.FetchResponse{
			Name:      fmv3.Name,
			StartTime: int32(fmv3.StartTime),
			StopTime:  int32(fmv3.StopTime),
			StepTime:  int32(fmv3.StepTime),
			Values:    v,
			IsAbsent:  isAbsent,
		}
		response.Metrics = append(response.Metrics, fm)
	}
	b, err := response.Marshal()
	if err != nil {
		return nil, err
	}

	return b, nil
}

// MarshalProtobufV3 marshals metric data to protobuf
func MarshalProtobufV3(results []*MetricData) ([]byte, error) {
	response := pb.MultiFetchResponse{}
	for _, metric := range results {
		response.Metrics = append(response.Metrics, (*metric).FetchResponse)
	}
	b, err := response.Marshal()
	if err != nil {
		return nil, err
	}

	return b, nil
}

// MarshalProtobuf returns same things as MarshalProtobufV3
//
// Deprecated: replaced by MarshalProtobufV3 and will be removed after February 2020
func MarshalProtobuf(results []*MetricData) ([]byte, error) {
	return MarshalProtobufV3(results)
}

// MarshalRaw marshals metric data to graphite's internal format, called 'raw'
func MarshalRaw(results []*MetricData) []byte {

	var b []byte

	for _, r := range results {

		b = append(b, r.Name...)

		b = append(b, ',')
		b = strconv.AppendInt(b, r.StartTime, 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, r.StopTime, 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, r.StepTime, 10)
		b = append(b, '|')

		var comma bool
		for _, v := range r.Values {
			if comma {
				b = append(b, ',')
			}
			comma = true
			if math.IsNaN(v) {
				b = append(b, "None"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
		}

		b = append(b, '\n')
	}
	return b
}

// SetValuesPerPoint sets value per point coefficient.
func (r *MetricData) SetValuesPerPoint(v int) {
	r.ValuesPerPoint = v
	r.aggregatedValues = nil
}

// AggregatedTimeStep aggregates time step
func (r *MetricData) AggregatedTimeStep() int64 {
	if r.ValuesPerPoint == 1 || r.ValuesPerPoint == 0 {
		return r.StepTime
	}

	return r.StepTime * int64(r.ValuesPerPoint)
}

// AggregatedValues aggregates values (with cache)
func (r *MetricData) AggregatedValues() []float64 {
	if r.aggregatedValues == nil {
		r.AggregateValues()
	}
	return r.aggregatedValues
}

// AggregateValues aggregates values
func (r *MetricData) AggregateValues() {
	if r.ValuesPerPoint == 1 || r.ValuesPerPoint == 0 {
		r.aggregatedValues = make([]float64, len(r.Values))
		copy(r.aggregatedValues, r.Values)
		return
	}

	if r.AggregateFunction == nil {
		var ok bool
		if r.AggregateFunction, ok = consolidations.ConsolidationToFunc[strings.ToLower(r.ConsolidationFunc)]; !ok {
			fmt.Printf("\nconsolidateFunc = %+v\n\nstack:\n%v\n\n", r.ConsolidationFunc, string(debug.Stack()))
		}
	}

	n := len(r.Values)/r.ValuesPerPoint + 1
	aggV := make([]float64, 0, n)

	v := r.Values

	for len(v) >= r.ValuesPerPoint {
		val := r.AggregateFunction(v[:r.ValuesPerPoint])
		aggV = append(aggV, val)
		v = v[r.ValuesPerPoint:]
	}

	if len(v) > 0 {
		val := r.AggregateFunction(v)
		aggV = append(aggV, val)
	}

	r.aggregatedValues = aggV
}

// MakeMetricData creates new metrics data with given metric timeseries
func MakeMetricData(name string, values []float64, step, start int64) *MetricData {
	return makeMetricDataWithTags(name, values, step, start, tags.ExtractTags(name))
}

// MakeMetricDataWithTags creates new metrics data with given metric Time Series (with tags)
func makeMetricDataWithTags(name string, values []float64, step, start int64, tags map[string]string) *MetricData {
	stop := start + int64(len(values))*step

	return &MetricData{FetchResponse: pb.FetchResponse{
		Name:      name,
		Values:    values,
		StartTime: start,
		StepTime:  step,
		StopTime:  stop,
	},
		Tags: tags,
	}
}
