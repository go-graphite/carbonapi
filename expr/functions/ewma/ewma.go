package ewma

import (
	"fmt"
	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("ewma", &Function{})
	metadata.RegisterFunction("exponentialWeightedMovingAverage", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// ewma(seriesList, alpha)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	alpha, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	e.SetTarget("ewma")

	// ugh, helper.ForEachSeriesDo does not handle arguments properly
	var results []*types.MetricData
	for _, a := range arg {
		name := fmt.Sprintf("ewma(%s,%v)", a.Name, alpha)

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		ewma := onlinestats.NewExpWeight(alpha)

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			}

			ewma.Push(v)
			r.Values[i] = ewma.Mean()
		}
		results = append(results, &r)
	}
	return results, nil
}
