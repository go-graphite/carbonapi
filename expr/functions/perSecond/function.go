package perSecond

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
	metadata.RegisterFunction("perSecond", &function{})
}

type function struct {
	interfaces.FunctionBase
}

// perSecond(seriesList, maxValue=None)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	maxValue, err := e.GetFloatArgDefault(1, math.NaN())
	if err != nil {
		return nil, err
	}

	var result []*types.MetricData
	for _, a := range args {
		r := *a
		if len(e.Args()) == 1 {
			r.Name = fmt.Sprintf("%s(%s)", e.Target(), a.Name)
		} else {
			r.Name = fmt.Sprintf("%s(%s,%g)", e.Target(), a.Name, maxValue)
		}
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
				r.Values[i] = diff / float64(a.StepTime)
			} else if !math.IsNaN(maxValue) && maxValue >= v {
				r.Values[i] = (maxValue - prev + v + 1) / float64(a.StepTime)
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
