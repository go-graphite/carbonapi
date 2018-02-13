package minMax

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	metadata.RegisterFunction("maxSeries", &Max{})
	metadata.RegisterFunction("minSeries", &Min{})
}

// TODO: Merge with Min
type Max struct {
	interfaces.FunctionBase
}

// maxSeries(*seriesLists)
func (f *Max) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		max := math.Inf(-1)
		for _, value := range values {
			if value > max {
				max = value
			}
		}
		return max
	})
}

type Min struct {
	interfaces.FunctionBase
}

// minSeries(*seriesLists)
func (f *Min) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		min := math.Inf(1)
		for _, value := range values {
			if value < min {
				min = value
			}
		}
		return min
	})
}
