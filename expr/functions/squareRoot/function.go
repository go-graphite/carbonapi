package squareRoot

import (
	"context"
	"fmt"
	"math"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type squareRoot struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &squareRoot{}
	functions := []string{"squareRoot"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// squareRoot(seriesList)
func (f *squareRoot) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("squareRoot(%s)", a.Name)
		r.Values = make([]float64, len(a.Values))
		r.Tags["squareRoot"] = "1"

		for i, v := range a.Values {
			r.Values[i] = math.Sqrt(v)
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *squareRoot) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"squareRoot": {
			Description: "Takes one metric or a wildcard seriesList, and computes the square root of each datapoint.\n\nExample:\n\n.. code-block:: none\n\n  &target=squareRoot(Server.instance01.threads.busy)",
			Function:    "squareRoot(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "squareRoot",
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
