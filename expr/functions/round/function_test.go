package round

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestRound(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"round(metric1, 3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 245}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1,3)", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 245}, 1, now32).SetTag("round", "3")},
		},
		{
			"round(metric1, 1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 245}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1,1)", []float64{0.5, 2.3, math.NaN(), 91.0, -524.8, 245}, 1, now32).SetTag("round", "1")},
		},
		{
			"round(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 1.5, 2.298, math.NaN(), 91.019, -524.82, 245}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1)", []float64{0, 2, 2, math.NaN(), 91, -525, 245}, 1, now32).SetTag("round", "0")},
		},
		{
			"round(metric1, -2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 275}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1,-2)", []float64{0, 0, math.NaN(), 100, -500, 300}, 1, now32).SetTag("round", "-2")},
		},
		{
			"round(metric1, precision=-2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 275}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1,-2)", []float64{0, 0, math.NaN(), 100, -500, 300}, 1, now32).SetTag("round", "-2")},
		},
		{
			"round(metric1, -10)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, 2.298, math.NaN(), 91.019, -524.82, 245}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("round(metric1,-10)", []float64{0, 0, math.NaN(), 0, 0, 0}, 1, now32).SetTag("round", "-10")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
