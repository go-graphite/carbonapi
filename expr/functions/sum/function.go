package sum

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &sum{}
	functions := []string{"sum", "sumSeries"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type sum struct {
	interfaces.FunctionBase
}

// sumSeries(*seriesLists)
func (f *sum) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	e.SetTarget("sumSeries")
	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		sum := 0.0
		for _, value := range values {
			sum += value
		}
		return sum
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sum) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"sum": {
			Description: "Short form: sum()\n\nThis will add metrics together and return the sum at each datapoint. (See\nintegral for a sum over time)\n\nExample:\n\n.. code-block:: none\n\n  &target=sum(company.server.application*.requestsHandled)\n\nThis would show the sum of all requests handled per minute (provided\nrequestsHandled are collected once a minute).   If metrics with different\nretention rates are combined, the coarsest metric is graphed, and the sum\nof the other metrics is averaged for the metrics with finer retention rates.\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``sum``.",
			Function:    "sum(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "sum",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sumSeries": {
			Description: "Short form: sum()\n\nThis will add metrics together and return the sum at each datapoint. (See\nintegral for a sum over time)\n\nExample:\n\n.. code-block:: none\n\n  &target=sum(company.server.application*.requestsHandled)\n\nThis would show the sum of all requests handled per minute (provided\nrequestsHandled are collected once a minute).   If metrics with different\nretention rates are combined, the coarsest metric is graphed, and the sum\nof the other metrics is averaged for the metrics with finer retention rates.\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``sum``.",
			Function:    "sumSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "sumSeries",
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
