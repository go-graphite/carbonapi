package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// derivative(seriesList)
func derivative(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
		prev := a.Values[0]
		for i, v := range a.Values {
			if i == 0 || a.IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v - prev
			prev = v
		}
		return r
	})
}

// nonNegativeDerivative(seriesList, maxValue=None)
func nonNegativeDerivative(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	maxValue, err := getFloatArgDefault(e, 1, math.NaN())
	if err != nil {
		return nil
	}

	var result []*metricData
	for _, a := range args {
		var name string
		if len(e.args) == 1 {
			name = fmt.Sprintf("nonNegativeDerivative(%s)", a.GetName())
		} else {
			name = fmt.Sprintf("nonNegativeDerivative(%s,%g)", a.GetName(), maxValue)
		}

		r := *a
		r.Name = proto.String(name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := a.Values[0]
		for i, v := range a.Values {
			if i == 0 || a.IsAbsent[i] || a.IsAbsent[i-1] {
				r.IsAbsent[i] = true
				prev = v
				continue
			}
			diff := v - prev
			if diff >= 0 {
				r.Values[i] = diff
			} else if !math.IsNaN(maxValue) && maxValue >= v {
				r.Values[i] = ((maxValue - prev) + v + 1)
			} else {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
			prev = v
		}
		result = append(result, &r)
	}
	return result
}
