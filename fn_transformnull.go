package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// transformNull(seriesList, default=0)
func transformNull(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	defv, err := getFloatArgDefault(e, 1, 0)
	if err != nil {
		return nil
	}
	var results []*metricData

	for _, a := range arg {

		var name string
		if len(e.args) == 1 {
			name = fmt.Sprintf("transformNull(%s)", a.GetName())
		} else {
			name = fmt.Sprintf("transformNull(%s,%g)", a.GetName(), defv)
		}

		r := *a
		r.Name = proto.String(name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				v = defv
			}

			r.Values[i] = v
		}

		results = append(results, &r)
	}
	return results
}
