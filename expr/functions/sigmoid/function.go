package sigmoid

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type sigmoid struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sigmoid{}
	functions := []string{"sigmoid"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sigmoid(seriesList)
func (f *sigmoid) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := a.CopyLink()
		r.Name = "sigmoid(" + a.Name + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["sigmoid"] = "sigmoid"

		for i, v := range a.Values {
			if math.IsNaN(v) || math.Exp(-v) == -1 { // check for -1 result as this would cause a divide by zero error
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = (1 / (1 + math.Exp(-v)))
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sigmoid) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"sigmoid": {
			Description: "Takes one metric or a wildcard seriesList and applies the sigmoid function 1 / (1 + exp(-x)) to each datapoint.\n\nExample:\n\n.. code-block:: none\n\n  &target=sigmoid(Server.instance01.threads.busy)\n  &target=sigmoid(Server.instance*.threads.busy)",
			Function:    "sigmoid(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "sigmoid",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
