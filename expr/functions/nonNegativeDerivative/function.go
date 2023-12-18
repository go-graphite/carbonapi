package nonNegativeDerivative

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
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

func (f *nonNegativeDerivative) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
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
	if _, ok := e.NamedArg("maxValue"); ok || e.ArgsLen() > 1 {
		argMask |= 1
	}
	if _, ok := e.NamedArg("minValue"); ok || e.ArgsLen() > 2 {
		argMask |= 2
	}

	var maxValueStr string
	var minValueStr string
	if hasMax {
		maxValueStr = strconv.FormatFloat(maxValue, 'g', -1, 64)
	}
	if hasMin {
		minValueStr = strconv.FormatFloat(minValue, 'g', -1, 64)
	}

	result := make([]*types.MetricData, len(args))
	for i, a := range args {
		var name string
		switch argMask {
		case 3:
			name = "nonNegativeDerivative(" + a.Name + "," + maxValueStr + "," + minValueStr + ")"
		case 2:
			name = "nonNegativeDerivative(" + a.Name + ",minValue=" + minValueStr + ")"
		case 1:
			name = "nonNegativeDerivative(" + a.Name + "," + maxValueStr + ")"
		case 0:
			name = "nonNegativeDerivative(" + a.Name + ")"
		}

		r := a.CopyLink()
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.Tags["nonNegativeDerivative"] = "1"
		result[i] = r

		if len(a.Values) == 0 {
			continue
		}

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
				r.Values[i] = diff
			} else if hasMax && maxValue >= v {
				r.Values[i] = ((maxValue - prev) + (v - minValue) + 1)
			} else if hasMin && minValue <= v {
				r.Values[i] = (v - minValue)
			} else {
				r.Values[i] = math.NaN()
			}
			prev = v
		}
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
				{
					Name: "minValue",
					Type: types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
