package removeEmptySeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &Function{}
	functions := []string{"removeEmptySeries", "removeZeroSeries"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// removeEmptySeries(seriesLists, n), removeZeroSeries(seriesLists, n)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		for i, v := range a.IsAbsent {
			if !v {
				if e.Target() == "removeEmptySeries" || (a.Values[i] != 0) {
					results = append(results, a)
					break
				}
			}
		}
	}
	return results, nil
}
