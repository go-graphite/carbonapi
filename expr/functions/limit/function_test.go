package limit

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

func TestLimit(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"limit(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{0, 0, 1, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricC", []float64{0, 0, 0, 1, 0, 0}, 1, now32),
					types.MakeMetricData("metricD", []float64{0, 0, 0, 0, 1, 0}, 1, now32),
					types.MakeMetricData("metricE", []float64{0, 0, 0, 0, 0, 1}, 1, now32),
				},
			},
			"limit",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32)},
				"metricB": {types.MakeMetricData("metricB", []float64{0, 0, 1, 0, 0, 0}, 1, now32)},
			},
		},
		{
			"limit(metric1,20)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{0, 0, 1, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricC", []float64{0, 0, 0, 1, 0, 0}, 1, now32),
					types.MakeMetricData("metricD", []float64{0, 0, 0, 0, 1, 0}, 1, now32),
					types.MakeMetricData("metricE", []float64{0, 0, 0, 0, 0, 1}, 1, now32),
				},
			},
			"limit",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{0, 1, 0, 0, 0, 0}, 1, now32)},
				"metricB": {types.MakeMetricData("metricB", []float64{0, 0, 1, 0, 0, 0}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{0, 0, 0, 1, 0, 0}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{0, 0, 0, 0, 1, 0}, 1, now32)},
				"metricE": {types.MakeMetricData("metricE", []float64{0, 0, 0, 0, 0, 1}, 1, now32)},
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
