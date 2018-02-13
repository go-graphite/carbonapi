package divideSeries

import (
	"errors"
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("divideSeries", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// divideSeries(dividendSeriesList, divisorSeriesList)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if len(e.Args()) < 1 {
		return nil, parser.ErrMissingTimeseries
	}

	firstArg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var useMetricNames bool

	var numerators []*types.MetricData
	var denominator *types.MetricData
	if len(e.Args()) == 2 {
		useMetricNames = true
		numerators = firstArg
		denominators, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
		if err != nil {
			return nil, err
		}
		if len(denominators) != 1 {
			return nil, types.ErrWildcardNotAllowed
		}

		denominator = denominators[0]
	} else if len(firstArg) == 2 && len(e.Args()) == 1 {
		numerators = append(numerators, firstArg[0])
		denominator = firstArg[1]
	} else {
		return nil, errors.New("must be called with 2 series or a wildcard that matches exactly 2 series")
	}

	for _, numerator := range numerators {
		if numerator.StepTime != denominator.StepTime || len(numerator.Values) != len(denominator.Values) {
			return nil, errors.New(fmt.Sprintf("series %s must have the same length as %s", numerator.Name, denominator.Name))
		}
	}

	var results []*types.MetricData
	for _, numerator := range numerators {
		r := *numerator
		if useMetricNames {
			r.Name = fmt.Sprintf("divideSeries(%s,%s)", numerator.Name, denominator.Name)
		} else {
			r.Name = fmt.Sprintf("divideSeries(%s)", e.RawArgs())
		}
		r.Values = make([]float64, len(numerator.Values))
		r.IsAbsent = make([]bool, len(numerator.Values))

		for i, v := range numerator.Values {

			if numerator.IsAbsent[i] || denominator.IsAbsent[i] || denominator.Values[i] == 0 {
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v / denominator.Values[i]
		}
		results = append(results, &r)
	}

	return results, nil

}
