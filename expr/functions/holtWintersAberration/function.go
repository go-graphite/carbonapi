package holtWintersAberration

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
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

func (f *holtWintersAberration) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var results []*types.MetricData
	args, err := helper.GetSeriesArg(e.Args()[0], from-7*86400, until, values)
	if err != nil {
		return nil, err
	}

	delta, err := e.GetFloatNamedOrPosArgDefault("delta", 1, 3)
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		var aberration []float64

		stepTime := arg.StepTime

		lowerBand, upperBand := holtwinters.HoltWintersConfidenceBands(arg.Values, stepTime, delta)

		windowPoints := 7 * 86400 / stepTime
		series := arg.Values[windowPoints:]
		absent := arg.IsAbsent[windowPoints:]

		for i := range series {
			if absent[i] {
				aberration = append(aberration, 0)
			} else if !math.IsNaN(upperBand[i]) && series[i] > upperBand[i] {
				aberration = append(aberration, series[i]-upperBand[i])
			} else if !math.IsNaN(lowerBand[i]) && series[i] < lowerBand[i] {
				aberration = append(aberration, series[i]-lowerBand[i])
			} else {
				aberration = append(aberration, 0)
			}
		}

		r := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("holtWintersAberration(%s)", arg.Name),
			Values:    aberration,
			IsAbsent:  make([]bool, len(aberration)),
			StepTime:  arg.StepTime,
			StartTime: arg.StartTime + 7*86400,
			StopTime:  arg.StopTime,
		}}

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
