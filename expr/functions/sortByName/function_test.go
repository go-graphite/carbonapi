package sortByName

import (
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
		{
			"sortByName(metric.foo.*)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // 100 is placed between 1 and 2 because it is alphabetical sort
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
			},
		},
		{
			"sortByName(metric.foo.*, true)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // "natural" sort method considers that metrics contain numbers
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
			},
		},
		{
			"sortByName(metric.foo.*, natural=false, reverse=true)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // alphabetical reverse sort
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
			},
		},
		{
			"sortByName(metric.foo.*, true, true)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				},
			},
			[]*types.MetricData{ // "natural" reverse sort
				types.MakeMetricData("metric.foo.x100", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x99", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x2", []float64{1}, 1, now32),
				types.MakeMetricData("metric.foo.x1", []float64{1}, 1, now32),
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
