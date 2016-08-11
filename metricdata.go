package main

import (
	"bytes"
	"math"
	"strconv"
	"time"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	pickle "github.com/kisielk/og-rek"
)

type metricData struct {
	pb.FetchResponse

	// extra options
	xStep          float64
	valuesPerPoint int
	color          string
	alpha          float64
	lineWidth      float64

	drawAsInfinite bool
	secondYAxis    bool
	dashed         bool // TODO (ikruglov) smth like lineType would be better
	hasAlpha       bool
	stacked        bool
	stackName      string

	aggregatedValues  []float64
	aggregateFunction func([]float64, []bool) float64
}

func marshalCSV(results []*metricData) []byte {

	var b []byte

	for _, r := range results {

		step := r.GetStepTime()
		t := r.GetStartTime()
		for i, v := range r.Values {
			b = append(b, '"')
			b = append(b, r.GetName()...)
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

func marshalJSON(results []*metricData) []byte {

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
		b = strconv.AppendQuoteToASCII(b, r.GetName())
		b = append(b, `,"datapoints":[`...)

		var innerComma bool
		t := r.GetStartTime()
		for i, v := range r.Values {
			if innerComma {
				b = append(b, ',')
			}
			innerComma = true

			b = append(b, '[')

			if r.IsAbsent[i] || math.IsInf(v, 0) {
				b = append(b, "null"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}

			b = append(b, ',')

			b = strconv.AppendInt(b, int64(t), 10)

			b = append(b, ']')

			t += r.GetStepTime()
		}

		b = append(b, `]}`...)
	}

	b = append(b, ']')

	return b
}

func marshalPickle(results []*metricData) []byte {

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
			"name":   r.GetName(),
			"start":  r.GetStartTime(),
			"end":    r.GetStopTime(),
			"step":   r.GetStepTime(),
			"values": values,
		})
	}

	var buf bytes.Buffer

	penc := pickle.NewEncoder(&buf)
	penc.Encode(p)

	return buf.Bytes()
}

func marshalProtobuf(results []*metricData) []byte {
	response := pb.MultiFetchResponse{}
	for _, metric := range results {
		response.Metrics = append(response.Metrics, &((*metric).FetchResponse))
	}
	b, err := response.Marshal()
	if err != nil {
		logger.Logf("proto.Marshal: %v", err)
	}

	return b
}

func marshalRaw(results []*metricData) []byte {

	var b []byte

	for _, r := range results {

		b = append(b, r.GetName()...)

		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStartTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStopTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStepTime()), 10)
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

func (r *metricData) AggregatedTimeStep() int32 {
	if r.valuesPerPoint == 1 || r.valuesPerPoint == 0 {
		return r.GetStepTime()
	}

	return r.GetStepTime() * int32(r.valuesPerPoint)
}

func (r *metricData) AggregatedValues() []float64 {
	if r.aggregatedValues != nil {
		return r.aggregatedValues
	}

	if r.valuesPerPoint == 1 || r.valuesPerPoint == 0 {
		v := make([]float64, len(r.Values))
		for i, vv := range r.Values {
			if r.IsAbsent[i] {
				vv = math.NaN()
			}
			v[i] = vv
		}

		r.aggregatedValues = v
		return r.aggregatedValues
	}

	if r.aggregateFunction == nil {
		r.aggregateFunction = aggMean
	}

	agg := make([]float64, 0, len(r.Values)/r.valuesPerPoint+1)

	v := r.Values
	absent := r.IsAbsent

	for len(v) >= r.valuesPerPoint {
		agg = append(agg, r.aggregateFunction(v[:r.valuesPerPoint], absent[:r.valuesPerPoint]))
		v = v[r.valuesPerPoint:]
		absent = absent[r.valuesPerPoint:]
	}

	if len(v) > 0 {
		agg = append(agg, r.aggregateFunction(v, absent))
	}

	r.aggregatedValues = agg
	return r.aggregatedValues
}

func aggMean(v []float64, absent []bool) float64 {
	var sum float64
	var n int
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] {
			sum += vv
			n++
		}
	}
	return sum / float64(n)
}

func aggMax(v []float64, absent []bool) float64 {
	m := math.Inf(-1)
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] && m < vv {
			m = vv
		}
	}
	return m
}

func aggMin(v []float64, absent []bool) float64 {
	m := math.Inf(1)
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] && m > vv {
			m = vv
		}
	}
	return m
}

func aggSum(v []float64, absent []bool) float64 {
	var sum float64
	for i, vv := range v {
		if !math.IsNaN(vv) && !absent[i] {
			sum += vv
		}
	}
	return sum
}
