package secondYAxis

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type secondYAxis struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &secondYAxis{}
	functions := []string{"secondYAxis"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *secondYAxis) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLink()
		r.Name = "secondYAxis(" + a.Name + ")"
		r.SecondYAxis = true
		r.Tags["secondYAxis"] = "1"

		results[i] = r
	}
	return results, nil
}

func (f *secondYAxis) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"secondYAxis": {
			Name: "secondYAxis",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Graph the series on the secondary Y axis.",
			Function:    "secondYAxis(seriesList)",
			Group:       "Graph",
		},
	}
}
