package averageSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type averageSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &averageSeries{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"avg", "averageSeries"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// averageSeries(*seriesLists)
func (f *averageSeries) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	e.SetTarget("averageSeries")
	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		sum := 0.0
		for _, value := range values {
			sum += value
		}
		return sum / float64(len(values))
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *averageSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"avg": {
			Description: "Short Alias: avg()\n\nTakes one metric or a wildcard seriesList.\nDraws the average value of all metrics passed at each time.\n\nExample:\n\n.. code-block:: none\n\n  &target=averageSeries(company.server.*.threads.busy)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``average``.",
			Function:    "avg(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "avg",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"averageSeries": {
			Description: "Short Alias: avg()\n\nTakes one metric or a wildcard seriesList.\nDraws the average value of all metrics passed at each time.\n\nExample:\n\n.. code-block:: none\n\n  &target=averageSeries(company.server.*.threads.busy)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``average``.",
			Function:    "averageSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "averageSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
