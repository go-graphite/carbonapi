package changed

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
	metadata.RegisterFunction("changed", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// changed(SeriesList)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var result []*types.MetricData
	for _, a := range args {
		r := *a
		r.Name = fmt.Sprintf("%s(%s)", e.Target(), a.Name)
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := math.NaN()
		for i, v := range a.Values {
			if math.IsNaN(prev) {
				prev = v
				r.Values[i] = 0
			} else if !math.IsNaN(v) && prev != v {
				r.Values[i] = 1
				prev = v
			} else {
				r.Values[i] = 0
			}
		}
		result = append(result, &r)
	}
	return result, nil
}
