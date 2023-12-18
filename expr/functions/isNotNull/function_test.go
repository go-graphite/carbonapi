package isNotNull

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"isNonNull(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("isNonNull(metric1)",
				[]float64{0, 1, 0, 1, 1, 1}, 1, now32).SetTag("isNonNull", "1").SetNameTag("isNonNull(metric1)")},
		},
		{
			"isNonNull(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricFoo", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32),
					types.MakeMetricData("metricBaz", []float64{1, -1, math.NaN(), -3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("isNonNull(metricFoo)", []float64{0, 1, 0, 1, 1, 1}, 1, now32).SetTag("isNonNull", "1").SetNameTag("isNonNull(metricFoo)"),
				types.MakeMetricData("isNonNull(metricBaz)", []float64{1, 1, 0, 1, 1, 1}, 1, now32).SetTag("isNonNull", "1").SetNameTag("isNonNull(metricBaz)"),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
