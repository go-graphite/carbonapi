package holtWinters

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
	metadata.RegisterFunction("holtWintersForecast", &Forecast{})
	metadata.RegisterFunction("holtWintersConfidenceBands", &ConfidenceBands{})
	metadata.RegisterFunction("holtWintersAberration", &Aberration{})

}

type Forecast struct {
	interfaces.FunctionBase
}

func (f *Forecast) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
			Name:      fmt.Sprintf("holtWintersForecast(%s)", arg.Name),
			Values:    predictionsOfInterest,
			IsAbsent:  make([]bool, len(predictionsOfInterest)),
			StepTime:  arg.StepTime,
			StartTime: arg.StartTime + 7*86400,
			StopTime:  arg.StopTime,
		}}

		results = append(results, &r)
	}
	return results, nil

}

type ConfidenceBands struct {
	interfaces.FunctionBase
}

func (f *ConfidenceBands) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

type Aberration struct {
	interfaces.FunctionBase
}

func (f *Aberration) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
