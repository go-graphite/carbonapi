package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// divideSeries(dividendSeriesList, divisorSeriesList)
func divideSeries(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	if len(e.args) != 2 {
		return nil
	}

	numerator, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	denominator, err := getSeriesArg(e.args[1], from, until, values)
	if err != nil {
		return nil
	}

	if len(numerator) != 1 || len(denominator) != 1 {
		return nil
	}

	if numerator[0].GetStepTime() != denominator[0].GetStepTime() || len(numerator[0].Values) != len(denominator[0].Values) {
		return nil
	}

	r := *numerator[0]
	r.Name = proto.String(fmt.Sprintf("divideSeries(%s)", e.argString))
	r.Values = make([]float64, len(numerator[0].Values))
	r.IsAbsent = make([]bool, len(numerator[0].Values))

	for i, v := range numerator[0].Values {

		if numerator[0].IsAbsent[i] || denominator[0].IsAbsent[i] || denominator[0].Values[i] == 0 {
			r.IsAbsent[i] = true
			continue
		}

		r.Values[i] = v / denominator[0].Values[i]
	}
	return []*metricData{&r}
}
