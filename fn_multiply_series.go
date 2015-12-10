package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// multiplySeries(factorsSeriesList)
func multiplySeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	firstFactor, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil || len(firstFactor) != 1 {
		return nil
	}

	r := *firstFactor[0]
	r.Name = proto.String(fmt.Sprintf("multiplySeries(%s)", e.argString))

	for j := 1; j < len(e.args); j++ {
		otherFactor, err := getSeriesArg(e.args[j], from, until, values)
		if err != nil || len(otherFactor) != 1 {
			return nil
		}

		if r.GetStepTime() != otherFactor[0].GetStepTime() || len(r.Values) != len(otherFactor[0].Values) {
			return nil
		}

		for i, v := range r.Values {
			if r.IsAbsent[i] || otherFactor[0].IsAbsent[i] {
				r.IsAbsent[i] = true
				r.Values[i] = math.NaN()
				continue
			}

			r.Values[i] = v * otherFactor[0].Values[i]
		}
	}

	return []*metricData{&r}
}
