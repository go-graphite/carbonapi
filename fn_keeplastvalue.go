package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// keepLastValue(seriesList, limit=inf)
func keepLastValue(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	keep, err := getIntArgDefault(e, 1, -1)
	if err != nil {
		return nil
	}
	var results []*metricData

	for _, a := range arg {
		var name string
		if len(e.args) == 1 {
			name = fmt.Sprintf("keepLastValue(%s)", a.GetName())
		} else {
			name = fmt.Sprintf("keepLastValue(%s,%d)", a.GetName(), keep)
		}

		r := *a
		r.Name = proto.String(name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := math.NaN()
		missing := 0

		for i, v := range a.Values {
			if a.IsAbsent[i] {

				if (keep < 0 || missing < keep) && !math.IsNaN(prev) {
					r.Values[i] = prev
					missing++
				} else {
					r.IsAbsent[i] = true
				}

				continue
			}
			missing = 0
			prev = v
			r.Values[i] = v
		}
		results = append(results, &r)
	}
	return results
}
