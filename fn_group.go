package main

import (
	"fmt"
	"strings"
)

// group(*seriesLists)
func group(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	args, err := getSeriesArgs(e.args, from, until, values)
	if err != nil {
		return nil
	}

	return args
}

// groupByNode(seriesList, nodeNum, callback)
func groupByNode(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	field, err := getIntArg(e, 1)
	if err != nil {
		return nil
	}

	callback, err := getStringArg(e, 2)
	if err != nil {
		return nil
	}

	var results []*metricData

	groups := make(map[string][]*metricData)

	for _, a := range args {

		metric := extractMetric(a.GetName())
		nodes := strings.Split(metric, ".")
		node := nodes[field]

		groups[node] = append(groups[node], a)
	}

	for k, v := range groups {

		// create a stub context to evaluate the callback in
		nexpr, _, err := parseExpr(fmt.Sprintf("%s(%s)", callback, k))
		if err != nil {
			return nil
		}

		nvalues := map[metricRequest][]*metricData{
			metricRequest{k, from, until}: v,
		}

		r := evalExpr(nexpr, from, until, nvalues)
		if r != nil {
			results = append(results, r...)
		}
	}

	return results
}
