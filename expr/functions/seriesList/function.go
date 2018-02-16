package seriesList

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	f := &seriesList{}
	functions := []string{"divideSeriesLists", "diffSeriesLists", "multiplySeriesLists", "powSeriesLists"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type seriesList struct {
	interfaces.FunctionBase
}

func (f *seriesList) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	numerators, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	denominators, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	if len(numerators) != len(denominators) {
		return nil, fmt.Errorf("both %s arguments must have equal length", e.Target())
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
	case "powSeriesLists":
		compute = func(l, r float64) float64 { return math.Pow(l, r)}
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

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *seriesList) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"divideSeriesLists": {
			Description: "Iterates over a two lists and divides list1[0} by list2[0}, list1[1} by list2[1} and so on.\nThe lists need to be the same length",
			Function:    "divideSeriesLists(dividendSeriesList, divisorSeriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "divideSeriesLists",
			Params: []types.FunctionParam{
				{
					Name:     "dividendSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "divisorSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"diffSeriesLists":     {
			Description: "Iterates over a two lists and substracts list1[0} by list2[0}, list1[1} by list2[1} and so on.\nThe lists need to be the same length",
			Function:    "diffSeriesLists(firstSeriesList, secondSeriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions.custom",
			Name:        "diffSeriesLists",
			Params: []types.FunctionParam{
				{
					Name:     "firstSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "secondSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"multiplySeriesLists": {
			Description: "Iterates over a two lists and multiplies list1[0} by list2[0}, list1[1} by list2[1} and so on.\nThe lists need to be the same length",
			Function:    "multiplySeriesLists(sourceSeriesList, factorSeriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions.custom",
			Name:        "multiplySeriesLists",
			Params: []types.FunctionParam{
				{
					Name:     "sourceSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "factorSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"powSeriesLists": {
			Description: "Iterates over a two lists and do list1[0} in power of list2[0}, list1[1} in power of  list2[1} and so on.\nThe lists need to be the same length",
			Function:    "powSeriesLists(sourceSeriesList, factorSeriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions.custom",
			Name:        "powSeriesLists",
			Params: []types.FunctionParam{
				{
					Name:     "sourceSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "factorSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}