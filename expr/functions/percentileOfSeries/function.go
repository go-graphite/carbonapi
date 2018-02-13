package percentileOfSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &Function{}
	functions := []string{"percentileOfSeries"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// percentileOfSeries(seriesList, n, interpolate=False)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	percent, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	interpolate, err := e.GetBoolNamedOrPosArgDefault("interpolate", 2, false)
	if err != nil {
		return nil, err
	}

	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		return helper.Percentile(values, percent, interpolate)
	})
}
