package main

func filter(fn summarizeFunc, isAbove, isInclusive bool) func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
		return filterFunc(e, from, until, values, fn, isAbove, isInclusive)
	}
}

// averageAbove(seriesList, n), averageBelow(seriesList, n), currentAbove(seriesList, n), currentBelow(seriesList, n), maximumAbove(seriesList, n), maximumBelow(seriesList, n), minimumAbove(seriesList, n), minimumBelow
func filterFunc(e *expr, from, until int32, values map[metricRequest][]*metricData, compute summarizeFunc, isAbove bool, isInclusive bool) []*metricData {
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	n, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData
	for _, a := range args {
		value := compute(a.Values, a.IsAbsent)
		if isAbove {
			if isInclusive {
				if value >= n {
					results = append(results, a)
				}
			} else {
				if value > n {
					results = append(results, a)
				}
			}
		} else {
			if value <= n {
				results = append(results, a)
			}
		}
	}

	return results
}
