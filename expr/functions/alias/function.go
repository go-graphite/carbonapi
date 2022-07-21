package alias

import (
	"context"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type alias struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &alias{}
	for _, n := range []string{"alias"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *alias) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	alias, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	allowFormatStr, err := e.GetBoolArgDefault(2, false)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(args))
	for i, arg := range args {
		name := alias
		if allowFormatStr {
			name = strings.ReplaceAll(name, "${expr}", arg.Name)
		}

		r := arg.CopyName(name)

		results[i] = r
	}

	return results, nil
}

func (f *alias) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"alias": {
			Description: "Takes one metric or a wildcard seriesList and a string in quotes.\n" +
				"Prints the string instead of the metric name in the legend.\n" +
				"If set to True, `allowFormatStr` will replace all occurrences of `${expr}` with name of expression\n\n" +
				".. code-block:: none\n\n" +
				"  &target=alias(Sales.widgets.largeBlue,\"Large Blue Widgets\")",
			Function: "alias(seriesList, newName)",
			Group:    "Alias",
			Module:   "graphite.render.functions",
			Name:     "alias",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "newName",
					Required: true,
					Type:     types.String,
				},
				{
					Default:  types.NewSuggestion(false),
					Name:     "allowFormatStr",
					Required: false,
					Type:     types.Boolean,
				},
			},
		},
	}
}
