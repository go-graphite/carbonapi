package sumSeriesWithWildcards

import (
	"context"
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

func TestFunctionSumSeriesWithWildcards(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"sumSeriesWithWildcards(metric1.foo.*.*,1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"sumSeriesWithWildcards",
			map[string][]*types.MetricData{
				"sumSeriesWithWildcards(metric1.baz)": {types.MakeMetricData("sumSeriesWithWildcards(metric1.baz)", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"sumSeriesWithWildcards(metric1.qux)": {types.MakeMetricData("sumSeriesWithWildcards(metric1.qux)", []float64{13, 15, 17, 19, 21}, 1, now32)},
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

func BenchmarkSumSeriesWithWildcards(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "sumSeriesWithWildcards(metric1.foo.bar*.*,1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.bar*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, 1),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, 1),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, 1),
				},
			},
		},
		{
			target: "sumSeriesWithWildcards(metric1.foo.*.*,1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, 1),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, 1),

					types.MakeMetricData("metric1.foo.bar3.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar3.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar4.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar4.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar5.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar5.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar6.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar6.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar7.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar7.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar8.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar8.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar9.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar9.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar10.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar10.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar11.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar11.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar12.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar12.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar13.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar13.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar14.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar14.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar15.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar15.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar16.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar16.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar17.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar17.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar18.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar18.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar19.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar19.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar20.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar20.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar21.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar21.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar22.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar22.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar23.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar23.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar24.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar24.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar25.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar25.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar26.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar26.qux", []float64{6, 7, 8, 9, 10}, 1, 1),

					types.MakeMetricData("metric1.foo.bar27.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
					types.MakeMetricData("metric1.foo.bar27.qux", []float64{6, 7, 8, 9, 10}, 1, 1),
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
