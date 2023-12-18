package integralByInterval

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type integralByInterval struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &integralByInterval{}
	functions := []string{"integralByInterval"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// integralByInterval(seriesList, intervalString)
func (f *integralByInterval) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, nil
	}

	bucketSizeInt32, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}
	bucketSize := int64(bucketSizeInt32)
	intervalString, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	startTime := from
	results := make([]*types.MetricData, len(args))
	for j, arg := range args {
		current := 0.0
		currentTime := arg.StartTime

		name := "integralByInterval(" + arg.Name + ",'" + intervalString + "')"
		result := arg.CopyLink()
		result.Name = name
		result.PathExpression = name
		result.Values = make([]float64, len(arg.Values))

		result.Tags["integralByInterval"] = intervalString

		for i, v := range arg.Values {
			if (currentTime-startTime)/bucketSize != (currentTime-startTime-arg.StepTime)/bucketSize {
				current = 0
			}
			if math.IsNaN(v) {
				v = 0
			}
			current += v
			result.Values[i] = current
			currentTime += arg.StepTime
		}

		results[j] = result
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *integralByInterval) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"integralByInterval": {
			Description: "This will do the same as integralByInterval() funcion, except resetting the total to 0 at the given time in the parameter “from” Useful for finding totals per hour/day/week/..",
			Function:    "integralByInterval(seriesList, intervalString)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "integralByInterval",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				}, {
					Name:     "intervalString",
					Required: true,
					Suggestions: types.NewSuggestions(
						"10min",
						"1h",
						"1d",
					),
					Type: types.Interval,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
