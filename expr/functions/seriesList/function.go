package seriesList

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &function{}
	functions := []string{"divideSeriesLists", "diffSeriesLists", "multiplySeriesLists"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	numerators, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	denominators, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	if len(numerators) != len(denominators) {
		return nil, fmt.Errorf("Both %s arguments must have equal length", e.Target())
	}

	var results []*types.MetricData
	functionName := e.Target()[:len(e.Target())-len("Lists")]

	var compute func(l, r float64) float64

	switch e.Target() {
	case "divideSeriesLists":
		compute = func(l, r float64) float64 { return l / r }
	case "multiplySeriesLists":
		compute = func(l, r float64) float64 { return l * r }
	case "diffSeriesLists":
		compute = func(l, r float64) float64 { return l - r }

	}
	for i, numerator := range numerators {
		denominator := denominators[i]
		if numerator.StepTime != denominator.StepTime || len(numerator.Values) != len(denominator.Values) {
			return nil, fmt.Errorf("series %s must have the same length as %s", numerator.Name, denominator.Name)
		}
		r := *numerator
		r.Name = fmt.Sprintf("%s(%s,%s)", functionName, numerator.Name, denominator.Name)
		r.Values = make([]float64, len(numerator.Values))
		r.IsAbsent = make([]bool, len(numerator.Values))
		for i, v := range numerator.Values {
			if numerator.IsAbsent[i] || denominator.IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			}

			switch e.Target() {
			case "divideSeriesLists":
				if denominator.Values[i] == 0 {
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = compute(v, denominator.Values[i])
			default:
				r.Values[i] = compute(v, denominator.Values[i])
			}
		}
		results = append(results, &r)
	}
	return results, nil
}
