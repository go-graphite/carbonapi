package removeBetweenPercentile

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type removeBetweenPercentile struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &removeBetweenPercentile{}
	functions := []string{"removeBetweenPercentile"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// removeBetweenPercentile(seriesLists, percent)
func (f *removeBetweenPercentile) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	number, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	var percentile float64
	var results []*types.MetricData

	if number < 50.0 {
		percentile = 100.0 - number
	} else {
		percentile = number
	}

	var lowerThresholds []float64
	var higherThresholds []float64

	for i := range args[0].Values {
		var values []float64
		for _, arg := range args {
			if !math.IsNaN(arg.Values[i]) {
				values = append(values, arg.Values[i])
			}
		}
		if len(values) > 0 {
			lowerThresholds = append(lowerThresholds, consolidations.Percentile(values, (100.0-percentile), false))
			higherThresholds = append(higherThresholds, consolidations.Percentile(values, percentile, false))
		}
	}

	numberStr := strconv.FormatFloat(number, 'f', -1, 64)
	for i, a := range args {
		r := a.CopyLink()
		r.Name = e.Target() + "(" + a.Name + ", " + numberStr + ")"

		for _, v := range a.Values {
			if !(v > lowerThresholds[i] && v < higherThresholds[i]) {
				results = append(results, r)
				break
			}
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *removeBetweenPercentile) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"removeBetweenPercentile": {
			Description: "Removes series that do not have an value lying in the x-percentile of all the values at a moment",
			Function:    "removeBetweenPercentile(seriesList, n)",
			Group:       "Filter Data",
			Module:      "graphite.render.functions",
			Name:        "removeBetweenPercentile",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
	}
}
