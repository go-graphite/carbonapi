package scale

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type scale struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &scale{}
	functions := []string{"scale"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// scale(seriesList, factor)
func (f *scale) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	scale, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("scale(%s,%g)", a.Name, scale)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = v * scale
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *scale) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"scale": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and multiplies the datapoint\nby the constant provided at each point.\n\nExample:\n\n.. code-block:: none\n\n  &target=scale(Server.instance01.threads.busy,10)\n  &target=scale(Server.instance*.threads.busy,10)",
			Function:    "scale(seriesList, factor)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "scale",
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
