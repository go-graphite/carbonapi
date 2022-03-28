package join

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type join struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &join{}, Name: "join"},
	}
}

func (f *join) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"join": {
			Description: "Returns only those metrics from `seriesA` which are presented in `seriesB`.\n\nExample:\n\n.. code-block:: none\n\n  &target=join(some.data.series.aaa, some.other.series.bbb)",
			Function:    "join(seriesA, seriesB)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "join",
			Params: []types.FunctionParam{
				{
					Name:     "seriesA",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesB",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}

func (f *join) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesA, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	seriesB, err := helper.GetSeriesArg(ctx, e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[md.Name] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA))
	for _, md := range seriesA {
		if metricsB[md.Name] {
			results = append(results, md)
		}
	}
	return results, nil
}
