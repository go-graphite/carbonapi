package consolidateBy

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &Function{}
	functions := []string{"consolidateBy"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// consolidateBy(seriesList, aggregationMethod)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	name, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a

		var f func([]float64, []bool) (float64, bool)

		switch name {
		case "max":
			f = types.AggMax
		case "min":
			f = types.AggMin
		case "sum":
			f = types.AggSum
		case "average":
			f = types.AggMean
		}

		r.AggregateFunction = f

		results = append(results, &r)
	}

	return results, nil
}
