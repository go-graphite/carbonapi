package round

import (
	"context"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type round struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &round{}
	functions := []string{"round"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// round(seriesList,precision)
func (f *round) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var withPrecision bool
	precision, withPrecision, err := e.GetIntNamedOrPosArgWithIndication("precision", 1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(arg))
	for _, a := range arg {
		r := *a
		if withPrecision {
			r.Name = fmt.Sprintf("round(%s,%d)", a.Name, precision)
		} else {
			r.Name = fmt.Sprintf("round(%s)", a.Name)
		}
		r.Values = make([]float64, len(a.Values))

		for i, v := range a.Values {
			r.Values[i] = doRound(v, precision)
		}
		results = append(results, &r)
	}
	return results, nil
}

func doRound(x float64, precision int) float64 {
	if math.IsNaN(x) {
		return x
	}
	roundTo := math.Pow10(precision)
	return math.RoundToEven(x*roundTo) / roundTo
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *round) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"round": {
			Description: "Takes one metric or a wildcard seriesList optionally followed by a precision, and rounds each\ndatapoint to the specified precision.\n\nExample:\n\n.. code-block:: none\n\n  &target=round(Server.instance01.threads.busy)\n  &target=round(Server.instance01.threads.busy,2)",
			Function:    "round(seriesList, precision)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "round",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "precision",
					Required: false,
					Type:     types.Integer,
				},
			},
		},
	}
}
