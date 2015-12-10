package main

import (
	"fmt"
	"math"

	"github.com/gogo/protobuf/proto"
)

// asPercent(seriesList, total=None)
func asPercent(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	var getTotal func(i int) float64
	var formatName func(a *metricData) string

	if len(e.args) == 1 {
		getTotal = func(i int) float64 {
			var t float64
			var atLeastOne bool
			for _, a := range arg {
				if a.IsAbsent[i] {
					continue
				}
				atLeastOne = true
				t += a.Values[i]
			}
			if !atLeastOne {
				t = math.NaN()
			}

			return t
		}
		formatName = func(a *metricData) string {
			return fmt.Sprintf("asPercent(%s)", a.GetName())
		}
	} else if len(e.args) == 2 && e.args[1].etype == etConst {
		total, err := getFloatArg(e, 1)
		if err != nil {
			return nil
		}
		getTotal = func(i int) float64 { return total }
		formatName = func(a *metricData) string {
			return fmt.Sprintf("asPercent(%s,%g)", a.GetName(), total)
		}
	} else if len(e.args) == 2 && (e.args[1].etype == etName || e.args[1].etype == etFunc) {
		total, err := getSeriesArg(e.args[1], from, until, values)
		if err != nil || len(total) != 1 {
			return nil
		}
		getTotal = func(i int) float64 {
			if total[0].IsAbsent[i] {
				return math.NaN()
			}
			return total[0].Values[i]
		}
		var totalString string
		if e.args[1].etype == etName {
			totalString = e.args[1].target
		} else {
			totalString = fmt.Sprintf("%s(%s)", e.args[1].target, e.args[1].argString)
		}
		formatName = func(a *metricData) string {
			return fmt.Sprintf("asPercent(%s,%s)", a.GetName(), totalString)
		}
	} else {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(formatName(a))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		results = append(results, &r)
	}

	for i := range results[0].Values {

		total := getTotal(i)

		for j := range results {
			r := results[j]
			a := arg[j]

			if a.IsAbsent[i] || math.IsNaN(total) || total == 0 {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = (a.Values[i] / total) * 100
		}
	}
	return results
}
