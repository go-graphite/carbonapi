package pow

import (
	"context"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type pow struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &pow{}
	functions := []string{"pow"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// pow(seriesList,factor)
func (f *pow) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	factor, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("pow(%s,%g)", a.Name, factor)
		r.Values = make([]float64, len(a.Values))

		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = math.Pow(v, factor)
			}
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *pow) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"pow": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and raises the datapoint\nby the power of the constant provided at each point.\n\nExample:\n\n.. code-block:: none\n\n  &target=pow(Server.instance01.threads.busy,10)\n  &target=pow(Server.instance*.threads.busy,10)",
			Function:    "pow(seriesList, factor)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "pow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "factor",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}
