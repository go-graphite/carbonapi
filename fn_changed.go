package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// changed(SeriesList)
func changed(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	var result []*metricData
	for _, a := range args {
		r := *a
		r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := math.NaN()
		for i, v := range a.Values {
			if math.IsNaN(prev) {
				prev = v
				r.Values[i] = 0
			} else if !math.IsNaN(v) && prev != v {
				r.Values[i] = 1
				prev = v
			} else {
				r.Values[i] = 0
			}
		}
		result = append(result, &r)
	}
	return result
}
