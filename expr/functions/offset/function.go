package offset

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
	metadata.RegisterFunction("offset", &Offset{})
	metadata.RegisterFunction("offsetToZero", &OffsetToZero{})
}

type Offset struct {
	interfaces.FunctionBase
}

// offset(seriesList,factor)
func (f *Offset) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	factor, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("offset(%s,%g)", a.Name, factor)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = v + factor
		}
		results = append(results, &r)
	}
	return results, nil
}

type OffsetToZero struct {
	interfaces.FunctionBase
}

// offsetToZero(seriesList)
func (f *OffsetToZero) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		minimum := math.Inf(1)
		for i, v := range a.Values {
			if !a.IsAbsent[i] && v < minimum {
				minimum = v
			}
		}
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = v - minimum
		}
		return r
	})
}
