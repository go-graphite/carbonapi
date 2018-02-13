package removeBelowSeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
	"strings"
)

func init() {
	f := &Function{}
	functions := []string{"removeBelowValue", "removeAboveValue", "removeBelowPercentile", "removeAbovePercentile"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// removeBelowValue(seriesLists, n), removeAboveValue(seriesLists, n), removeBelowPercentile(seriesLists, percent), removeAbovePercentile(seriesLists, percent)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	number, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	condition := func(v float64, threshold float64) bool {
		return v < threshold
	}

	if strings.HasPrefix(e.Target(), "removeAbove") {
		condition = func(v float64, threshold float64) bool {
			return v > threshold
		}
	}

	var results []*types.MetricData

	for _, a := range args {
		threshold := number
		if strings.HasSuffix(e.Target(), "Percentile") {
			var values []float64
			for i, v := range a.IsAbsent {
				if !v {
					values = append(values, a.Values[i])
				}
			}

			threshold = helper.Percentile(values, number, true)
		}

		r := *a
		r.Name = fmt.Sprintf("%s(%s, %g)", e.Target(), a.Name, number)
		r.IsAbsent = make([]bool, len(a.Values))
		r.Values = make([]float64, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] || condition(v, threshold) {
				r.Values[i] = math.NaN()
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v
		}

		results = append(results, &r)
	}

	return results, nil
}
