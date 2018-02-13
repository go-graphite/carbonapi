package isNotNull

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("isNotNull", &Function{})
	metadata.RegisterFunction("isNonNull", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// isNonNull(seriesList)
// alias: isNotNull(seriesList)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	e.SetTarget("isNonNull")

	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		for i := range a.Values {
			r.IsAbsent[i] = false
			if a.IsAbsent[i] {
				r.Values[i] = 0
			} else {
				r.Values[i] = 1
			}

		}
		return r
	})
}
