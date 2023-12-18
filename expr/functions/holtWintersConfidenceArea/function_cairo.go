//go:build cairo
// +build cairo

package holtWintersConfidenceArea

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type holtWintersConfidenceArea struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &holtWintersConfidenceArea{}
	functions := []string{"holtWintersConfidenceArea"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *holtWintersConfidenceArea) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	bootstrapInterval, err := e.GetIntervalNamedOrPosArgDefault("bootstrapInterval", 2, 1, holtwinters.DefaultBootstrapInterval)
	if err != nil {
		return nil, err
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from-bootstrapInterval, until, values)
	if err != nil {
		return nil, err
	}

	delta, err := e.GetFloatNamedOrPosArgDefault("delta", 1, 3)
	if err != nil {
		return nil, err
	}

	seasonality, err := e.GetIntervalNamedOrPosArgDefault("seasonality", 3, 1, holtwinters.DefaultSeasonality)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args)*2)
	for _, arg := range args {
		stepTime := arg.StepTime

		lowerBand, upperBand := holtwinters.HoltWintersConfidenceBands(arg.Values, stepTime, delta, bootstrapInterval, seasonality)

		lowerSeries := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              fmt.Sprintf("holtWintersConfidenceArea(%s)", arg.Name),
				Values:            lowerBand,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				ConsolidationFunc: arg.ConsolidationFunc,
				XFilesFactor:      arg.XFilesFactor,
				PathExpression:    fmt.Sprintf("holtWintersConfidenceArea(%s)", arg.Name),
			},
			Tags: helper.CopyTags(arg),
			GraphOptions: types.GraphOptions{
				Stacked:   true,
				StackName: types.DefaultStackName,
				Invisible: true,
			},
		}
		lowerSeries.Tags["holtWintersConfidenceArea"] = "1"

		upperSeries := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              fmt.Sprintf("holtWintersConfidenceArea(%s)", arg.Name),
				Values:            upperBand,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				ConsolidationFunc: arg.ConsolidationFunc,
				XFilesFactor:      arg.XFilesFactor,
				PathExpression:    fmt.Sprintf("holtWintersConfidenceArea(%s)", arg.Name),
			},
			Tags: helper.CopyTags(arg),
			GraphOptions: types.GraphOptions{
				Stacked:   true,
				StackName: types.DefaultStackName,
			},
		}

		upperSeries.Tags["holtWintersConfidenceArea"] = "1"

		results = append(results, &lowerSeries, &upperSeries)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersConfidenceArea) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"holtWintersConfidenceArea": {
			Description: "Performs a Holt-Winters forecast using the series as input data and plots\n the area between the upper and lower bands of the predicted forecast deviations.",
			Function:    "holtWintersConfidenceArea(seriesList, delta=3, bootstrapInterval='7d')",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "holtWintersConfidenceArea",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(3),
					Name:    "delta",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion("7d"),
					Name:    "bootstrapInterval",
					Suggestions: types.NewSuggestions(
						"7d",
						"30d",
					),
					Type: types.Interval,
				},
			},
		},
	}
}
