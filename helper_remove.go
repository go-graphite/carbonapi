package main

import "math"

type removeFunc func(float64, float64) bool

func removeByValue(a *metricData, threshold float64, condition removeFunc) metricData {
	r := *a
	r.Values = make([]float64, len(a.Values))
	r.IsAbsent = make([]bool, len(a.Values))

	for i, v := range a.Values {
		if a.IsAbsent[i] || condition(v, threshold) {
			r.Values[i] = math.NaN()
			r.IsAbsent[i] = true
			continue
		}

		r.Values[i] = v
	}

	return r
}
