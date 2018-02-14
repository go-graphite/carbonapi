package transformNull

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &function{}
	functions := []string{"transformNull"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// transformNull(seriesList, default=0)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	defv, err := e.GetFloatNamedOrPosArgDefault("default", 1, 0)
	if err != nil {
		return nil, err
	}

	_, ok := e.NamedArgs()["default"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	var results []*types.MetricData

	for _, a := range arg {

		var name string
		if ok {
			name = fmt.Sprintf("transformNull(%s,%g)", a.Name, defv)
		} else {
			name = fmt.Sprintf("transformNull(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				v = defv
			}

			r.Values[i] = v
		}

		results = append(results, &r)
	}
	return results, nil
}
