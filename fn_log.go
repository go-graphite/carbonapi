package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// logarithm(seriesList, base=10)
func log(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	base, err := getIntArgDefault(e, 1, 10)
	if err != nil {
		return nil
	}
	baseLog := math.Log(float64(base))

	var results []*metricData

	for _, a := range arg {

		var name string
		if len(e.args) == 1 {
			name = fmt.Sprintf("logarithm(%s)", a.GetName())
		} else {
			name = fmt.Sprintf("logarithm(%s,%d)", a.GetName(), base)
		}

		r := *a
		r.Name = proto.String(name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = math.Log(v) / baseLog
		}
		results = append(results, &r)
	}
	return results
}
