package color

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type color struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &color{}
	functions := []string{"color"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *color) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	color, err := e.GetStringArg(1) // get color
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLinkTags()
		r.Color = color
		results[i] = r
	}

	return results, nil
}

func (f *color) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"color": {
			Name: "color",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "theColor",
					Required: true,
					Type:     types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Assigns the given color to the seriesList\n\nExample:\n\n.. code-block:: none\n\n  &target=color(collectd.hostname.cpu.0.user, 'green')\n  &target=color(collectd.hostname.cpu.0.system, 'ff0000')\n  &target=color(collectd.hostname.cpu.0.idle, 'gray')\n  &target=color(collectd.hostname.cpu.0.idle, '6464ffaa')",
			Function:    "color(seriesList, theColor)",
			Group:       "Graph",
		},
	}
}
