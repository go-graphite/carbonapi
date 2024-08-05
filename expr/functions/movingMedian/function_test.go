package movingMedian

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"github.com/go-graphite/carbonapi/tests/compare"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestMovingMedian(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"movingMedian(metric1,4)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMedian(metric1,4)", []float64{math.NaN(), math.NaN(), math.NaN(), 1, 1, 1.5, 2, 2, 3, 4, 5, 6},
				1, 0).SetTag("movingMedian", "4").SetNameTag("metric1")}, // StartTime = from
		},
		{
			"movingMedian(metric1,5)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, 1, 2, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMedian(metric1,5)", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 1, 1, 2, 2, 2, 4, 4, 6, 6, 4, 2},
				1, 0).SetTag("movingMedian", "5").SetNameTag("metric1")}, // StartTime = from
		},
		{
			"movingMedian(metric1,\"1s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: -1, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, 1, 2, 0}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMedian(metric1,'1s')", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, 1, 2, 0},
				1, 0).SetTag("movingMedian", "'1s'").SetNameTag("metric1")}, // StartTime = from
		},
		{
			"movingMedian(metric1,\"3s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: -3, Until: 1}: {types.MakeMetricData("metric1", []float64{0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, 1, 2}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMedian(metric1,'3s')", []float64{0, 1, 1, 1, 1, 2, 2, 2, 4, 4, 6, 6, 6, 2},
				1, 0).SetTag("movingMedian", "'3s'").SetNameTag("metric1")}, // StartTime = from
		},
		{
			"movingMedian(metric1,\"5s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: -5, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, now32)}, // step > windowSize
			},
			[]*types.MetricData{types.MakeMetricData("movingMedian(metric1,'5s')", []float64{math.NaN(), math.NaN(), math.NaN()},
				10, now32).SetTag("movingMedian", "'5s'").SetNameTag("metric1")}, // StartTime = from
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExpr(t, eval, &tt)
		})
	}

}

func BenchmarkMovingMedian(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "movingMedian(metric1,'5s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: -5, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, 1)}, // step > windowSize
			},
		},
		{
			target: "movingMedian(metric1,4)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMedian(metric1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMedian(metric1,600)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
	}

	eval := th.EvaluatorFromFunc(md[0].F)

	for _, bm := range benchmarks {
		b.Run(bm.target, func(b *testing.B) {
			exp, _, err := parser.ParseExpr(bm.target)
			if err != nil {
				b.Fatalf("failed to parse %s: %+v", bm.target, err)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				g, err := eval.Eval(context.Background(), exp, 0, 1, bm.M)
				if err != nil {
					b.Fatalf("failed to eval %s: %+v", bm.target, err)
				}
				_ = g
			}
		})
	}
}
