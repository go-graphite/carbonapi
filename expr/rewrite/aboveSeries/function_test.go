package aboveSeries

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
	evaluator := th.DummyEvaluator()
	helper.SetEvaluator(evaluator)
	metadata.SetEvaluator(evaluator)

	md := New("")
	for _, m := range md {
		metadata.RegisterRewriteFunction(m.Name, m.F)
	}
}

func TestDiffSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.RewriteTestItem{
		{
			`aboveSeries(metric1, 7, "Kotik", "Bog")`,
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {
					types.MakeMetricData("metricSobaka", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricKotik", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricHomyak", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			th.RewriteTestResult{
				Rewritten: true,
				Targets:   []string{"metricBog"},
				Err:       nil,
			},
		},
		{
			`aboveSeries(metric1, 7, ".*Ko.ik$", "Bog")`,
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {
					types.MakeMetricData("metricSobaka", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricKotik", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricHomyak", []float64{4, 4, 5, 5, 6, 8}, 1, now32),
				},
			},
			th.RewriteTestResult{
				Rewritten: true,
				Targets:   []string{"Bog", "metricHomyak"},
				Err:       nil,
			},
		},
		{
			`aboveSeries(statsd.timers.metric.rate, 1000, 'rate', 'median')`,
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "statsd.timers.metric.rate", From: 0, Until: 1}: {
					types.MakeMetricData("statsd.timers.metric.rate", []float64{500, 1500}, 1, now32),
					types.MakeMetricData("statsd.timers.metric.median", []float64{50, 55}, 1, now32),
				},
			},
			th.RewriteTestResult{
				Rewritten: true,
				Targets:   []string{"statsd.timers.metric.median"},
				Err:       nil,
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestRewriteExpr(t, &tt)
		})
	}

}
