package aggregateWithWildcards

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/helper/metric"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type aggregateWithWildcards struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &aggregateWithWildcards{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aggregateWithWildcards"} {
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

func (f *aggregateWithWildcards) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	callback, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(2)
	if err != nil {
		return nil, err
	}

	aggFunc, ok := consolidations.ConsolidationToFunc[callback]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", callback)
	}
	target := fmt.Sprintf("%sSeries", callback)
	e.SetTarget(target)
	e.SetRawArgs(e.Args()[0].Target())

	groups := make(map[string][]*types.MetricData)
	nodeList := []string{}

	for _, a := range args {
		metric := metric.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")
		var s []string
		// Yes, this is O(n^2), but len(nodes) < 10 and len(fields) < 3
		// Iterating an int slice is faster than a map for n ~ 30
		// http://www.antoine.im/posts/someone_is_wrong_on_the_internet
		for i, n := range nodes {
			if !helper.Contains(fields, i) {
				s = append(s, n)
			}
		}

		node := strings.Join(s, ".")

		if len(groups[node]) == 0 {
			nodeList = append(nodeList, node)
		}

		groups[node] = append(groups[node], a)
	}

	results := make([]*types.MetricData, 0, len(groups))

	for _, node := range nodeList {
		res, err := helper.AggregateSeries(e, groups[node], aggFunc, -1) // Pass in -1 for xFilesFactor because aggregateWithWildcards doesn't support xFilesFactor
		if err != nil {
			return nil, err
		}
		res[0].Name = node

		results = append(results, res...)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aggregateWithWildcards) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aggregateWithWildcards": {
			Name:        "aggregateWithWildcards",
			Function:    "aggregateWithWildcards(seriesList, func, *positions)",
			Description: "Call aggregator after inserting wildcards at the given position(s).\n\nExample:\n\n.. code-block:: none\n\n &target=aggregateWithWildcards(host.cpu-[0-7].cpu-{user,system}.value, 'sum', 1)",
			Module:      "graphite.render.functions",
			Group:       "Calculate",
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
					Name:     "positions",
					Type:     types.Node,
					Required: true,
				},
			},
		},
	}
}
