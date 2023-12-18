package pearsonClosest

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
			"pearsonClosest(metric1,metric2,1,direction=\"abs\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricX", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
				{"metric2", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, math.NaN(), 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB",
				[]float64{3, math.NaN(), 5, 6, 7, 8}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func TestFunctionMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"pearsonClosest(metricC,metric*,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricD", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricE", []float64{4, 7, 7, 7, 7, 1}, 1, now32),
				},
				{"metricC", 0, 1}: {
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			"pearsonClosest",
			map[string][]*types.MetricData{
				"metricC": {types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{4, 4, 5, 5, 6, 6}, 1, now32)},
			},
		},
		{
			"pearsonClosest(metricC,metric*,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricD", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricE", []float64{4, 7, 7, 7, 7, 1}, 1, now32),
				},
				{"metricC", 0, 1}: {
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			"pearsonClosest",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{4, 4, 5, 5, 6, 6}, 1, now32)},
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
