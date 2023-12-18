package mapSeries

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type mapSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &mapSeries{}
	functions := []string{"mapSeries", "map"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// mapSeries(seriesList, *mapNodes)
// Alias: map
func (f *mapSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	nodesOrTags, err := e.GetNodeOrTagArgs(1, false)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]*types.MetricData)
	var nodeList []string

	for _, a := range args {
		node := helper.AggKey(a, nodesOrTags)
		if len(groups[node]) == 0 {
			nodeList = append(nodeList, node)
		}

		groups[node] = append(groups[node], a)
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, node := range nodeList {
		results = append(results, groups[node]...)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *mapSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"mapSeries": {
			Description: "Short form: ``map()``\n\nTakes a seriesList and maps it to a list of seriesList. Each seriesList has the\ngiven mapNodes in common.\n\n.. note:: This function is not very useful alone. It should be used with :py:func:`reduceSeries`\n\n.. code-block:: none\n\n  mapSeries(servers.*.cpu.*,1) =>\n\n    [\n      servers.server1.cpu.*,\n      servers.server2.cpu.*,\n      ...\n      servers.serverN.cpu.*\n    }\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.",
			Function:    "mapSeries(seriesList, *mapNodes)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "mapSeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "mapNodes",
					Required: true,
					Type:     types.NodeOrTag,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
		"map": {
			Description: "Short form: ``map()``\n\nTakes a seriesList and maps it to a list of seriesList. Each seriesList has the\ngiven mapNodes in common.\n\n.. note:: This function is not very useful alone. It should be used with :py:func:`reduceSeries`\n\n.. code-block:: none\n\n  mapSeries(servers.*.cpu.*,1) =>\n\n    [\n      servers.server1.cpu.*,\n      servers.server2.cpu.*,\n      ...\n      servers.serverN.cpu.*\n    }\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.",
			Function:    "map(seriesList, *mapNodes)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "map",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "mapNodes",
					Required: true,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}
