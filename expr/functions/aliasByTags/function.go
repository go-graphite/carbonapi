package aliasByTags

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"strings"
)

type aliasByTags struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aliasByTags{}
	for _, n := range []string{"aliasByTags"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasByTags) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	tags, err := e.GetNodeOrTagArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		var matched []string
		metricTags := a.Tags
		nodes := strings.Split(metricTags["name"], ".")
		for _, tag := range tags {
			if tag.IsTag {
				tagStr := tag.Value.(string)
				matched = append(matched, metricTags[tagStr])
			} else {
				f := tag.Value.(int)
				if f < 0 {
					f += len(nodes)
				}
				if f >= len(nodes) || f < 0 {
					continue
				}
				matched = append(matched, nodes[f])
			}
		}
		r := *a
		if len(matched) > 0 {
			r.Name = strings.Join(matched, ".")
		}
		results = append(results, &r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aliasByTags) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByTags": {
			Description: "Takes a seriesList and applies an alias derived from one or more tags",
			Function:    "aliasByTags(seriesList, *tags)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByTags",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "tags",
					Required: true,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}
