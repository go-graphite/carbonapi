package sortByName

import (
	"context"
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

func TestSortByName(t *testing.T) {
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

func BenchmarkSortByName(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "sortByName(metric.foo.*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, 1),
				},
			},
		},
		{
			target: "sortByName(metric.foo.*, true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, 1),
				},
			},
		},
		{
			target: "sortByName(metric.foo.*, natural=false, reverse=true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, 1),
				},
			},
		},
		{
			target: "sortByName(metric.foo.*, true, true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "metric.foo.*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric.foo.x99", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x1", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x2", []float64{1}, 1, 1),
					types.MakeMetricData("metric.foo.x100", []float64{1}, 1, 1),
				},
			},
		},
	}

	evaluator := metadata.GetEvaluator()

	for _, bm := range benchmarks {
		b.Run(bm.target, func(b *testing.B) {
			exp, _, err := parser.ParseExpr(bm.target)
			if err != nil {
				b.Fatalf("failed to parse %s: %+v", bm.target, err)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				g, err := evaluator.Eval(context.Background(), exp, 0, 1, bm.M)
				if err != nil {
					b.Fatalf("failed to eval %s: %+v", bm.target, err)
				}
				_ = g
			}
		})
	}
}
