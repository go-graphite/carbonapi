package sortBy

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type sortBy struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sortBy{}
	functions := []string{"sortByMaxima", "sortByMinima", "sortByTotal", "sortBy"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sortByMaxima(seriesList), sortByMinima(seriesList), sortByTotal(seriesList)
func (f *sortBy) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	reverse, err := e.GetBoolArgDefault(2, false)
	if err != nil {
		return nil, err
	}
	ascending := !reverse

	sortByFunc, err := e.GetStringArgDefault(1, "average")
	if err != nil {
		return nil, err
	}

	aggFuncMap := map[string]struct {
		name      string
		ascending bool
	}{
		"sortByTotal":  {"sum", false},
		"sortByMaxima": {"max", false},
		"sortByMinima": {"min", true},
		"sortBy":       {sortByFunc, true},
	}

	target := e.Target()
	aggFunc, exists := aggFuncMap[target]
	if !exists {
		return nil, fmt.Errorf("invalid function called: %s", target)
	}
	if err := consolidations.CheckValidConsolidationFunc(aggFunc.name); err != nil {
		return nil, err
	}

	// some function by default are not ascending so we need to reverse behaviour
	if !aggFunc.ascending {
		ascending = !ascending
	}

	return doSort(aggFunc.name, ascending, original), nil
}

func doSort(aggFuncName string, ascending bool, original []*types.MetricData) []*types.MetricData {
	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	vals := make([]float64, len(arg))

	for i, a := range arg {
		vals[i] = consolidations.SummarizeValues(aggFuncName, a.Values, a.XFilesFactor)
	}

	if ascending {
		sort.Sort(helper.ByVals{Vals: vals, Series: arg})
	} else {
		sort.Sort(sort.Reverse(helper.ByVals{Vals: vals, Series: arg}))
	}

	return arg
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sortBy) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"sortByMaxima": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics in descending order by the maximum value across the time period\nspecified.  Useful with the &areaMode=all parameter, to keep the\nlowest value lines visible.\n\nExample:\n\n.. code-block:: none\n\n  &target=sortByMaxima(server*.instance*.memory.free)",
			Function:    "sortByMaxima(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByMaxima",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sortByMinima": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics by the lowest value across the time period\nspecified, including only series that have a maximum value greater than 0.\n\nExample:\n\n.. code-block:: none\n\n  &target=sortByMinima(server*.instance*.memory.free)",
			Function:    "sortByMinima(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByMinima",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sortByTotal": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics in descending order by the sum of values across the time period\nspecified.",
			Function:    "sortByTotal(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByTotal",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sortBy": {
			Description: "Takes one metric or a wildcard seriesList followed by an aggregation function and an optional reverse parameter.\nReturns the metrics sorted according to the specified function.",
			Function:    "sortBy(seriesList, func='average', reverse=False)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortBy",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "func",
					Required: false,
					Type:     types.AggFunc,
				},
				{
					Name:     "reverse",
					Required: false,
					Type:     types.Boolean,
				},
			},
		},
	}
}
