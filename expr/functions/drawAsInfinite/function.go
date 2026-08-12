package drawAsInfinite

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type drawAsInfinite struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &drawAsInfinite{}
	functions := []string{"drawAsInfinite"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *drawAsInfinite) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLink()
		r.Name = "drawAsInfinite(" + a.Name + ")"
		r.DrawAsInfinite = true
		r.Tags["drawAsInfinite"] = "1"

		results[i] = r
	}
	return results, nil
}

func (f *drawAsInfinite) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"drawAsInfinite": {
			Name: "drawAsInfinite",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList.\nIf the value is zero, draw the line at 0.  If the value is above zero, draw\nthe line at infinity. If the value is null or less than zero, do not draw\nthe line.\n\nUseful for displaying on/off metrics, such as exit codes. (0 = success,\nanything else = failure.)\n\nExample:\n\n.. code-block:: none\n\n  drawAsInfinite(Testing.script.exitCode)",
			Function:    "drawAsInfinite(seriesList)",
			Group:       "Graph",
		},
	}
}
