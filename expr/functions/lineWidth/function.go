package lineWidth

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type lineWidth struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &lineWidth{}
	functions := []string{"lineWidth"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *lineWidth) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	width, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLinkTags()
		r.LineWidth = width
		r.HasLineWidth = true
		results[i] = r
	}

	return results, nil
}

func (f *lineWidth) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"lineWidth": {
			Name: "lineWidth",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "width",
					Required: true,
					Type:     types.Float,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList, followed by a float F.\n\nDraw the selected metrics with a line width of F, overriding the default\nvalue of 1, or the &lineWidth=X.X parameter.\n\nUseful for highlighting a single metric out of many, or having multiple\nline widths in one graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=lineWidth(server01.instance01.memory.free,5)",
			Function:    "lineWidth(seriesList, width)",
			Group:       "Graph",
		},
	}
}
