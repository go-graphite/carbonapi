package main

// integral(seriesList)
func integral(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
		current := 0.0
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			current += v
			r.Values[i] = current
		}
		return r
	})
}
