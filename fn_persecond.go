package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// perSecond(seriesList, maxValue=None)
func perSecond(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
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
		r := *a
		if len(e.args) == 1 {
			r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
		} else {
			r.Name = proto.String(fmt.Sprintf("%s(%s,%g)", e.target, a.GetName(), maxValue))
		}
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
				r.Values[i] = diff / float64(a.GetStepTime())
			} else if !math.IsNaN(maxValue) && maxValue >= v {
				r.Values[i] = ((maxValue - prev) + v + 1/float64(a.GetStepTime()))
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
