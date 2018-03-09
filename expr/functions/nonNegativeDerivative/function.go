package nonNegativeDerivative

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type nonNegativeDerivative struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &nonNegativeDerivative{}
	functions := []string{"nonNegativeDerivative"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *nonNegativeDerivative) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	maxValue, err := e.GetFloatNamedOrPosArgDefault("maxValue", 1, math.NaN())
	if err != nil {
		return nil, err
	}
	_, ok := e.NamedArgs()["maxValue"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	var result []*types.MetricData
	for _, a := range args {
		var name string
		if ok {
			name = fmt.Sprintf("nonNegativeDerivative(%s,%g)", a.Name, maxValue)
		} else {
			name = fmt.Sprintf("nonNegativeDerivative(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := a.Values[0]
		for i, v := range a.Values {
			if i == 0 || a.IsAbsent[i] || a.IsAbsent[i-1] {
				r.IsAbsent[i] = true
				prev = v
				continue
			}
			diff := v - prev
			if diff >= 0 {
				r.Values[i] = diff
			} else if !math.IsNaN(maxValue) && maxValue >= v {
				r.Values[i] = ((maxValue - prev) + v + 1)
			} else {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
			prev = v
		}
		result = append(result, &r)
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *nonNegativeDerivative) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"nonNegativeDerivative": {
			Description: "Same as the derivative function above, but ignores datapoints that trend\ndown.  Useful for counters that increase for a long time, then wrap or\nreset. (Such as if a network interface is destroyed and recreated by unloading\nand re-loading a kernel module, common with USB / WiFi cards.\n\nExample:\n\n.. code-block:: none\n\n  &target=nonNegativederivative(company.server.application01.ifconfig.TXPackets)",
			Function:    "nonNegativeDerivative(seriesList, maxValue=None)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "nonNegativeDerivative",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "maxValue",
					Type: types.Float,
				},
			},
		},
	}
}
