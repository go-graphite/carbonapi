package holtWintersConfidenceBands

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type holtWintersConfidenceBands struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &holtWintersConfidenceBands{}
	functions := []string{"holtWintersConfidenceBands"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *holtWintersConfidenceBands) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

		name := "holtWintersConfidenceLower(" + arg.Name + ")"
		lowerSeries := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				Values:            lowerBand,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				ConsolidationFunc: arg.ConsolidationFunc,
				XFilesFactor:      arg.XFilesFactor,
				PathExpression:    name,
			},
			Tags: helper.CopyTags(arg),
		}
		lowerSeries.Tags["holtWintersConfidenceLower"] = "1"

		name = "holtWintersConfidenceUpper(" + arg.Name + ")"
		upperSeries := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				Values:            upperBand,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				ConsolidationFunc: arg.ConsolidationFunc,
				XFilesFactor:      arg.XFilesFactor,
				PathExpression:    name,
			},
			Tags: helper.CopyTags(arg),
		}
		upperSeries.Tags["holtWintersConfidenceUpper"] = "1"

		results = append(results, lowerSeries, upperSeries)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersConfidenceBands) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"holtWintersConfidenceBands": {
			Description: "Performs a Holt-Winters forecast using the series as input data and plots\nupper and lower bands with the predicted forecast deviations.",
			Function:    "holtWintersConfidenceBands(seriesList, delta=3, bootstrapInterval='7d')",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "holtWintersConfidenceBands",
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
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
