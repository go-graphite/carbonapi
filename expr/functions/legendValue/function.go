package legendValue

import (
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("legendValue", &LegendValue{})
}

type LegendValue struct {
	interfaces.FunctionBase
}

// legendValue(seriesList, newName)
func (f *LegendValue) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	methods := make([]string, len(e.Args())-1)
	for i := 1; i < len(e.Args()); i++ {
		method, err := e.GetStringArg(i)
		if err != nil {
			return nil, err
		}

		methods[i-1] = method
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		for _, method := range methods {
			summary := helper.SummarizeValues(method, a.Values)
			r.Name = fmt.Sprintf("%s (%s: %f)", r.Name, method, summary)
		}

		results = append(results, &r)
	}
	return results, nil
}
