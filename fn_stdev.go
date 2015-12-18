package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// stdev(seriesList, points, missingThreshold=0.1)
func stdev(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	points, err := getIntArg(e, 1)
	if err != nil {
		return nil
	}

	missingThreshold, err := getFloatArgDefault(e, 2, 0.1)
	if err != nil {
		return nil
	}

	minLen := int((1 - missingThreshold) * float64(points))

	var result []*metricData

	for _, a := range arg {
		w := &Windowed{data: make([]float64, points)}

		r := *a
		r.Name = proto.String(fmt.Sprintf("stdev(%s,%d)", a.GetName(), points))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				// make sure missing values are ignored
				v = math.NaN()
			}
			w.Push(v)
			r.Values[i] = w.Stdev()
			if math.IsNaN(r.Values[i]) || (i >= minLen && w.Len() < minLen) {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		result = append(result, &r)
	}
	return result
}
