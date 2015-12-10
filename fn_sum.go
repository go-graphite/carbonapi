package main

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
)

// sumSeries(*seriesLists)
func sum(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := getSeriesArgs(e.args, from, until, values)
	if err != nil {
		return nil
	}

	e.target = "sumSeries"
	return aggregateSeries(e, args, func(values []float64) float64 {
		sum := 0.0
		for _, value := range values {
			sum += value
		}
		return sum
	})
}

//sumSeriesWithWildcards(seriesList, *position)
func sumSeriesWithWildcards(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	fields, err := getIntArgs(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData

	groups := make(map[string][]*metricData)

	for _, a := range args {
		metric := extractMetric(a.GetName())
		nodes := strings.Split(metric, ".")
		var s []string
		// Yes, this is O(n^2), but len(nodes) < 10 and len(fields) < 3
		// Iterating an int slice is faster than a map for n ~ 30
		// http://www.antoine.im/posts/someone_is_wrong_on_the_internet
		for i, n := range nodes {
			if !contains(fields, i) {
				s = append(s, n)
			}
		}

		node := strings.Join(s, ".")

		groups[node] = append(groups[node], a)
	}

	for series, args := range groups {
		r := *args[0]
		r.Name = proto.String(fmt.Sprintf("sumSeriesWithWildcards(%s)", series))
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}
				atLeastOne[i] = true
				r.Values[i] += v
			}
		}

		for i, v := range atLeastOne {
			if !v {
				r.IsAbsent[i] = true
			}
		}

		results = append(results, &r)
	}
	return results
}
