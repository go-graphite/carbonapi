package integral

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("integral", &integral{})
}

type integral struct {
	interfaces.FunctionBase
}

// integral(seriesList)
func (f *integral) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		current := 0.0
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			current += v
			r.Values[i] = current
		}
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *integral) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
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
		},
	}
}