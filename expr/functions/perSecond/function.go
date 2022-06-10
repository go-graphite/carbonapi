package perSecond

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

// TODO(civil): See if it's possible to merge it with NonNegativeDerivative
type perSecond struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &perSecond{}
	functions := []string{"perSecond"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// perSecond(seriesList, maxValue=None)
func (f *perSecond) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	maxValue, err := e.GetFloatNamedOrPosArgDefault("maxValue", 1, math.NaN())
	if err != nil {
		return nil, err
	}
	minValue, err := e.GetFloatNamedOrPosArgDefault("minValue", 2, math.NaN())
	if err != nil {
		return nil, err
	}
	hasMax := !math.IsNaN(maxValue)
	hasMin := !math.IsNaN(minValue)

	if hasMax && hasMin && maxValue <= minValue {
		return nil, errors.New("minValue must be lower than maxValue")
	}
	if hasMax && !hasMin {
		minValue = 0
	}

	argMask := 0
	if _, ok := e.NamedArgs()["maxValue"]; ok || len(e.Args()) > 1 {
		argMask |= 1
	}
	if _, ok := e.NamedArgs()["minValue"]; ok || len(e.Args()) > 2 {
		argMask |= 2
	}

	var result []*types.MetricData
	for _, a := range args {
		var name string
		switch argMask {
		case 3:
			name = fmt.Sprintf("perSecond(%s,%g,%g)", a.Name, maxValue, minValue)
		case 2:
			name = fmt.Sprintf("perSecond(%s,minValue=%g)", a.Name, minValue)
		case 1:
			name = fmt.Sprintf("perSecond(%s,%g)", a.Name, maxValue)
		case 0:
			name = fmt.Sprintf("perSecond(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))

		prev := a.Values[0]
		for i, v := range a.Values {
			if i == 0 || math.IsNaN(a.Values[i]) || math.IsNaN(a.Values[i-1]) {
				r.Values[i] = math.NaN()
				prev = v
				continue
			}
			// TODO(civil): Figure out if we can optimize this now when we have NaNs
			diff := v - prev
			if diff >= 0 {
				r.Values[i] = diff / float64(a.StepTime)
			} else if hasMax && maxValue >= v {
				r.Values[i] = ((maxValue - prev) + (v - minValue) + 1) / float64(a.StepTime)
			} else if hasMin && minValue <= v {
				r.Values[i] = (v - minValue) / float64(a.StepTime)
			} else {
				r.Values[i] = math.NaN()
			}
			prev = v
		}
		result = append(result, &r)
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *perSecond) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"perSecond": {
			Description: "NonNegativeDerivative adjusted for the series time interval\nThis is useful for taking a running total metric and showing how many requests\nper second were handled.\n\nExample:\n\n.. code-block:: none\n\n  &target=perSecond(company.server.application01.ifconfig.TXPackets)\n\nEach time you run ifconfig, the RX and TXPackets are higher (assuming there\nis network traffic.) By applying the perSecond function, you can get an\nidea of the packets per second sent or received, even though you're only\nrecording the total.",
			Function:    "perSecond(seriesList, maxValue=None)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "perSecond",
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
				{
					Name: "minValue",
					Type: types.Float,
				},
			},
		},
	}
}
