package ewma

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"math"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestEWMA(t *testing.T) {
	now32 := uint32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("ewma",
				"metric1",
				0.9,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 1, 1, 1, math.NaN(), 1, 1}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("ewma(metric1,0.9)", []float64{0, 0.9, 0.99, 0.999, math.NaN(), 0.9999, 0.99999}, 1, now32),
			},
		},
		{
			parser.NewExpr("exponentialWeightedMovingAverage",
				"metric1",
				0.9,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 1, 1, 1, math.NaN(), 1, 1}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("ewma(metric1,0.9)", []float64{0, 0.9, 0.99, 0.999, math.NaN(), 0.9999, 0.99999}, 1, now32),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
