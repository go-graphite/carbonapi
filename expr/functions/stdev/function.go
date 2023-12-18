package stdev

import (
	"context"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
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
	res = append(res, interfaces.FunctionMetadata{Name: "stdev", F: f})
	return res
}

// stdev(seriesList, points, missingThreshold=0.1)
// Alias: stddev
func (f *stdev) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	points, err := e.GetIntArg(1)
	if err != nil {
		return nil, err
	}
	pointsStr := e.Arg(1).StringValue()

	missingThreshold, err := e.GetFloatArgDefault(2, 0.1)
	if err != nil {
		return nil, err
	}

	minLen := int((1 - missingThreshold) * float64(points))

	result := make([]*types.MetricData, len(arg))

	for n, a := range arg {
		w := &types.Windowed{Data: make([]float64, points)}

		r := a.CopyLink()
		r.Name = "stdev(" + a.Name + "," + pointsStr + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["stdev"] = fmt.Sprintf("%d", points)

		for i, v := range a.Values {
			w.Push(v)
			if !math.IsNaN(v) {
				r.Values[i] = w.Stdev()
			} else {
				r.Values[i] = math.NaN()
			}
			if math.IsNaN(r.Values[i]) || (i >= minLen && w.Len() < minLen) {
				r.Values[i] = math.NaN()
			}
		}
		result[n] = r
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
	}
}
