package delay

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type delay struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &delay{}
	functions := []string{"delay"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// delay(seriesList, steps)
func (f *delay) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	seriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	steps, err := e.GetIntArg(1)
	if err != nil {
		return nil, err
	}
	stepsStr := e.Arg(1).StringValue()

	results := make([]*types.MetricData, len(seriesList))

	for i, series := range seriesList {
		length := len(series.Values)

		result := series.CopyName("delay(" + series.Name + "," + stepsStr + ")")
		result.Tags["delay"] = stepsStr

		var newValues []float64
		if steps < 0 {
			newValues = make([]float64, length)
			copy(newValues, series.Values[-steps:])
			for i := -steps; i < length; i++ {
				newValues[i] = math.NaN()
			}
			result.Values = newValues
		} else if steps != 0 {
			newValues = make([]float64, length)
			for i := 0; i < steps; i++ {
				newValues[i] = math.NaN()
			}
			copy(newValues[steps:], series.Values[:length-steps])
			result.Values = newValues
		}

		results[i] = result
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *delay) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"delay": {
			Description: "This shifts all samples later by an integer number of steps. This can be\nused for custom derivative calculations, among other things. Note: this\nwill pad the early end of the data with None for every step shifted.\n\nThis complements other time-displacement functions such as timeShift and\ntimeSlice, in that this function is indifferent about the step intervals\nbeing shifted.\n\nExample:\n\n.. code-block:: none\n\n  &target=divideSeries(server.FreeSpace,delay(server.FreeSpace,1))\n\nThis computes the change in server free space as a percentage of the previous\nfree space.",
			Function:    "delay(seriesList, steps)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "delay",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "steps",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
