package holtWintersConfidenceBands

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
)

func init() {
	metadata.RegisterFunction("holtWintersConfidenceBands", &holtWintersConfidenceBands{})

}

type holtWintersConfidenceBands struct {
	interfaces.FunctionBase
}

func (f *holtWintersConfidenceBands) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		stepTime := arg.StepTime

		lowerBand, upperBand := holtwinters.HoltWintersConfidenceBands(arg.Values, stepTime, delta)

		lowerSeries := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("holtWintersConfidenceLower(%s)", arg.Name),
			Values:    lowerBand,
			IsAbsent:  make([]bool, len(lowerBand)),
			StepTime:  arg.StepTime,
			StartTime: arg.StartTime + 7*86400,
			StopTime:  arg.StopTime,
		}}

		for i, val := range lowerSeries.Values {
			if math.IsNaN(val) {
				lowerSeries.Values[i] = 0
				lowerSeries.IsAbsent[i] = true
			}
		}

		upperSeries := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("holtWintersConfidenceUpper(%s)", arg.Name),
			Values:    upperBand,
			IsAbsent:  make([]bool, len(upperBand)),
			StepTime:  arg.StepTime,
			StartTime: arg.StartTime + 7*86400,
			StopTime:  arg.StopTime,
		}}

		for i, val := range upperSeries.Values {
			if math.IsNaN(val) {
				upperSeries.Values[i] = 0
				upperSeries.IsAbsent[i] = true
			}
		}

		results = append(results, &lowerSeries)
		results = append(results, &upperSeries)
	}
	return results, nil

}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersConfidenceBands) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
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
					Default: "3",
					Name:    "delta",
					Type:    types.Integer,
				},
				{
					Default: "7d",
					Name:    "bootstrapInterval",
					Suggestions: []string{
						"7d",
						"30d",
					},
					Type: types.Interval,
				},
			},
		},
	}
}
