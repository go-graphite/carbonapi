package stddevSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	metadata.RegisterFunction("stddevSeries", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// stddevSeries(*seriesLists)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	e.SetTarget("stddevSeries")
	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		sum := 0.0
		diffSqr := 0.0
		for _, value := range values {
			sum += value
		}
		average := sum / float64(len(values))
		for _, value := range values {
			diffSqr += (value - average) * (value - average)
		}
		return math.Sqrt(diffSqr / float64(len(values)))
	})
}
