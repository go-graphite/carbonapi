package countSeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("countSeries", &CountSeries{})
}

type CountSeries struct {
	interfaces.FunctionBase
}

// countSeries(seriesList)
func (f *CountSeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(civil): Check that series have equal length
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	r := *args[0]
	r.Name = fmt.Sprintf("countSeries(%s)", e.RawArgs())
	r.Values = make([]float64, len(args[0].Values))
	r.IsAbsent = make([]bool, len(args[0].Values))
	count := float64(len(args))

	for i := range args[0].Values {
		r.Values[i] = count
	}

	return []*types.MetricData{&r}, nil
}
