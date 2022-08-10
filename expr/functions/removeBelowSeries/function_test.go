package removeBelowSeries

import (
	"math"
	"testing"
	"time"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
	th "github.com/grafana/carbonapi/tests"
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
			"removeBelowValue(metric1, 0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("removeBelowValue(metric1,0)",
				[]float64{1, 2, math.NaN(), 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"removeAboveValue(metric1, 10)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("removeAboveValue(metric1,10)",
				[]float64{1, 2, -1, 7, 8, math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
		},
		{
			"removeBelowPercentile(metric1, 50)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("removeBelowPercentile(metric1,50)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"removeAbovePercentile(metric1, 50)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("removeAbovePercentile(metric1,50)",
				[]float64{1, 2, -1, 7, math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
