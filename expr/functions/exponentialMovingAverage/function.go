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
		windowSize int
		argstr     string
		err        error
	)

	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	switch e.Arg(1).Type() {
	case parser.EtConst:
		// In this case, zipper does not request additional retrospective points,
		// and leading `n` values, that used to calculate window, become NaN
		windowSize, err = e.GetIntArg(1)
		argstr = strconv.Itoa(windowSize)
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		if err != nil {
			return nil, err
		}
		argstr = strconv.Quote(e.Arg(1).StringValue())
		windowSize = int(n32)
	default:
		err = parser.ErrBadType
	}
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	if windowSize < 1 {
		return nil, fmt.Errorf("invalid window size %d", windowSize)
	}

	windowStr := strconv.Itoa(windowSize)

	start := from

	arg, err := helper.GetSeriesArg(ctx, e.Arg(0), start, until, values)
	if err != nil {
		return nil, err
	}
	if len(arg) == 0 {
		return results, nil
	}

	constant := float64(2 / (float64(windowSize) + 1))

	for _, a := range arg {
		r := a.CopyLink()
		r.Name = e.Target() + "(" + a.Name + "," + argstr + ")"

		var vals []float64

		if windowSize > len(a.Values) {
			mean := consolidations.AggMean(a.Values)
			vals = append(vals, helper.SafeRound(mean, 6))
		} else {
			ema := consolidations.AggMean(a.Values[:windowSize])

			vals = append(vals, helper.SafeRound(ema, 6))
			for _, v := range a.Values[windowSize:] {
				if math.IsNaN(v) {
					vals = append(vals, math.NaN())
					continue
				}
				ema = constant*v + (1-constant)*ema
				vals = append(vals, helper.SafeRound(ema, 6))
			}
		}

		r.Tags[e.Target()] = windowStr
		r.Values = vals
		r.StartTime = (from + r.StepTime - 1) / r.StepTime * r.StepTime // align StartTime to closest >= StepTime
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
