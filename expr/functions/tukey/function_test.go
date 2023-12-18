package tukey

import (
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

func TestFunctionMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"tukeyAbove(metric*,1.5,5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			"tukeyAbove",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32)},
			},
		},
		{
			"tukeyAbove(metric*, 3, 5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			"tukeyAbove",
			map[string][]*types.MetricData{
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32)},
			},
		},
		{
			"tukeyAbove(metric*, 1.5, 5, 6)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{20, 20, 20, 20, 21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 20, 20, 20, 20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{20, 20, 20, 20, 19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{20, 20, 20, 20, 18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{20, 20, 20, 20, 17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			"tukeyAbove(metric*, 1.5, 5, 6)",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{20, 20, 20, 20, 21, 17, 20, 20, 10, 29}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{20, 20, 20, 20, 18, 20, 22, 14, 26, 20}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{20, 20, 20, 20, 17, 21, 8, 30, 18, 28}, 1, now32)},
			},
		},
		{
			"tukeyAbove(metric*,1.5,5,\"6s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{20, 20, 20, 20, 21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 20, 20, 20, 20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{20, 20, 20, 20, 19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{20, 20, 20, 20, 18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{20, 20, 20, 20, 17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			`tukeyAbove(metric*, 1.5, 5, "6s")`,
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{20, 20, 20, 20, 21, 17, 20, 20, 10, 29}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{20, 20, 20, 20, 18, 20, 22, 14, 26, 20}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{20, 20, 20, 20, 17, 21, 8, 30, 18, 28}, 1, now32)},
			},
		},
		{
			"tukeyBelow(metric*,1.5,5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			"tukeyBelow",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32)},
			},
		},
		{
			"tukeyBelow(metric*,1.5,5,-4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28, 20, 20, 20, 20}, 1, now32),
				},
			},

			"tukeyBelow",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29, 20, 20, 20, 20}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28, 20, 20, 20, 20}, 1, now32)},
			},
		},
		{
			"tukeyBelow(metric*,1.5,5,\"-4s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20, 20, 20, 20, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28, 20, 20, 20, 20}, 1, now32),
				},
			},

			"tukeyBelow",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29, 20, 20, 20, 20}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28, 20, 20, 20, 20}, 1, now32)},
			},
		},
		{
			"tukeyBelow(metric*,3,5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{21, 17, 20, 20, 10, 29}, 1, now32),
					types.MakeMetricData("metricB", []float64{20, 18, 21, 19, 20, 20}, 1, now32),
					types.MakeMetricData("metricC", []float64{19, 19, 21, 17, 23, 20}, 1, now32),
					types.MakeMetricData("metricD", []float64{18, 20, 22, 14, 26, 20}, 1, now32),
					types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32),
				},
			},

			"tukeyBelow",
			map[string][]*types.MetricData{
				"metricE": {types.MakeMetricData("metricE", []float64{17, 21, 8, 30, 18, 28}, 1, now32)},
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
