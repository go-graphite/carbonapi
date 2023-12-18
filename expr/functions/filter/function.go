package filter

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type filterSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &filterSeries{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"filterSeries"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

var supportedOperators = map[string]struct{}{
	"=":  struct{}{},
	"!=": struct{}{},
	">":  struct{}{},
	">=": struct{}{},
	"<":  struct{}{},
	"<=": struct{}{},
}

// filterSeries(*seriesLists)
func (f *filterSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 4 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	callback, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	operator, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}

	if _, ok := supportedOperators[operator]; !ok {
		return nil, fmt.Errorf("unsupported operator %v, supported operators: %v", operator, supportedOperators)
	}

	threshold, err := e.GetFloatArg(3)
	if err != nil {
		return nil, err
	}

	aggFunc, ok := consolidations.ConsolidationToFunc[callback]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", callback)
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, a := range args {
		val := aggFunc(a.Values)
		keepSeries := false
		switch operator {
		case "=":
			if val == threshold {
				keepSeries = true
			}
		case "!=":
			if val != threshold {
				keepSeries = true
			}
		case ">":
			if val > threshold {
				keepSeries = true
			}
		case ">=":
			if val >= threshold {
				keepSeries = true
			}
		case "<":
			if val < threshold {
				keepSeries = true
			}
		case "<=":
			if val <= threshold {
				keepSeries = true
			}
		}
		if !keepSeries {
			continue
		}

		results = append(results, a)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *filterSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"filterSeries": {
			Name:        "filterSeries",
			Function:    "filterSeries(seriesList, func, operator, threshold)",
			Description: "Takes one metric or a wildcard seriesList followed by a consolidation function, an operator and a threshold.\nDraws only the metrics which match the filter expression.\n\nExample:\n\n.. code-block:: none\n\n  &target=filterSeries(system.interface.eth*.packetsSent, 'max', '>', 1000)\n\nThis would only display interfaces which has a peak throughput higher than 1000 packets/min.\n\nSupported aggregation functions: ``average``, ``median``, ``sum``, ``min``,\n``max``, ``diff``, ``stddev``, ``range``, ``multiply`` & ``last``.\n\nSupported operators: ``=``, ``!=``, ``>``, ``>=``, ``<`` & ``<=``.",
			Module:      "graphite.render.functions",
			Group:       "Filter Series",
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
				{
					Name:     "operator",
					Type:     types.String,
					Required: true,
					Options: types.StringsToSuggestionList([]string{
						"!=",
						"<",
						"<=",
						"=",
						">",
						">=",
					}),
				},
				{
					Name:     "threshold",
					Type:     types.Float,
					Required: true,
				},
			},
		},
	}
}
