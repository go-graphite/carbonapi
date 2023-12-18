package integral

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type integral struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &integral{}
	functions := []string{"integral"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// integral(seriesList)
func (f *integral) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(ctx, f.GetEvaluator(), e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		current := 0.0
		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = math.NaN()
				continue
			}
			current += v
			r.Values[i] = current
		}
		r.Tags["integral"] = "1"
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *integral) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"integral": {
			Description: "This will show the sum over time, sort of like a continuous addition function.\nUseful for finding totals or trends in metrics that are collected per minute.\n\nExample:\n\n.. code-block:: none\n\n  &target=integral(company.sales.perMinute)\n\nThis would start at zero on the left side of the graph, adding the sales each\nminute, and show the total sales for the time period selected at the right\nside, (time now, or the time specified by '&until=').",
			Function:    "integral(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "integral",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
