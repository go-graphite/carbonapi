package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// nPercentile(seriesList, n)
func nPercentile(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	percent, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData
	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("nPercentile(%s,%g)", a.GetName(), percent))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		var values []float64
		for i, v := range a.IsAbsent {
			if !v {
				values = append(values, a.Values[i])
			}
		}

		value := percentile(values, percent, true)
		for i := range r.Values {
			r.Values[i] = value
		}

		results = append(results, &r)
	}
	return results
}
