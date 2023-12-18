package sortByName

import (
	"context"
	"sort"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type sortByName struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sortByName{}
	functions := []string{"sortByName"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sortByName(seriesList, natural=false)
func (f *sortByName) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	isNatSort, err := e.GetBoolNamedOrPosArgDefault("natural", 1, false)
	if err != nil {
		return nil, err
	}

	isReverseSort, err := e.GetBoolNamedOrPosArgDefault("reverse", 2, false)
	if err != nil {
		return nil, err
	}

	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	var dataToSort sort.Interface
	if isNatSort {
		dataToSort = helper.ByNameNatural(arg)
	} else {
		dataToSort = helper.ByName(arg)
	}

	if isReverseSort {
		dataToSort = sort.Reverse(dataToSort)
	}

	sort.Sort(dataToSort)
	return arg, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sortByName) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"sortByName": {
			Description: "Takes one metric or a wildcard seriesList.\nSorts the list of metrics by the metric name using either alphabetical order or natural sorting.\nNatural sorting allows names containing numbers to be sorted more naturally, e.g:\n- Alphabetical sorting: server1, server11, server12, server2\n- Natural sorting: server1, server2, server11, server12",
			Function:    "sortByName(seriesList, natural=False, reverse=False)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByName",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(false),
					Name:    "natural",
					Type:    types.Boolean,
				},
				{
					Default: types.NewSuggestion(false),
					Name:    "reverse",
					Type:    types.Boolean,
				},
			},
		},
	}
}
