package aliasByTags

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"strings"
)

const NAME = "name"

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

func metricToTagMap(s string) map[string]string {
	r := make(map[string]string)
	for _, p := range strings.Split(s, ";") {
		if strings.Contains(p, "=") {
			tagValue := strings.SplitN(p, "=", 2)
			r[tagValue[0]] = tagValue[1]
		} else {
			r[NAME] = p
		}
	}
	return r
}

func (f *aliasByTags) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	tags, err := e.GetStringArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		var matched []string
		var name string
		metricTags := metricToTagMap(a.Name)
		for _, tag := range tags {
			if value, ok := metricTags[tag]; ok {
				if tag == NAME {
					name = value
				} else {
					matched = append(matched, strings.Join([]string{tag, value}, "="))
				}
			}
		}
		if len(name) > 0 {
			matched = append([]string{name}, matched...)
		}
		r := *a
		if len(matched) > 0 {
			r.Name = strings.Join(matched, ";")
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
					Name:     "nodes",
					Required: true,
					Type:     types.Tag,
				},
			},
		},
	}
}
