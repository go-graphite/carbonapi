package minMax

import (
	"context"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type minMax struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &minMax{}
	functions := []string{"minMax"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)
func (f *minMax) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := a.CopyLinkTags()
		r.Name = fmt.Sprintf("minMax(%s)", a.Name)
		r.Values = make([]float64, len(a.Values))

		min := consolidations.MinValue(a.Values)
		if math.IsInf(min, 1) {
			min = 0.0
		}
		max := consolidations.MaxValue(a.Values)
		if math.IsInf(max, -1) {
			max = 0.0
		}

		for i, v := range a.Values {
			if !math.IsNaN(v) {
				if max != min {
					r.Values[i] = (v - min) / (max - min)
				} else {
					r.Values[i] = 0.0
				}
			} else {
				r.Values[i] = math.NaN()
			}
		}
		results = append(results, r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *minMax) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"minMax": {
			Description: "Applies the popular min max normalization technique, which takes each point and applies the following normalization transformation to it: normalized = (point - min) / (max - min).\n\nExample:\n\n.. code-block:: none\n\n  &target=minMax(Server.instance01.threads.busy)",
			Function:    "minMax(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "minMax",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
