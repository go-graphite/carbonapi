package main

import "math"

// maxSeries(*seriesLists)
func maxSeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArgs(e.args, from, until, values)
	if err != nil {
		return nil
	}

	return aggregateSeries(e, args, func(values []float64) float64 {
		max := math.Inf(-1)
		for _, value := range values {
			if value > max {
				max = value
			}
		}
		return max
	})

}

// minSeries(*seriesLists)
func minSeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArgs(e.args, from, until, values)
	if err != nil {
		return nil
	}

	return aggregateSeries(e, args, func(values []float64) float64 {
		min := math.Inf(1)
		for _, value := range values {
			if value < min {
				min = value
			}
		}
		return min
	})

}
