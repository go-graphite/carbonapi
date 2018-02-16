package types

import (
	"bytes"
	"errors"
	"math"
	"strconv"
	"time"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	pickle "github.com/lomik/og-rek"
)

var (
	// ErrWildcardNotAllowed is an eval error returned when a wildcard/glob argument is found where a single series is required.
	ErrWildcardNotAllowed = errors.New("found wildcard where series expected")
	// ErrTooManyArguments is an eval error returned when too many arguments are provided.
	ErrTooManyArguments = errors.New("too many arguments")
)

type MetricData struct {
	pb.FetchResponse

	GraphOptions

	ValuesPerPoint    int
	aggregatedValues  []float64
	aggregatedAbsent  []bool
	AggregateFunction func([]float64, []bool) (float64, bool)
}

func MarshalCSV(results []*MetricData) []byte {

	var b []byte

	for _, r := range results {

		step := r.StepTime
		t := r.StartTime
		for i, v := range r.Values {
			b = append(b, '"')
			b = append(b, r.Name...)
			b = append(b, '"')
			b = append(b, ',')
			b = append(b, time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")...)
			b = append(b, ',')
			if !r.IsAbsent[i] {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
			b = append(b, '\n')
			t += step
		}
	}
	return b
}

func ConsolidateJSON(maxDataPoints int, results []*MetricData) {
	var startTime int32 = -1
	var endTime int32 = -1

	for _, r := range results {
		t := r.StartTime
		if startTime == -1 || startTime > t {
			startTime = t
		}
		t = r.StopTime
		if endTime == -1 || endTime < t {
			endTime = t
		}
	}

	timeRange := endTime - startTime

	if timeRange <= 0 {
		return
	}

	for _, r := range results {
		numberOfDataPoints := math.Floor(float64(timeRange / r.StepTime))
		if numberOfDataPoints > float64(maxDataPoints) {
			valuesPerPoint := math.Ceil(numberOfDataPoints / float64(maxDataPoints))
			r.SetValuesPerPoint(int(valuesPerPoint))
		}
	}
}

func MarshalJSON(results []*MetricData) []byte {
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
		t := r.StartTime
		absent := r.AggregatedAbsent()
		for i, v := range r.AggregatedValues() {
			if innerComma {
				b = append(b, ',')
			}
			innerComma = true

			b = append(b, '[')

			if absent[i] || math.IsInf(v, 0) || math.IsNaN(v) {
				b = append(b, "null"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}

			b = append(b, ',')

			b = strconv.AppendInt(b, int64(t), 10)

			b = append(b, ']')

			t += r.AggregatedTimeStep()
		}

		b = append(b, `]}`...)
	}

	b = append(b, ']')

	return b
}

func MarshalPickle(results []*MetricData) []byte {

	var p []map[string]interface{}

	for _, r := range results {
		values := make([]interface{}, len(r.Values))
		for i, v := range r.Values {
			if r.IsAbsent[i] {
				values[i] = pickle.None{}
			} else {
				values[i] = v
			}

		}
		p = append(p, map[string]interface{}{
			"name":   r.Name,
			"start":  r.StartTime,
			"end":    r.StopTime,
			"step":   r.StepTime,
			"values": values,
		})
	}

	var buf bytes.Buffer

	penc := pickle.NewEncoder(&buf)
	penc.Encode(p)

	return buf.Bytes()
}

func MarshalProtobuf(results []*MetricData) ([]byte, error) {
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

func MarshalRaw(results []*MetricData) []byte {

	var b []byte

	for _, r := range results {

		b = append(b, r.Name...)

		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.StartTime), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.StopTime), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.StepTime), 10)
		b = append(b, '|')

		var comma bool
		for i, v := range r.Values {
			if comma {
				b = append(b, ',')
			}
			comma = true
			if r.IsAbsent[i] {
				b = append(b, "None"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
		}

		b = append(b, '\n')
	}
	return b
}

func (r *MetricData) SetValuesPerPoint(v int) {
	r.ValuesPerPoint = v
	r.aggregatedValues = nil
	r.aggregatedAbsent = nil
}

func (r *MetricData) AggregatedTimeStep() int32 {
	if r.ValuesPerPoint == 1 || r.ValuesPerPoint == 0 {
		return r.StepTime
	}

	return r.StepTime * int32(r.ValuesPerPoint)
}

func (r *MetricData) AggregatedValues() []float64 {
	if r.aggregatedValues == nil {
		r.AggregateValues()
	}
	return r.aggregatedValues
}

func (r *MetricData) AggregatedAbsent() []bool {
	if r.aggregatedAbsent == nil {
		r.AggregateValues()
	}
	return r.aggregatedAbsent
}
func (r *MetricData) AggregateValues() {
	if r.ValuesPerPoint == 1 || r.ValuesPerPoint == 0 {
		r.aggregatedValues = make([]float64, len(r.Values))
		r.aggregatedAbsent = make([]bool, len(r.Values))
		copy(r.aggregatedValues, r.Values)
		copy(r.aggregatedAbsent, r.IsAbsent)
		return
	}

	if r.AggregateFunction == nil {
		r.AggregateFunction = AggMean
	}

	n := len(r.Values)/r.ValuesPerPoint + 1
	aggV := make([]float64, 0, n)
	aggA := make([]bool, 0, n)

	v := r.Values
	absent := r.IsAbsent

	for len(v) >= r.ValuesPerPoint {
		val, abs := r.AggregateFunction(v[:r.ValuesPerPoint], absent[:r.ValuesPerPoint])
		aggV = append(aggV, val)
		aggA = append(aggA, abs)
		v = v[r.ValuesPerPoint:]
		absent = absent[r.ValuesPerPoint:]
	}

	if len(v) > 0 {
		val, abs := r.AggregateFunction(v, absent)
		aggV = append(aggV, val)
		aggA = append(aggA, abs)
	}

	r.aggregatedValues = aggV
	r.aggregatedAbsent = aggA
}

func AggMean(v []float64, absent []bool) (float64, bool) {
	var sum float64
	var n int
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] {
			sum += vv
			n++
		}
	}
	return sum / float64(n), n == 0
}

func AggMax(v []float64, absent []bool) (float64, bool) {
	var m = math.Inf(-1)
	var abs = true
	for i, vv := range v {
		if !absent[i] && !math.IsNaN(vv) {
			abs = false
			if m < vv {
				m = vv
			}
		}
	}
	return m, abs
}

func AggMin(v []float64, absent []bool) (float64, bool) {
	var m = math.Inf(1)
	var abs = true
	for i, vv := range v {
		if !absent[i] && !math.IsNaN(vv) {
			abs = false
			if m > vv {
				m = vv
			}
		}
	}
	return m, abs
}

func AggSum(v []float64, absent []bool) (float64, bool) {
	var sum float64
	var abs = true
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] {
			sum += vv
			abs = false
		}
	}
	return sum, abs
}

func AggFirst(v []float64, absent []bool) (float64, bool) {
	var m = math.Inf(-1)
	var abs = true
	if len(v) > 0 {
		m = v[0]
		abs = absent[0]
	}
	return m, abs
}

func AggLast(v []float64, absent []bool) (float64, bool) {
	var m = math.Inf(-1)
	var abs = true
	if len(v) > 0 {
		m = v[len(v)-1]
		abs = absent[len(v)-1]
	}
	return m, abs
}