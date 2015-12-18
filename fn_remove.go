package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// removeBelowValue(seriesLists, n)
func removeBelowValue(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	threshold, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range args {
		r := removeByValue(a, threshold, func(v float64, threshold float64) bool {
			return v < threshold
		})
		r.Name = proto.String(fmt.Sprintf("removeBelowValue(%s, %g)", a.GetName(), threshold))

		results = append(results, &r)
	}
	return results
}

// removeAboveValue(seriesLists, n)
func removeAboveValue(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	threshold, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range args {
		r := removeByValue(a, threshold, func(v float64, threshold float64) bool {
			return v > threshold
		})
		r.Name = proto.String(fmt.Sprintf("removeAboveValue(%s, %g)", a.GetName(), threshold))

		results = append(results, &r)
	}
	return results
}
