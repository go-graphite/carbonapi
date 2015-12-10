package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// ignored
func draw(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))

		switch e.target {
		case "dashed":
			r.dashed = true
		case "drawAsInfinite":
			r.drawAsInfinite = true
		case "secondYAxis":
			r.secondYAxis = true
		}

		results = append(results, &r)
	}
	return results
}
