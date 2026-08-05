package dashed

import (
	"context"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type dashed struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &dashed{}
	functions := []string{"dashed"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *dashed) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	dashLen, err := e.GetFloatArgDefault(1, 2.5)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLink()
		r.Name = "dashed(" + a.Name + ")"
		r.Dashed = dashLen
		r.Tags["dashed"] = strconv.FormatFloat(dashLen, 'g', -1, 64)

		results[i] = r
	}

	return results, nil
}

func (f *dashed) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"dashed": {
			Name: "dashed",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(5),
					Name:    "dashLength",
					Type:    types.Integer,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList, followed by a float F.\n\nDraw the selected metrics with a dotted line with segments of length F\nIf omitted, the default length of the segments is 5.0\n\nExample:\n\n.. code-block:: none\n\n  &target=dashed(server01.instance01.memory.free,2.5)",
			Function:    "dashed(seriesList, dashLength=5)",
			Group:       "Graph",
		},
	}
}
