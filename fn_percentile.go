package main

// percentileOfSeries(seriesList, n, interpolate=False)
func percentileOfSeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	percent, err := getFloatArg(e, 1)
	if err != nil {
		return nil
	}

	interpolate, err := getBoolArgDefault(e, 2, false)
	if err != nil {
		return nil
	}

	return aggregateSeries(e, args, func(values []float64) float64 {
		return percentile(values, percent, interpolate)
	})
}
