package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// scale(seriesList, factor)
func scale(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	scale, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}
	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("scale(%s,%g)", a.GetName(), scale))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = v * scale
		}
		results = append(results, &r)
	}
	return results
}

// scaleToSeconds(seriesList, seconds)
func scaleToSeconds(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	seconds, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("scaleToSeconds(%s,%d)", a.GetName(), int(seconds)))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		factor := seconds / float64(a.GetStepTime())

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = v * factor
		}
		results = append(results, &r)
	}
	return results
}
