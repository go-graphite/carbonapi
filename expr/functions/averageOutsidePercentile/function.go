package averageOutsidePercentile

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type averageOutsidePercentile struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &averageOutsidePercentile{}
	for _, n := range []string{"averageOutsidePercentile"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// averageOutsidePercentile(seriesList, n)
func (f *averageOutsidePercentile) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

	var results []*types.MetricData

	if number < 50.0 {
		number = 100.0 - number
	}

	var lowerThreshold float64
	var higherThreshold float64

	var averages = make([]float64, len(args))
	for i, arg := range args {
		averages[i] = consolidations.AggMean(arg.Values)
	}

	if len(averages) > 0 {
		lowerThreshold = consolidations.Percentile(averages, (100.0 - number), false)
		higherThreshold = consolidations.Percentile(averages, number, false)
	}

	for i, arg := range args {
		if !(averages[i] > lowerThreshold && averages[i] < higherThreshold) {
			results = append(results, arg)
		}

	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *averageOutsidePercentile) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"averageOutsidePercentile": {
			Description: "Removes series lying inside an average percentile interval.\n\n",
			Function:    "averageOutsidePercentile(seriesList, n)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "averageSeriesWithWildcards",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Required: true,
					Name:     "n",
					Type:     types.Integer,
				},
			},
		},
	}
}
