package ewma

import (
	"context"
	"math"

	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type ewma struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &ewma{}
	functions := []string{"ewma", "exponentialWeightedMovingAverage"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// ewma(seriesList, alpha)
func (f *ewma) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	alpha, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	alphaStr := e.Arg(1).StringValue()

	e.SetTarget("ewma")

	// ugh, helper.ForEachSeriesDo does not handle arguments properly
	results := make([]*types.MetricData, len(arg))
	for i, a := range arg {
		name := "ewma(" + a.Name + "," + alphaStr + ")"

		r := a.CopyTag(name, a.Tags)
		r.Values = make([]float64, len(a.Values))

		ewma := onlinestats.NewExpWeight(alpha)

		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = math.NaN()
				continue
			}

			ewma.Push(v)
			r.Values[i] = ewma.Mean()
		}
		results[i] = r
	}
	return results, nil
}

func (f *ewma) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"exponentialWeightedMovingAverage": {
			Description: "Takes a series of values and a alpha and produces an exponential moving\naverage using algorithm described at https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average\n\nExample:\n\n.. code-block:: none\n\n  &target=exponentialWeightedMovingAverage(*.transactions.count, 0.1)",
			Function:    "exponentialWeightedMovingAverage(seriesList, alpha)",
			Group:       "Calculate",
			Module:      "graphite.render.functions.custom",
			Name:        "exponentialWeightedMovingAverage",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "alpha",
					Required: true,
					Suggestions: types.NewSuggestions(
						0.1,
						0.5,
						0.7,
					),
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"ewma": {
			Description: "Takes a series of values and a alpha and produces an exponential moving\naverage using algorithm described at https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average\n\nExample:\n\n.. code-block:: none\n\n  &target=exponentialWeightedMovingAverage(*.transactions.count, 0.1)",
			Function:    "exponentialWeightedMovingAverage(seriesList, alpha)",
			Group:       "Calculate",
			Module:      "graphite.render.functions.custom",
			Name:        "ewma",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "alpha",
					Required: true,
					Suggestions: types.NewSuggestions(
						0.1,
						0.5,
						0.7,
					),
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
