package main

import "regexp"

// exclude(seriesList, pattern)
func exclude(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	pat, err := getStringArg(e, 1)
	if err != nil {
		return nil
	}

	patre, err := regexp.Compile(pat)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		if !patre.MatchString(a.GetName()) {
			results = append(results, a)
		}
	}

	return results
}
