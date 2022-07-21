package aliasSub

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"regexp"
)

type aliasSub struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aliasSub{}
	for _, n := range []string{"aliasSub"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasSub) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	search, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	replace, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return nil, err
	}

	replace = helper.Backref.ReplaceAllString(replace, "$${$1}")

	results := make([]*types.MetricData, len(args))

	for i, a := range args {
		r := a.CopyLink()

		r.Name = re.ReplaceAllString(a.Name, replace)
		r.Tags["name"] = r.Name

		results[i] = r
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aliasSub) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasSub": {
			Description: "Runs series names through a regex search/replace.\n\n.. code-block:: none\n\n  &target=aliasSub(ip.*TCP*,\"^.*TCP(\\d+)\",\"\\1\")",
			Function:    "aliasSub(seriesList, search, replace)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasSub",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "search",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "replace",
					Required: true,
					Type:     types.String,
				},
			},
			NameChange: true, // name changed
			TagsChange: true, // name tag changed
		},
	}
}
