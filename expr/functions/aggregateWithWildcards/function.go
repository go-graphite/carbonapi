package aggregateWithWildcards

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

type aggregateWithWildcards struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &aggregateWithWildcards{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"multiplySeriesWithWildcards", "averageSeriesWithWildcards", "aggregateWithWildcards", "sumSeriesWithWildcards"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aggregateWithWildcards) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var callback string
	fieldsArgsIx := 1
	switch e.Target() {
	case "aggregateWithWildcards":
		callback, err = e.GetStringArg(1)
		if err != nil {
			return nil, err
		}
		fieldsArgsIx = 2
	case "sumSeriesWithWildcards":
		callback = "sum"
	case "averageSeriesWithWildcards":
		callback = "average"
	case "multiplySeriesWithWildcards":
		callback = "multiply"
	}

	var fields []int
	fields, err = e.GetIntArgs(fieldsArgsIx)
	if err != nil {
		return nil, err
	}

	aggFunc, ok := consolidations.ConsolidationToFunc[callback]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", callback)
	}
	groups := make(map[string][]*types.MetricData)
	nodeList := make([]string, 0, 256)

	for _, a := range args {
		metric := types.ExtractNameTag(a.Name)
		nodes := strings.Split(metric, ".")
		s := make([]string, 0, len(nodes))
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
		// Pass in -1 for xFilesFactor because aggregateWithWildcards doesn't support xFilesFactor
		res, err := helper.AggregateSeries(e, groups[node], aggFunc, -1, false)
		if err != nil {
			return nil, err
		}
		res[0].Name = node
		res[0].Tags["aggregatedBy"] = callback

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
		"averageSeriesWithWildcards": {
			Description: "Call averageSeries after inserting wildcards at the given position(s).\n\nExample:\n\n.. code-block:: none\n\n  &target=averageSeriesWithWildcards(host.cpu-[0-7}.cpu-{user,system}.value, 1)\n\nThis would be the equivalent of\n\n.. code-block:: none\n\n  &target=averageSeries(host.*.cpu-user.value)&target=averageSeries(host.*.cpu-system.value)\n\nThis is an alias for :py:func:`aggregateWithWildcards <aggregateWithWildcards>` with aggregation ``average``.",
			Function:    "averageSeriesWithWildcards(seriesList, *position)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "averageSeriesWithWildcards",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "position",
					Type:     types.Node,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
		"multiplySeriesWithWildcards": {
			Description: "Call multiplySeries after inserting wildcards at the given position(s).\n\nExample:\n\n.. code-block:: none\n\n  &target=multiplySeriesWithWildcards(web.host-[0-7}.{avg-response,total-request}.value, 2)\n\nThis would be the equivalent of\n\n.. code-block:: none\n\n  &target=multiplySeries(web.host-0.{avg-response,total-request}.value)&target=multiplySeries(web.host-1.{avg-response,total-request}.value)...\n\nThis is an alias for :py:func:`aggregateWithWildcards <aggregateWithWildcards>` with aggregation ``multiply``.",
			Function:    "multiplySeriesWithWildcards(seriesList, *position)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "multiplySeriesWithWildcards",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "position",
					Type:     types.Node,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
		"sumSeriesWithWildcards": {
			Description: "Call sumSeries after inserting wildcards at the given position(s).\n\nExample:\n\n.. code-block:: none\n\n  &target=sumSeriesWithWildcards(host.cpu-[0-7}.cpu-{user,system}.value, 1)\n\nThis would be the equivalent of\n\n.. code-block:: none\n\n  &target=sumSeries(host.cpu-[0-7}.cpu-user.value)&target=sumSeries(host.cpu-[0-7}.cpu-system.value)\n\nThis is an alias for :py:func:`aggregateWithWildcards <aggregateWithWildcards>` with aggregation ``sum``.",
			Function:    "sumSeriesWithWildcards(seriesList, *position)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "sumSeriesWithWildcards",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "position",
					Type:     types.Node,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
