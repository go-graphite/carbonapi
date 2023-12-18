package holtWintersForecast

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type holtWintersForecast struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &holtWintersForecast{}
	functions := []string{"holtWintersForecast"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *holtWintersForecast) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	bootstrapInterval, err := e.GetIntervalNamedOrPosArgDefault("bootstrapInterval", 1, 1, holtwinters.DefaultBootstrapInterval)
	if err != nil {
		return nil, err
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from-bootstrapInterval, until, values)
	if err != nil {
		return nil, err
	}

	seasonality, err := e.GetIntervalNamedOrPosArgDefault("seasonality", 2, 1, holtwinters.DefaultSeasonality)
	if err != nil {
		return nil, err
	}

	var predictionsOfInterest []float64
	results := make([]*types.MetricData, len(args))
	for i, arg := range args {
		stepTime := arg.StepTime

		predictions, _ := holtwinters.HoltWintersAnalysis(arg.Values, stepTime, seasonality)

		windowPoints := int(bootstrapInterval / stepTime)
		if len(predictions) < windowPoints {
			predictionsOfInterest = predictions
		} else {
			predictionsOfInterest = predictions[windowPoints:]
		}

		name := "holtWintersForecast(" + arg.Name + ")"
		r := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				Values:            predictionsOfInterest,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				PathExpression:    name,
				XFilesFactor:      arg.XFilesFactor,
				ConsolidationFunc: arg.ConsolidationFunc,
			},
			Tags: helper.CopyTags(arg),
		}
		r.Tags["holtWintersForecast"] = "1"
		results[i] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersForecast) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"holtWintersForecast": {
			Description: "Performs a Holt-Winters forecast using the series as input data. Data from\n`bootstrapInterval` (one week by default) previous to the series is used to bootstrap the initial forecast.",
			Function:    "holtWintersForecast(seriesList, bootstrapInterval='7d')",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "holtWintersForecast",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
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
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
