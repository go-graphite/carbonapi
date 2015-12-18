package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// diffSeries(*seriesLists)
func diffSeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	if len(e.args) < 2 {
		return nil
	}

	minuend, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	subtrahends, err := getSeriesArgs(e.args[1:], from, until, values)
	if err != nil {
		return nil
	}

	// FIXME: need more error checking on minuend, subtrahends here
	r := *minuend[0]
	r.Name = proto.String(fmt.Sprintf("diffSeries(%s)", e.argString))
	r.Values = make([]float64, len(minuend[0].Values))
	r.IsAbsent = make([]bool, len(minuend[0].Values))

	for i, v := range minuend[0].Values {

		if minuend[0].IsAbsent[i] {
			r.IsAbsent[i] = true
			continue
		}

		var sub float64
		for _, s := range subtrahends {
			if s.IsAbsent[i] {
				continue
			}
			sub += s.Values[i]
		}

		r.Values[i] = v - sub
	}
	return []*metricData{&r}

}
