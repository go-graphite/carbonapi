package alpha

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type alpha struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &alpha{}
	functions := []string{"alpha"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *alpha) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	theAlpha, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLinkTags()
		r.Alpha = theAlpha
		r.HasAlpha = true
		results[i] = r
	}

	return results, nil
}

func (f *alpha) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"alpha": {
			Name: "alpha",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "alpha",
					Required: true,
					Type:     types.Float,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Assigns the given alpha transparency setting to the series. Takes a float value between 0 and 1.",
			Function:    "alpha(seriesList, alpha)",
			Group:       "Graph",
		},
	}
}
