package areaBetween

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type areaBetween struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &areaBetween{}
	functions := []string{"areaBetween"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *areaBetween) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	if len(arg) != 2 {
		return nil, fmt.Errorf("areaBetween needs exactly two arguments (%d given)", len(arg))
	}

	name := e.Target() + "(" + e.RawArgs() + ")"

	lower := arg[0].CopyTag(name, arg[0].Tags)
	lower.Stacked = true
	lower.StackName = types.DefaultStackName
	lower.Invisible = true

	upper := arg[1].CopyTag(name, arg[1].Tags)
	upper.Stacked = true
	upper.StackName = types.DefaultStackName

	vals := make([]float64, len(upper.Values))

	for i, v := range upper.Values {
		vals[i] = v - lower.Values[i]
	}

	upper.Values = vals

	return []*types.MetricData{lower, upper}, nil
}

func (f *areaBetween) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"areaBetween": {
			Name: "areaBetween",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Draws the vertical area in between the two series in seriesList. Useful for\nvisualizing a range such as the minimum and maximum latency for a service.\n\nareaBetween expects **exactly one argument** that results in exactly two series\n(see example below). The order of the lower and higher values series does not\nmatter. The visualization only works when used in conjunction with\n``areaMode=stacked``.\n\nMost likely use case is to provide a band within which another metric should\nmove. In such case applying an ``alpha()``, as in the second example, gives\nbest visual results.\n\nExample:\n\n.. code-block:: none\n\n  &target=areaBetween(service.latency.{min,max})&areaMode=stacked\n\n  &target=alpha(areaBetween(service.latency.{min,max}),0.3)&areaMode=stacked\n\nIf for instance, you need to build a seriesList, you should use the ``group``\nfunction, like so:\n\n.. code-block:: none\n\n  &target=areaBetween(group(minSeries(a.*.min),maxSeries(a.*.max)))",
			Function:    "areaBetween(seriesList)",
			Group:       "Graph",
		},
	}
}
