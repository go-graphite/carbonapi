package holtWintersForecast

import (
	"fmt"
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

func (f *holtWintersForecast) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var results []*types.MetricData
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from-7*86400, until, values)
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		stepTime := arg.StepTime

		predictions, _ := holtwinters.HoltWintersAnalysis(arg.Values, stepTime)

		windowPoints := 7 * 86400 / stepTime
		predictionsOfInterest := predictions[windowPoints:]

		r := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:              fmt.Sprintf("holtWintersForecast(%s)", arg.Name),
			Values:            predictionsOfInterest,
			StepTime:          arg.StepTime,
			StartTime:         arg.StartTime + 7*86400,
			StopTime:          arg.StopTime,
			PathExpression:    fmt.Sprintf("holtWintersForecast(%s)", arg.Name),
			XFilesFactor:      arg.XFilesFactor,
			ConsolidationFunc: arg.ConsolidationFunc,
		}}

		results = append(results, &r)
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
		},
	}
}
