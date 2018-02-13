package keepLastValue

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
	metadata.RegisterFunction("keepLastValue", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// keepLastValue(seriesList, limit=inf)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	keep, err := e.GetIntNamedOrPosArgDefault("limit", 1, -1)
	if err != nil {
		return nil, err
	}
	_, ok := e.NamedArgs()["limit"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	var results []*types.MetricData

	for _, a := range arg {
		var name string
		if ok {
			name = fmt.Sprintf("keepLastValue(%s,%d)", a.Name, keep)
		} else {
			name = fmt.Sprintf("keepLastValue(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		prev := math.NaN()
		missing := 0

		for i, v := range a.Values {
			if a.IsAbsent[i] {

				if (keep < 0 || missing < keep) && !math.IsNaN(prev) {
					r.Values[i] = prev
					missing++
				} else {
					r.IsAbsent[i] = true
				}

				continue
			}
			missing = 0
			prev = v
			r.Values[i] = v
		}
		results = append(results, &r)
	}
	return results, err
}
