package removeEmptySeries

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
		{
			"removeEmptySeries(metric*,0.00001)",
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
			"removeZeroSeries(metric*,0.000001)",
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
		{
			"removeEmptySeries(metric*,0.8)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, -1, 7, 8, 20, 23, math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric4", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric5", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
				types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32),
				types.MakeMetricData("metric5", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			"removeEmptySeries(metric*,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32),
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
