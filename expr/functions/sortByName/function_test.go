package sortByName

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"sortByName(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricX", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{0, 0, 2, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricC", []float64{0, 0, 0, 3, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricB", []float64{0, 0, 2, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricC", []float64{0, 0, 0, 3, 0, 0}, 1, now32),
				types.MakeMetricData("metricX", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			"sortByName(metric*,natural=true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metric12", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metric1234567890", []float64{0, 0, 0, 5, 0, 0}, 1, now32),
					types.MakeMetricData("metric2", []float64{0, 0, 2, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metric11", []float64{0, 0, 0, 3, 0, 0}, 1, now32),
					types.MakeMetricData("metric", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metric1", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metric2", []float64{0, 0, 2, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metric11", []float64{0, 0, 0, 3, 0, 0}, 1, now32),
				types.MakeMetricData("metric12", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metric1234567890", []float64{0, 0, 0, 5, 0, 0}, 1, now32),
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

