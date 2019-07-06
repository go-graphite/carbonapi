package removeEmptySeries

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"removeEmptySeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
				types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			"removeZeroSeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
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

