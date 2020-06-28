package aggregate

import (
	"context"
	"fmt"

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
	}
	return res
}

// aggregate(*seriesLists)
func (f *aggregate) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	callback, err := e.GetStringArg(1)
	if err != nil {
		if e.Target() == "aggregate" {
			return nil, err
		} else {
			callback = e.Target()
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
	e.SetRawArgs(e.Args()[0].Target())
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
	}
}
