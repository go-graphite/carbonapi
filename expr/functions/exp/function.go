package exp

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type exp struct {
	interfaces.FunctionBase
}

// offset(seriesList,factor)
func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &exp{}
	for _, n := range []string{"exp"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *exp) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range args {
		r := a.CopyLink()
		r.Name = "exp(" + a.Name + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["exp"] = "e"

		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = math.Exp(v)
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *exp) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"exp": {
			Description: "Raise e to the power of the datapoint, where e = 2.718281… is the base of natural logarithms.\n\nExample:\n\n.. code-block:: none\n\n  &target=exp(Server.instance01.threads.busy)",
			Function:    "exp(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "exp",
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
