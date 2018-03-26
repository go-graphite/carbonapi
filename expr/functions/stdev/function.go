package stdev

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type stdev struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &stdev{}
	functions := []string{"stdev", "stddev"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// stdev(seriesList, points, missingThreshold=0.1)
// Alias: stddev
func (f *stdev) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	points, err := e.GetIntArg(1)
	if err != nil {
		return nil, err
	}

	missingThreshold, err := e.GetFloatArgDefault(2, 0.1)
	if err != nil {
		return nil, err
	}

	minLen := int((1 - missingThreshold) * float64(points))

	var result []*types.MetricData

	for _, a := range arg {
		w := &types.Windowed{Data: make([]float64, points)}

		r := *a
		r.Name = fmt.Sprintf("stdev(%s,%d)", a.Name, points)
		r.Values = make([]float64, len(a.Values))

		for i, v := range a.Values {
			w.Push(v)
			r.Values[i] = w.Stdev()
			if math.IsNaN(r.Values[i]) || (i >= minLen && w.Len() < minLen) {
				r.Values[i] = math.NaN()
			}
		}
		result = append(result, &r)
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *stdev) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"stdev": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nDraw the Standard Deviation of all metrics passed for the past N datapoints.\nIf the ratio of null points in the window is greater than windowTolerance,\nskip the calculation. The default for windowTolerance is 0.1 (up to 10% of points\nin the window can be missing). Note that if this is set to 0.0, it will cause large\ngaps in the output anywhere a single point is missing.\n\nExample:\n\n.. code-block:: none\n\n  &target=stdev(server*.instance*.threads.busy,30)\n  &target=stdev(server*.instance*.cpu.system,30,0.0)",
			Function:    "stdev(seriesList, points, windowTolerance=0.1)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "stdev",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "points",
					Required: true,
					Type:     types.Integer,
				},
				{
					Default: types.NewSuggestion(0.1),
					Name:    "windowTolerance",
					Type:    types.Float,
				},
			},
		},
		"stddev": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nDraw the Standard Deviation of all metrics passed for the past N datapoints.\nIf the ratio of null points in the window is greater than windowTolerance,\nskip the calculation. The default for windowTolerance is 0.1 (up to 10% of points\nin the window can be missing). Note that if this is set to 0.0, it will cause large\ngaps in the output anywhere a single point is missing.\n\nExample:\n\n.. code-block:: none\n\n  &target=stdev(server*.instance*.threads.busy,30)\n  &target=stdev(server*.instance*.cpu.system,30,0.0)",
			Function:    "stdev(seriesList, points, windowTolerance=0.1)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "stdev",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "points",
					Required: true,
					Type:     types.Integer,
				},
				{
					Default: types.NewSuggestion(0.1),
					Name:    "windowTolerance",
					Type:    types.Float,
				},
			},
		},
	}
}
