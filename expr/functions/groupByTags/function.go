package groupByTags

import (
	"context"
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
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
func (f *groupByTags) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 3 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
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

	var named bool
	for _, tag := range tagNames {
		if tag == "name" {
			named = true
			break
		}
	}

	names := make(map[string]string)
	tags := make(map[string]map[string]string)
	groups := make(map[string][]*types.MetricData)
	// name := args[1].Name

	// TODO(civil): Think how to optimize it, as it's ugly
	for _, a := range args {
		metricTags := a.Tags
		var keyBuilder strings.Builder
		keyBuilder.Grow(len(a.Name))
		if named {
			keyBuilder.WriteString(metricTags["name"])
		} else {
			keyBuilder.WriteString(callback)
		}
		for _, tag := range tagNames {
			if tag != "name" {
				keyBuilder.WriteString(";")
				keyBuilder.WriteString(tag)
				keyBuilder.WriteString("=")
				keyBuilder.WriteString(metricTags[tag])
			}
		}
		key := keyBuilder.String()
		groups[key] = append(groups[key], a)

		if _, ok := names[key]; !ok {
			newTags := make(map[string]string)
			for _, tag := range tagNames {
				newTags[tag] = metricTags[tag]
			}
			if named {
				names[key] = metricTags["name"]
			} else {
				names[key] = callback
				newTags["name"] = callback
			}
			tags[key] = newTags
		}
	}

	results := make([]*types.MetricData, 0, len(groups))

	for k, v := range groups {
		k := k // k's reference is used later, so it's important to make it unique per loop
		v := v

		var expr string
		_, ok := consolidations.ConsolidationToFunc[callback]
		if ok {
			expr = "aggregate(stub, \"" + callback + "\")"
		} else {
			expr = callback + "(stub)"
		}

		// create a stub context to evaluate the callback in
		nexpr, _, err := parser.ParseExpr(expr)
		if err != nil {
			return nil, err
		}

		nvalues := map[parser.MetricRequest][]*types.MetricData{
			{Metric: "stub", From: from, Until: until}: v,
		}

		r, err := f.Evaluator.Eval(ctx, nexpr, from, until, nvalues)
		if err != nil {
			return nil, err
		}
		if r != nil {
			rg := types.CopyMetricDataSliceWithTags(r, k, tags[k])
			results = append(results, rg...)
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
					Options:  types.StringsToSuggestionList(consolidations.AvailableSummarizers),
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
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
