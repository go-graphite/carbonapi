package rangeOfSeries

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
	metadata.RegisterFunction("rangeOfSeries", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// rangeOfSeries(*seriesLists)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	series, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	r := *series[0]
	r.Name = fmt.Sprintf("%s(%s)", e.Target(), e.RawArgs())
	r.Values = make([]float64, len(series[0].Values))
	r.IsAbsent = make([]bool, len(series[0].Values))

	for i := range series[0].Values {
		var min, max float64
		count := 0
		for _, s := range series {
			if s.IsAbsent[i] {
				continue
			}

			if count == 0 {
				min = s.Values[i]
				max = s.Values[i]
			} else {
				min = math.Min(min, s.Values[i])
				max = math.Max(max, s.Values[i])
			}

			count++
		}

		if count >= 2 {
			r.Values[i] = max - min
		} else {
			r.IsAbsent[i] = true
		}
	}
	return []*types.MetricData{&r}, err
}
