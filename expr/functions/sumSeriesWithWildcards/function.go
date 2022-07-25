package sumSeriesWithWildcards

import (
	"context"
	"math"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type sumSeriesWithWildcards struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sumSeriesWithWildcards{}
	functions := []string{"sumSeriesWithWildcards"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sumSeriesWithWildcards(*seriesLists)
func (f *sumSeriesWithWildcards) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	nodeList := make([]string, 0, 256)
	groups := make(map[string][]*types.MetricData)

	for _, a := range args {
		metric := a.Tags["name"]
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

	results := make([]*types.MetricData, len(nodeList))
	for n, series := range nodeList {
		args := groups[series]
		name := "sumSeriesWithWildcards(" + series + ")"
		r := args[0].CopyTag(name, map[string]string{"name": series})
		r.Values = make([]float64, len(args[0].Values))
		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if math.IsNaN(v) {
					continue
				}
				atLeastOne[i] = true
				r.Values[i] += v
			}
		}

		for i, v := range atLeastOne {
			if !v {
				r.Values[i] = math.NaN()
			}
		}

		results[n] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sumSeriesWithWildcards) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
