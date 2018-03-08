package multiplySeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
)

type multiplySeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &multiplySeries{}
	functions := []string{"multiplySeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// multiplySeries(factorsSeriesList)
func (f *multiplySeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	r := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("multiplySeries(%s)", e.RawArgs()),
			StartTime: from,
			StopTime:  until,
		},
	}
	for _, arg := range e.Args() {
		series, err := helper.GetSeriesArg(arg, from, until, values)
		if err != nil {
			return nil, err
		}

		if r.Values == nil {
			r.IsAbsent = make([]bool, len(series[0].IsAbsent))
			r.StepTime = series[0].StepTime
			r.Values = make([]float64, len(series[0].Values))
			copy(r.IsAbsent, series[0].IsAbsent)
			copy(r.Values, series[0].Values)
			series = series[1:]
		}

		for _, factor := range series {
			for i, v := range r.Values {
				if r.IsAbsent[i] || factor.IsAbsent[i] {
					r.IsAbsent[i] = true
					r.Values[i] = math.NaN()
					continue
				}

				r.Values[i] = v * factor.Values[i]
			}
		}
	}

	return []*types.MetricData{&r}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *multiplySeries) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"multiplySeries": {
			Description: "Takes two or more series and multiplies their points. A constant may not be\nused. To multiply by a constant, use the scale() function.\n\nExample:\n\n.. code-block:: none\n\n  &target=multiplySeries(Series.dividends,Series.divisors)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``multiply``.",
			Function:    "multiplySeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "multiplySeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
