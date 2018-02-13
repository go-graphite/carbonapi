package delay

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("delay", &Delay{})
}

type Delay struct {
	interfaces.FunctionBase
}

// delay(seriesList, steps)
func (f *Delay) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	steps, err := e.GetIntArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, series := range seriesList {
		length := len(series.Values)

		newValues := make([]float64, length)
		newIsAbsents := make([]bool, length)
		var prevValues []float64
		var prevIsAbsent []bool

		for i, value := range series.Values {
			if len(prevValues) < steps {
				newValues[i] = 0
				newIsAbsents[i] = true
			} else {
				newValue := prevValues[0]
				newIsAbsent := prevIsAbsent[0]
				prevValues = prevValues[1:]
				prevIsAbsent = prevIsAbsent[1:]

				newValues[i] = newValue
				newIsAbsents[i] = newIsAbsent
			}

			prevValues = append(prevValues, value)
			prevIsAbsent = append(prevIsAbsent, series.IsAbsent[i])
		}

		result := *series
		result.Name = fmt.Sprintf("delay(%s,%d)", series.Name, steps)
		result.Values = newValues
		result.IsAbsent = newIsAbsents

		results = append(results, &result)
	}

	return results, nil
}
