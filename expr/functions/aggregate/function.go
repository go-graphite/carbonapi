package aggregate

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type aggregate struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &aggregate{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aggregate"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	// Also register aliases for each and every summarizer
	for _, n := range consolidations.AvailableSummarizers {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
		res = append(res, interfaces.FunctionMetadata{Name: n + "Series", F: f})
	}
	return res
}

// aggregate(*seriesLists)
func (f *aggregate) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var args []*types.MetricData
	isAggregateFunc := true

	callback, err := e.GetStringArg(1)
	if err != nil {
		if e.Target() == "aggregate" {
			return nil, err
		} else {
			args, err = helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
			if err != nil {
				return nil, err
			}
			callback = strings.Replace(e.Target(), "Series", "", 1)
			isAggregateFunc = false
		}
	} else {
		args, err = helper.GetSeriesArg(e.Args()[0], from, until, values)
		if err != nil {
			return nil, err
		}
	}

	// TODO: Implement xFilesFactor
	/*
		xFilesFactor, err := e.GetFloatArgDefault(2, 0)
		if err != nil {
			return false, nil, err
		}
	*/
	aggFunc, ok := consolidations.ConsolidationToFunc[callback]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", callback)
	}
	target := fmt.Sprintf("%sSeries", callback)

	e.SetTarget(target)
	if isAggregateFunc {
		e.SetRawArgs(e.Args()[0].Target())
	}
	return helper.AggregateSeries(e, args, aggFunc)
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aggregate) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aggregate": {
			Name:        "aggregate",
			Function:    "aggregate(seriesList, func, xFilesFactor=None)",
			Description: "Aggregate series using the specified function.\n\nExample:\n\n.. code-block:: none\n\n  &target=aggregate(host.cpu-[0-7}.cpu-{user,system}.value, \"sum\")\n\nThis would be the equivalent of\n\n.. code-block:: none\n\n  &target=sumSeries(host.cpu-[0-7}.cpu-{user,system}.value)\n\nThis function can be used with aggregation functions ``average``, ``median``, ``sum``, ``min``,\n``max``, ``diff``, ``stddev``, ``count``, ``range``, ``multiply`` & ``last``.",
			Module:      "graphite.render.functions",
			Group:       "Combine",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "func",
					Type:     types.AggFunc,
					Required: true,
					Options:  consolidations.AvailableConsolidationFuncs(),
				},
				/*
					{
						Name: "xFilesFactor",
						Type: types.Float,
					},
				*/
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
		"max": {
			Description: "Takes one metric or a wildcard seriesList.\nFor each datapoint from each metric passed in, pick the maximum value and graph it.\n\nExample:\n\n.. code-block:: none\n\n  &target=maxSeries(Server*.connections.total)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``max``.",
			Function:    "maxSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "maxSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"maxSeries": {
			Description: "Takes one metric or a wildcard seriesList.\nFor each datapoint from each metric passed in, pick the maximum value and graph it.\n\nExample:\n\n.. code-block:: none\n\n  &target=maxSeries(Server*.connections.total)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``max``.",
			Function:    "maxSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "maxSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"min": {
			Description: "Takes one metric or a wildcard seriesList.\nFor each datapoint from each metric passed in, pick the minimum value and graph it.\n\nExample:\n\n.. code-block:: none\n\n  &target=minSeries(Server*.connections.total)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``min``.",
			Function:    "minSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "minSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"minSeries": {
			Description: "Takes one metric or a wildcard seriesList.\nFor each datapoint from each metric passed in, pick the minimum value and graph it.\n\nExample:\n\n.. code-block:: none\n\n  &target=minSeries(Server*.connections.total)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``min``.",
			Function:    "minSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "minSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
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
		"stddevSeries": {
			Description: "Takes one metric or a wildcard seriesList.\nDraws the standard deviation of all metrics passed at each time.\n\nExample:\n\n.. code-block:: none\n\n  &target=stddevSeries(company.server.*.threads.busy)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``stddev``.",
			Function:    "stddevSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "stddevSeries",
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
