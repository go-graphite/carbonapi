package nPercentile

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("nPercentile", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// nPercentile(seriesList, n)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	percent, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("nPercentile(%s,%g)", a.Name, percent)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		var values []float64
		for i, v := range a.IsAbsent {
			if !v {
				values = append(values, a.Values[i])
			}
		}

		value := helper.Percentile(values, percent, true)
		for i := range r.Values {
			r.Values[i] = value
		}

		results = append(results, &r)
	}
	return results, nil
}
