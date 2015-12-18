package main

import "sort"

// sortByMaxima(seriesList), sortByMinima(seriesList), sortByTotal(seriesList)
func sortBy(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	vals := make([]float64, len(arg))

	for i, a := range arg {
		switch e.target {
		case "sortByTotal":
			vals[i] = summarizeValues("sum", a.GetValues())
		case "sortByMaxima":
			vals[i] = summarizeValues("max", a.GetValues())
		case "sortByMinima":
			vals[i] = 1 / summarizeValues("min", a.GetValues())
		}
	}

	sort.Sort(byVals{vals: vals, series: arg})

	return arg
}

// sortByName(seriesList)
func sortByName(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	sort.Sort(ByName(arg))

	return arg
}
