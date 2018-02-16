package averageSeriesWithWildcards

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	metadata.RegisterFunction("averageSeriesWithWildcards", &averageSeriesWithWildcards{})
}

type averageSeriesWithWildcards struct {
	interfaces.FunctionBase
}

// averageSeriesWithWildcards(seriesLIst, *position)
func (f *averageSeriesWithWildcards) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/* TODO(dgryski): make sure the arrays are all the same 'size'
	   (duplicated from sumSeriesWithWildcards because of similar logic but aggregation) */
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	nodeList := []string{}
	groups := make(map[string][]*types.MetricData)

	for _, a := range args {
		metric := helper.ExtractMetric(a.Name)
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

	for _, series := range nodeList {
		args := groups[series]
		r := *args[0]
		r.Name = fmt.Sprintf("averageSeriesWithWildcards(%s)", series)
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		length := make([]float64, len(args[0].Values))
		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}
				atLeastOne[i] = true
				length[i]++
				r.Values[i] += v
			}
		}

		for i, v := range atLeastOne {
			if v {
				r.Values[i] = r.Values[i] / length[i]
			} else {
				r.IsAbsent[i] = true
			}
		}

		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *averageSeriesWithWildcards) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
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
		},
	}
}