package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// color(seriesList, theColor) ignored
func applyColor(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	color, err := getStringArg(e, 1) // get color
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
		r.color = color

		results = append(results, &r)
	}

	return results
}
