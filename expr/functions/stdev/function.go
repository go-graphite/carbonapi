package stdev

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
	f := &Function{}
	functions := []string{"stdev", "stddev"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// stdev(seriesList, points, missingThreshold=0.1)
// Alias: stddev
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	points, err := e.GetIntArg(1)
	if err != nil {
		return nil, err
	}

	missingThreshold, err := e.GetFloatArgDefault(2, 0.1)
	if err != nil {
		return nil, err
	}

	minLen := int((1 - missingThreshold) * float64(points))

	var result []*types.MetricData

	for _, a := range arg {
		w := &types.Windowed{Data: make([]float64, points)}

		r := *a
		r.Name = fmt.Sprintf("stdev(%s,%d)", a.Name, points)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				// make sure missing values are ignored
				v = math.NaN()
			}
			w.Push(v)
			r.Values[i] = w.Stdev()
			if math.IsNaN(r.Values[i]) || (i >= minLen && w.Len() < minLen) {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		result = append(result, &r)
	}
	return result, nil
}
