package nonNegativeDerivative

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

func TestNonNegativeDerivative(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"nonNegativeDerivative(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 10, 14, 20}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("nonNegativeDerivative(metric1)", []float64{math.NaN(), 2, 2, 4, 4, 6}, 1, now32).SetTag("nonNegativeDerivative", "1")},
		},
		{
			"nonNegativeDerivative(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 1, 4, math.NaN(), 8}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("nonNegativeDerivative(metric1)", []float64{math.NaN(), 2, 2, math.NaN(), 3, math.NaN(), math.NaN()}, 1, now32).SetTag("nonNegativeDerivative", "1")},
		},
		{
			"nonNegativeDerivative(metric1,32)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 0, 10, 1, math.NaN(), 8, 40, 37}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("nonNegativeDerivative(metric1,32)", []float64{math.NaN(), 2, 29, 10, 24, math.NaN(), math.NaN(), 32, math.NaN()}, 1, now32).SetTag("nonNegativeDerivative", "1")},
		},
		{
			"nonNegativeDerivative(metric1,minValue=1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 2, 10, 1, math.NaN(), 8, 40, 37}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("nonNegativeDerivative(metric1,minValue=1)", []float64{math.NaN(), 2, 1, 8, 0, math.NaN(), math.NaN(), 32, 36}, 1, now32).SetTag("nonNegativeDerivative", "1")},
		},
		{
			"nonNegativeDerivative(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("nonNegativeDerivative(metric1)", []float64{}, 1, now32).SetTag("nonNegativeDerivative", "1")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
