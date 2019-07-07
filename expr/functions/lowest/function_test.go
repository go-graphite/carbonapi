package lowest

import (
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

func TestLowestMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"lowestCurrent(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
					types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				},
			},
			"lowestCurrent",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32)},
				"metricB": {types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32)},
			},
		},
		{
			"lowestCurrent(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
					types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				},
			},
			"lowestCurrent",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}
}

func TestLowest(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"lowestCurrent(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 1, 3, 3, 4, 1}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
