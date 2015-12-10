package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func forEachSeriesDo(e *expr, from, until int32, values map[metricRequest][]*metricData, function seriesFunc) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		results = append(results, function(a, &r))
	}
	return results
}
