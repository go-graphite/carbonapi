package main

import (
	"regexp"
	"strings"

	"github.com/gogo/protobuf/proto"
)

// alias(seriesList, newName)
func alias(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	alias, err := getStringArg(e, 1)
	if err != nil {
		return nil
	}

	r := *arg[0]
	r.Name = proto.String(alias)
	return []*metricData{&r}
}

// aliasByMetric(seriesList)
func aliasByMetric(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
		metric := extractMetric(a.GetName())
		part := strings.Split(metric, ".")
		r.Name = proto.String(part[len(part)-1])
		r.Values = a.Values
		r.IsAbsent = a.IsAbsent
		return r
	})
}

// aliasByNode(seriesList, *nodes)
func aliasByNode(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	fields, err := getIntArgs(e, 1)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range args {

		metric := extractMetric(a.GetName())
		nodes := strings.Split(metric, ".")

		var name []string
		for _, f := range fields {
			if f < 0 {
				f += len(nodes)
			}
			if f >= len(nodes) || f < 0 {
				continue
			}
			name = append(name, nodes[f])
		}

		r := *a
		r.Name = proto.String(strings.Join(name, "."))
		results = append(results, &r)
	}

	return results
}

// aliasSub(seriesList, search, replace)
func aliasSub(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	search, err := getStringArg(e, 1)
	if err != nil {
		return nil
	}

	replace, err := getStringArg(e, 2)
	if err != nil {
		return nil
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range args {
		metric := extractMetric(a.GetName())

		r := *a
		r.Name = proto.String(re.ReplaceAllString(metric, replace))
		results = append(results, &r)
	}

	return results
}
