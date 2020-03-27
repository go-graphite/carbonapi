package groupByTags

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/tags"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type groupByTags struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &groupByTags{}
	functions := []string{"groupByTags"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// seriesByTag("name=cpu")|groupByTags("average","dc","os")
func (f *groupByTags) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	callback, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	tagNames, err := e.GetStringArgs(2)
	if err != nil {
		return nil, err
	}

	sort.Strings(tagNames)

	var results []*types.MetricData

	names := make(map[string]string)
	groups := make(map[string][]*types.MetricData)
	// name := args[1].Name

	// TODO(civil): Think how to optimize it, as it's ugly
	for _, a := range args {
		metricTags := tags.ExtractTags(a.Name)
		var keyBuilder strings.Builder
		for _, tag := range tagNames {
			value := metricTags[tag]
			keyBuilder.WriteString(";" + tag + "=" + value)
		}
		key := keyBuilder.String()
		groups[key] = append(groups[key], a)

		if name, ok := names[key]; ok {
			if name != metricTags["name"] {
				names[key] = callback
			}
		} else {
			names[key] = metricTags["name"]
		}
	}

	for k, v := range groups {
		k := k // k's reference is used later, so it's important to make it unique per loop
		v := v

		var expr string
		_, ok := consolidations.ConsolidationToFunc[callback]
		if ok {
			expr = fmt.Sprintf("aggregate(stub, \"%s\")", callback)
		} else {
			expr = fmt.Sprintf("%s(stub)", callback)
		}

		// create a stub context to evaluate the callback in
		nexpr, _, err := parser.ParseExpr(expr)
		if err != nil {
			return nil, err
		}

		nvalues := map[parser.MetricRequest][]*types.MetricData{
			parser.MetricRequest{"stub", from, until}: v,
		}

		r, err := f.Evaluator.Eval(nexpr, from, until, nvalues)
		if err != nil {
			return nil, err
		}
		if r != nil {
			r[0].Name = names[k] + k
			results = append(results, r...)
		}
	}

	return results, nil
}

func (f *groupByTags) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"groupByTags": {
			Description: "Takes a serieslist and maps a callback to subgroups within as defined by multiple tags\n\n.. code-block:: none\n\n  &target=seriesByTag(\"name=cpu\")|groupByTags(\"average\",\"dc\")\n\nWould return multiple series which are each the result of applying the \"averageSeries\" function\nto groups joined on the specified tags resulting in a list of targets like\n\n.. code-block :: none\n\n  averageSeries(seriesByTag(\"name=cpu\",\"dc=dc1\")),averageSeries(seriesByTag(\"name=cpu\",\"dc=dc2\")),...\n\nThis function can be used with all aggregation functions supported by\n:py:func:`aggregate <aggregate>`: ``average``, ``median``, ``sum``, ``min``, ``max``, ``diff``,\n``stddev``, ``range`` & ``multiply``.",
			Function:    "groupByTags(seriesList, callback, *tags)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "groupByTags",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "callback",
					Options:  consolidations.AvailableSummarizers,
					Required: true,
					Type:     types.AggFunc,
				},
				{
					Name:     "tags",
					Required: true,
					Multiple: true,
					Type:     types.Tag,
				},
			},
		},
	}
}
