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
		res = append(res,
			interfaces.FunctionMetadata{Name: n, F: f},
			interfaces.FunctionMetadata{Name: n + "Series", F: f},
		)
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
			args, err = helper.GetSeriesArgsAndRemoveNonExisting(ctx, e, from, until, values)
			if err != nil {
				return nil, err
			}
			callback = strings.Replace(e.Target(), "Series", "", 1)
			isAggregateFunc = false
		}
	} else {
		args, err = helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
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
	target := callback + "%sSeries"

	e.SetTarget(target)
	if isAggregateFunc {
		e.SetRawArgs(e.Args()[0].Target())
	}
	return helper.AggregateSeries(e, args, aggFunc)
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aggregate) Description() map[string]types.FunctionDescription {
	// TODO(Civil): this should be reworked. Graphite do not provide consistent mappings for some of the consolidation
	// functions. Also it's very easy to miss something obvious here.
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
					Options:  types.StringsToSuggestionList(consolidations.AvailableConsolidationFuncs()),
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
		"stddev": {
			Description: "Short form: stddev()\n\nTakes one metric or a wildcard seriesList.\nDraws the standard deviation of all metrics passed at each time.\n\nExample:\n\n.. code-block:: none\n\n  &target=stddevSeries(company.server.*.threads.busy)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``stddev``.",
			Function:    "stddev(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "stddev",
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
			Description: "Short form: stddev()\n\nTakes one metric or a wildcard seriesList.\nDraws the standard deviation of all metrics passed at each time.\n\nExample:\n\n.. code-block:: none\n\n  &target=stddevSeries(company.server.*.threads.busy)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``stddev``.",
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
		"count": {
			Description: "Draws a horizontal line representing the number of nodes found in the seriesList.\n\n.. code-block:: none\n\n  &target=count(carbon.agents.*.*)",
			Function:    "count(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "count",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"countSeries": {
			Description: "Draws a horizontal line representing the number of nodes found in the seriesList.\n\n.. code-block:: none\n\n  &target=countSeries(carbon.agents.*.*)",
			Function:    "countSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "countSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"diff": {
			Description: "Subtracts series 2 through n from series 1.\n\nExample:\n\n.. code-block:: none\n\n  &target=diff(service.connections.total,service.connections.failed)\n\nTo diff a series and a constant, one should use offset instead of (or in\naddition to) diffSeries\n\nExample:\n\n.. code-block:: none\n\n  &target=offset(service.connections.total,-5)\n\n  &target=offset(diffSeries(service.connections.total,service.connections.failed),-4)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``diff``.",
			Function:    "diff(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "diff",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"diffSeries": {
			Description: "Subtracts series 2 through n from series 1.\n\nExample:\n\n.. code-block:: none\n\n  &target=diffSeries(service.connections.total,service.connections.failed)\n\nTo diff a series and a constant, one should use offset instead of (or in\naddition to) diffSeries\n\nExample:\n\n.. code-block:: none\n\n  &target=offset(service.connections.total,-5)\n\n  &target=offset(diffSeries(service.connections.total,service.connections.failed),-4)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``diff``.",
			Function:    "diffSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "diffSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"multiply": {
			Description: "Takes two or more series and multiplies their points. A constant may not be\nused. To multiply by a constant, use the scale() function.\n\nExample:\n\n.. code-block:: none\n\n  &target=multiplySeries(Series.dividends,Series.divisors)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``multiply``.",
			Function:    "multiply(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "multiply",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"multiplySeries": {
			Description: "Takes two or more series and multiplies their points. A constant may not be\nused. To multiply by a constant, use the scale() function.\n\nExample:\n\n.. code-block:: none\n\n  &target=multiplySeries(Series.dividends,Series.divisors)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``multiply``.",
			Function:    "multiplySeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "multiplySeries",
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
