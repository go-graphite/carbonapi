package limit

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type limit struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &limit{}
	functions := []string{"limit"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// limit(seriesList, n)
func (f *limit) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	limit, err := e.GetIntArg(1) // get limit
	if err != nil {
		return nil, err
	}

	if limit >= len(arg) {
		return arg, nil
	}

	return arg[:limit], nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *limit) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"limit": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\n\nOnly draw the first N metrics.  Useful when testing a wildcard in a metric.\n\nExample:\n\n.. code-block:: none\n\n  &target=limit(server*.instance*.memory.free,5)\n\nDraws only the first 5 instance's memory free.",
			Function:    "limit(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "limit",
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
	}
}
