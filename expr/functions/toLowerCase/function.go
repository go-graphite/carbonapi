package toLowerCase

import (
	"context"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type toLowerCase struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &toLowerCase{}
	functions := []string{"lower", "toLowerCase"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// toLowerCase(seriesList, *pos)
func (f *toLowerCase) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	pos, err := e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args)+1)

	for _, a := range args {
		r := a.CopyLink()

		if len(pos) == 0 {
			r.Name = strings.ToLower(a.Name)
		} else {
			for _, i := range pos {
				if i < 0 { // Handle negative indices by indexing backwards
					i = len(r.Name) + i
				}
				lowered := strings.ToLower(string(r.Name[i]))
				r.Name = r.Name[:i] + lowered + r.Name[i+1:]
			}
		}
		r.Tags["name"] = r.Name

		results = append(results, r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *toLowerCase) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"lower": {
			Description: "Takes one metric or a wildcard seriesList and lowers the case of each letter. \n Optionally, a letter position to lower case can be specified, in which case only the letter at the specified position gets lower-cased.\n The position parameter may be given multiple times. The position parameter may be negative to define a position relative to the end of the metric name.",
			Function:    "lower(seriesList, *pos)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "lower",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "pos",
					Type:     types.Node,
					Required: false,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"toLowerCase": {
			Description: "Takes one metric or a wildcard seriesList and lowers the case of each letter. \n Optionally, a letter position to lower case can be specified, in which case only the letter at the specified position gets lower-cased.\n The position parameter may be given multiple times. The position parameter may be negative to define a position relative to the end of the metric name.",
			Function:    "toLowerCase(seriesList, *pos)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "toLowerCase",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "pos",
					Type:     types.Node,
					Required: false,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
