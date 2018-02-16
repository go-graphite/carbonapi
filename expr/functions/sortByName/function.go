package sortByName

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"sort"
)

func init() {
	metadata.RegisterFunction("sortByName", &sortByName{})
}

type sortByName struct {
	interfaces.FunctionBase
}

// sortByName(seriesList, natural=false)
func (f *sortByName) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	natSort, err := e.GetBoolNamedOrPosArgDefault("natural", 1, false)
	if err != nil {
		return nil, err
	}

	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	if natSort {
		sort.Sort(helper.ByNameNatural(arg))
	} else {
		sort.Sort(helper.ByName(arg))
	}

	return arg, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sortByName) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
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
					Default: "false",
					Name:    "natural",
					Type:    types.Boolean,
				},
				{
					Default: "false",
					Name:    "reverse",
					Type:    types.Boolean,
				},
			},
		},
	}
}
