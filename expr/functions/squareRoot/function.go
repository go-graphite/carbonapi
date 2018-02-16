package squareRoot

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	f := &squareRoot{}
	functions := []string{"squareRoot"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type squareRoot struct {
	interfaces.FunctionBase
}

// squareRoot(seriesList)
func (f *squareRoot) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("squareRoot(%s)", a.Name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = math.Sqrt(v)
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *squareRoot) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
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
