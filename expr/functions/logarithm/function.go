package logarithm

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
	metadata.RegisterFunction("logarithm", &Function{})
	metadata.RegisterFunction("log", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// logarithm(seriesList, base=10)
// Alias: log
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	base, err := e.GetIntNamedOrPosArgDefault("base", 1, 10)
	if err != nil {
		return nil, err
	}
	_, ok := e.NamedArgs()["base"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	baseLog := math.Log(float64(base))

	var results []*types.MetricData

	for _, a := range arg {

		var name string
		if ok {
			name = fmt.Sprintf("logarithm(%s,%d)", a.Name, base)
		} else {
			name = fmt.Sprintf("logarithm(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = math.Log(v) / baseLog
		}
		results = append(results, &r)
	}
	return results, nil
}
