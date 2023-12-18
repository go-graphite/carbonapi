package exponentialMovingAverage

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type exponentialMovingAverage struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &exponentialMovingAverage{}
	functions := []string{"exponentialMovingAverage"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *exponentialMovingAverage) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var (
		windowPoints   int
		previewSeconds int
		argstr         string
		err            error
		constant       float64
	)

	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	refetch := false
	switch e.Arg(1).Type() {
	case parser.EtConst:
		windowPoints, err = e.GetIntArg(1)
		argstr = strconv.Itoa(windowPoints)
		if windowPoints < 0 {
			// we only care about the absolute value
			windowPoints = windowPoints * -1
		}

		// When the window is an integer, we check the fetched data to get the
		// step, and use it to calculate the preview window, to then refetch the
		// data. The already fetched values are discarded.
		refetch = true
		var maxStep int64
		arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
		if err != nil || len(arg) == 0 {
			return arg, err
		}
		for _, a := range arg {
			if a.StepTime > maxStep {
				maxStep = a.StepTime
			}
		}
		previewSeconds = int(maxStep) * windowPoints
		constant = float64(2 / (float64(windowPoints) + 1))
	case parser.EtString:
		// When the window is a string, we already adjusted the fetch request using the preview window.
		// No need to refetch.
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		if err != nil {
			return nil, err
		}
		argstr = strconv.Quote(e.Arg(1).StringValue())
		previewSeconds = int(n32)
		if previewSeconds < 0 {
			// we only care about the absolute value
			previewSeconds = previewSeconds * -1
		}
		constant = float64(2 / (float64(previewSeconds) + 1))

	default:
		return nil, parser.ErrBadType
	}

	if previewSeconds < 1 {
		return nil, fmt.Errorf("invalid window size %s", e.Arg(1).StringValue())
	}
	from = from - int64(previewSeconds)
	if refetch {
		f.GetEvaluator().Fetch(ctx, []parser.Expr{e.Arg(0)}, from, until, values)
	}
	previewList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range previewList {
		r := a.CopyLink()
		r.Name = e.Target() + "(" + a.Name + "," + argstr + ")"
		if e.Arg(1).Type() == parser.EtString {
			// If the window is a string (time interval), we adjust depending on the step
			windowPoints = previewSeconds / int(a.StepTime)
		}

		vals := make([]float64, 0, len(a.Values)/windowPoints+1)

		if windowPoints > len(a.Values) {
			mean := consolidations.AggMean(a.Values)
			vals = append(vals, helper.SafeRound(mean, 6))
		} else {
			ema := consolidations.AggMean(a.Values[:windowPoints])
			if math.IsNaN(ema) {
				ema = 0
			}

			vals = append(vals, helper.SafeRound(ema, 6))
			for _, v := range a.Values[windowPoints:] {
				if math.IsNaN(v) {
					vals = append(vals, math.NaN())
					continue
				}
				ema = constant*v + (1-constant)*ema
				vals = append(vals, helper.SafeRound(ema, 6))
			}
		}

		r.Tags[e.Target()] = argstr
		r.Values = vals
		r.StartTime = r.StartTime + int64(previewSeconds)
		r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime
		results = append(results, r)
	}
	return results, nil
}

func (f *exponentialMovingAverage) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"exponentialMovingAverage": {
			Description: "Takes a series of values and a window size and produces an exponential moving average utilizing the following formula:\n\n ema(current) = constant * (Current Value) + (1 - constant) * ema(previous)\n The Constant is calculated as:\n constant = 2 / (windowSize + 1) \n The first period EMA uses a simple moving average for its value.\n Example:\n\n code-block:: none\n\n  &target=exponentialMovingAverage(*.transactions.count, 10) \n\n &target=exponentialMovingAverage(*.transactions.count, '-10s')",
			Function:    "exponentialMovingAverage(seriesList, windowSize)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "exponentialMovingAverage",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						0.1,
						0.5,
						0.7,
					),
					Type: types.Float,
				},
			},
		},
	}
}
