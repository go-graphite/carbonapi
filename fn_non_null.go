package main

// isNonNull(seriesList), isNotNull(seriesList)
func isNonNull(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	e.target = "isNonNull"

	return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
		for i := range a.Values {
			r.IsAbsent[i] = false
			if a.IsAbsent[i] {
				r.Values[i] = 0
			} else {
				r.Values[i] = 1
			}

		}
		return r
	})
}
