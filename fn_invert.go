package main

// invert(seriesList)
func invert(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
		for i, v := range a.Values {
			if a.IsAbsent[i] || v == 0 {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = 1 / v
		}
		return r
	})
}
