package multiplySeriesWithWildcards

import (
	"context"
	"math"
	"strings"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type multiplySeriesWithWildcards struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &multiplySeriesWithWildcards{}
	functions := []string{"multiplySeriesWithWildcards"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// multiplySeriesWithWildcards(seriesList, *position)
func (f *multiplySeriesWithWildcards) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/* TODO(dgryski): make sure the arrays are all the same 'size'
	   (duplicated from sumSeriesWithWildcards because of similar logic but multiplication) */
	args, err := helper.GetSeriesArg(ctx, e.Arg(0), from, until, values)
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

	results := make([]*types.MetricData, 0, len(nodeList))
	commonTags := helper.GetCommonTags(args)
	commonTags["aggregatedBy"] = "multiply"

	for _, series := range nodeList {
		args := groups[series]
		r := args[0].CopyLink()
		r.Name = series
		r.Tags = make(map[string]string)
		for k, v := range args[0].Tags {
			r.Tags[k] = v
		}
		r.Tags["name"] = series

		if _, ok := commonTags["name"]; !ok {
			commonTags["name"] = r.Name
		}
		r.Tags = commonTags

		r.Values = make([]float64, len(args[0].Values))

		atLeastOne := make([]bool, len(args[0].Values))
		hasVal := make([]bool, len(args[0].Values))

		for _, arg := range args {
			for i, v := range arg.Values {
				if math.IsNaN(v) {
					continue
				}

				atLeastOne[i] = true
				if !hasVal[i] {
					r.Values[i] = v
					hasVal[i] = true
				} else {
					r.Values[i] *= v
				}
			}
		}

		for i, v := range atLeastOne {
			if !v {
				r.Values[i] = math.NaN()
			}
		}

		if _, ok := commonTags["name"]; !ok {
			commonTags["name"] = r.Name
		}
		r.Tags = commonTags

		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *multiplySeriesWithWildcards) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
	}
}
