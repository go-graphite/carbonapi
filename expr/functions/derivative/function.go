package derivative

import (
	"math"

	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("derivative", &Derivative{})
	metadata.RegisterFunction("nonNegativeDerivative", &NonNegativeDerivative{})
}

type Derivative struct {
	interfaces.FunctionBase
}

// derivative(seriesList)
func (f *Derivative) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		prev := math.NaN()
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			} else if math.IsNaN(prev) {
				r.IsAbsent[i] = true
				prev = v
				continue
			}

			r.Values[i] = v - prev
			prev = v
		}
		return r
	})
}

type NonNegativeDerivative struct {
	interfaces.FunctionBase
}

// nonNegativeDerivative(seriesList, helper.MaxValue=None)
func (f *NonNegativeDerivative) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
