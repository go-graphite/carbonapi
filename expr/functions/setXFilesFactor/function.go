package setXFilesFactor

import (
	"context"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type setXFilesFactor struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &setXFilesFactor{}
	functions := []string{"setXFilesFactor", "xFilesFactor"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// setXFilesFactor(seriesList, xFilesFactor)
func (f *setXFilesFactor) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	xFilesFactor, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	for _, a := range args {
		a.XFilesFactor = float32(xFilesFactor)
	}
	return args, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *setXFilesFactor) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"setXFilesFactor": {
			Description: "Takes one metric or a wildcard seriesList and an xFilesFactor value between 0 and 1. When a series needs to be consolidated, this sets the fraction of values in an interval that must\nnot be null for the consolidation to be considered valid. If there are not enough values then None will be returned for that interval.\n\nExample:\n\n.. code-block:: none\n\n  &target=xFilesFactor(Sales.widgets.largeBlue, 0.5)\n  &target=Servers.web01.sda1.free_space|consolidateBy('max')|xFilesFactor(0.5)",
			Function:    "setXFilesFactor(seriesList, xFilesFactor)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "setXFilesFactor",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "xFilesFactor",
					Required: false,
					Type:     types.Float,
				},
			},
		},
	}
}
