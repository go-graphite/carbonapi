package lowPass

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("lowPass", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// lowPass(seriesList, cutPercent)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		fmt.Printf("lowPass failed: 1\n")
		return nil, err
	}

	cutPercent, err := e.GetFloatArg(1)
	if err != nil {
		fmt.Printf("lowPass failed: 2\n")
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range arg {
		name := fmt.Sprintf("lowPass(%s,%v)", a.Name, cutPercent)
		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		lowCut := int((cutPercent / 200) * float64(len(a.Values)))
		highCut := len(a.Values) - lowCut
		for i, v := range a.Values {
			if i < lowCut || i >= highCut {
				r.Values[i] = v
			} else {
				r.IsAbsent[i] = true
			}
		}

		results = append(results, &r)
	}
	return results, nil
}
