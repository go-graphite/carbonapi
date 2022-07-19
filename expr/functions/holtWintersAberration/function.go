package holtWintersAberration

import (
	"context"
	"fmt"
	"math"

	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/holtwinters"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type holtWintersAberration struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &holtWintersAberration{}
	functions := []string{"holtWintersAberration"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *holtWintersAberration) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	bootstrapInterval, err := e.GetIntervalNamedOrPosArgDefault("bootstrapInterval", 2, 1, 7*86400)
	if err != nil {
		return nil, err
	}

	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from-bootstrapInterval, until, values)
	if err != nil {
		return nil, err
	}

	delta, err := e.GetFloatNamedOrPosArgDefault("delta", 1, 3)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {
		var (
			aberration []float64
			series     []float64
		)

		stepTime := arg.StepTime

		lowerBand, upperBand := holtwinters.HoltWintersConfidenceBands(arg.Values, stepTime, delta, bootstrapInterval/86400)

		windowPoints := int(bootstrapInterval / stepTime)
		if len(arg.Values) > windowPoints {
			series = arg.Values[windowPoints:]
		}

		for i := range series {
			if math.IsNaN(series[i]) {
				aberration = append(aberration, 0)
			} else if !math.IsNaN(upperBand[i]) && series[i] > upperBand[i] {
				aberration = append(aberration, series[i]-upperBand[i])
			} else if !math.IsNaN(lowerBand[i]) && series[i] < lowerBand[i] {
				aberration = append(aberration, series[i]-lowerBand[i])
			} else {
				aberration = append(aberration, 0)
			}
		}

		r := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              fmt.Sprintf("holtWintersAberration(%s)", arg.Name),
				Values:            aberration,
				StepTime:          arg.StepTime,
				StartTime:         arg.StartTime + bootstrapInterval,
				StopTime:          arg.StopTime,
				PathExpression:    fmt.Sprintf("holtWintersAberration(%s)", arg.Name),
				ConsolidationFunc: arg.ConsolidationFunc,
				XFilesFactor:      arg.XFilesFactor,
			},
			Tags: arg.Tags,
		}
		r.Tags["holtWintersAberration"] = "1"
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersAberration) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"holtWintersAberration": {
			Description: "Performs a Holt-Winters forecast using the series as input data and plots the\npositive or negative deviation of the series data from the forecast.",
			Function:    "holtWintersAberration(seriesList, delta=3, bootstrapInterval='7d')",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "holtWintersAberration",
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
