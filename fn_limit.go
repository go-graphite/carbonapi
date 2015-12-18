package main

// limit(seriesList, n)
func limit(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	limit, err := getIntArg(e, 1) // get limit
	if err != nil {
		return nil
	}

	if limit >= len(arg) {
		return arg
	}

	return arg[:limit]
}
