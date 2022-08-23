package asPercent

import (
	"context"
	"math"
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

func TestAsPersent(t *testing.T) {
	now32 := int64(time.Now().Unix())
	NaN := math.NaN()

	tests := []th.EvalTestItem{
		{
			"asPercent(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, NaN, NaN, 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, NaN, 3, NaN, 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("asPercent(metric1,metric2)",
				[]float64{50, NaN, NaN, NaN, NaN, 200}, 1, now32)},
		},
		{
			"asPercent(metricA*,metricB*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metricA*", 0, 1}: {
					types.MakeMetricData("metricA1", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("metricA2", []float64{1, 10, 20}, 1, now32),
				},
				{"metricB*", 0, 1}: {
					types.MakeMetricData("metricB1", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("metricB2", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("asPercent(metricA1,metricB1)",
				[]float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(metricA2,metricB2)",
					[]float64{25, 62.5, 1000}, 1, now32)},
		},
		{
			"asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, now32),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,Server3.memory.total)", []float64{25, 62.5, 1000}, 1, now32),
			},
		},

		// Extend tests
		{
			"asPercent(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, NaN, NaN, 3, 4, 14}, 1, now32),
					types.MakeMetricData("metric2", []float64{4, NaN, 3, NaN, 0, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(metric1)", []float64{20, NaN, NaN, 100, 100, 70}, 1, now32),
				types.MakeMetricData("asPercent(metric2)", []float64{80, NaN, 100, NaN, 0, 30}, 1, now32),
			},
		},
		{
			"asPercent(Server{1,2}.memory.used,Server{1,2}.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 11, 20}, 1, now32),
				},
				{"Server{1,2}.memory.total", 0, 1}: {
					types.MakeMetricData("Server2.memory.total", []float64{4, 2, 2}, 1, now32),
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,Server2.memory.total)", []float64{25, 550, 1000}, 1, now32),
			},
		},
		{
			"asPercent(Server{1,2}.memory.used,Server{1,2,3}.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 15}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 11, 20}, 1, now32),
				},
				{"Server{1,2,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 40, 25}, 1, now32),
					types.MakeMetricData("Server2.memory.total", []float64{4, 20, 40}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 20, 40}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 50, 60}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,Server2.memory.total)", []float64{25, 55, 50}, 1, now32),
				types.MakeMetricData("asPercent(MISSING,Server3.memory.total)", []float64{NaN, NaN, NaN}, 1, now32),
			},
		},
		{
			"asPercent(Server{1,2,3}.memory.used,Server{1,2}.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2,3}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 15}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 11, 20}, 1, now32),
					types.MakeMetricData("Server3.memory.used", []float64{1, 11, 20}, 1, now32),
				},
				{"Server{1,2}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 40, 25}, 1, now32),
					types.MakeMetricData("Server2.memory.total", []float64{4, 20, 40}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 50, 60}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,Server2.memory.total)", []float64{25, 55, 50}, 1, now32),
				types.MakeMetricData("asPercent(Server3.memory.used,MISSING)", []float64{NaN, NaN, NaN}, 1, now32),
			},
		},
		// tagged series
		{
			"asPercent(seriesByTag('name=metric', 'tag=A*'),metricB*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('name=metric', 'tag=A*')", 0, 1}: {
					types.MakeMetricData("metric;tag=A1", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("metric;tag=A2", []float64{1, 10, 20}, 1, now32),
				},
				{"metricB*", 0, 1}: {
					types.MakeMetricData("metricB1", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("metricB2", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(metric;tag=A1,metricB1)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(metric;tag=A2,metricB2)", []float64{25, 62.5, 1000}, 1, now32),
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

func TestAsPersentAligment(t *testing.T) {
	now32 := int64(time.Now().Unix())
	NaN := math.NaN()
	testAlignments := []th.EvalTestItem{
		{
			"asPercent(Server{1,2}.aligned.memory.used,Server{1,3}.aligned.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.aligned.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, now32),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, now32-1),
				},
				{"Server{1,3}.aligned.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.total", []float64{1, 4, 4, 8}, 1, now32-1),
					types.MakeMetricData("Server3.aligned.memory.total", []float64{4, 16, 2, 10}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.aligned.memory.used,Server1.aligned.memory.total)", []float64{NaN, 25, 500, 125, NaN}, 1, now32-1),
				types.MakeMetricData("asPercent(Server2.aligned.memory.used,Server3.aligned.memory.total)", []float64{NaN, 25, 62.5, 1000, NaN}, 1, now32-1),
			},
		},
		{
			"asPercent(Server{1,2}.aligned.memory.used,Server3.aligned.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.aligned.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, now32),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 2, 10, 20}, 1, now32-1),
				},
				{"Server3.aligned.memory.total", 0, 1}: {
					types.MakeMetricData("Server3.aligned.memory.total", []float64{4, 16, 2, 10, 40}, 1, now32-1),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.aligned.memory.used,Server3.aligned.memory.total)", []float64{NaN, 6.25, 1000, 100, 50}, 1, now32-1),
				types.MakeMetricData("asPercent(Server2.aligned.memory.used,Server3.aligned.memory.total)", []float64{0, 12.5, 500, 200, NaN}, 1, now32-1),
			},
		},
		{
			"asPercent(Server{1,2}.aligned.memory.used,100)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.aligned.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, now32),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, now32-1),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.aligned.memory.used,100)", []float64{NaN, 1, 20, 10, 20}, 1, now32-1),
				types.MakeMetricData("asPercent(Server2.aligned.memory.used,100)", []float64{0, 1, 10, 20, NaN}, 1, now32-1),
			},
		},
	}

	for _, tt := range testAlignments {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestAsPersentGroup(t *testing.T) {
	now32 := int64(time.Now().Unix())
	NaN := math.NaN()

	tests := []th.EvalTestItem{
		{
			"asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, now32),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,MISSING)", []float64{NaN, NaN, NaN}, 1, now32),
				types.MakeMetricData("asPercent(MISSING,Server3.memory.total)", []float64{NaN, NaN, NaN}, 1, now32),
			},
		},

		// Broken tests with current
		{
			"asPercent(Server{1,2}.memory.{used,free},Server{1,3}.memory.total,0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.{used,free}", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server1.memory.free", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, now32),
					types.MakeMetricData("Server2.memory.free", []float64{1, 20, 10}, 1, now32),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.free,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.free,MISSING)", []float64{NaN, NaN, NaN}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,MISSING)", []float64{NaN, NaN, NaN}, 1, now32),
				types.MakeMetricData("asPercent(MISSING,Server3.memory.total)", []float64{NaN, NaN, NaN}, 1, now32),
			},
		},
		{
			"asPercent(Server{1,2}.memory.{used,free},None,0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.{used,free}", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{2, 1, NaN}, 1, now32),
					types.MakeMetricData("Server1.memory.free", []float64{3, NaN, 8}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{4, NaN, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.free,None)", []float64{60, NaN, 100}, 1, now32),
				types.MakeMetricData("asPercent(Server1.memory.used,None)", []float64{40, 100, NaN}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,None)", []float64{100, NaN, 100}, 1, now32),
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

func BenchmarkAsPercent(b *testing.B) {
	NaN := math.NaN()
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "asPercent(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, NaN, NaN, 3, 4, 14}, 1, 1),
					types.MakeMetricData("metric2", []float64{4, NaN, 3, NaN, 0, 6}, 1, 1),
				},
			},
		},
		{
			target: "asPercent(metric1,metric2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, NaN, NaN, 3, 4, 12}, 1, 1)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, NaN, 3, NaN, 0, 6}, 1, 1)},
			},
		},
		{
			target: "asPercent(Server{1,2}.memory.used,Server{1,2,3}.memory.total)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, 1),
				},
				{"Server{1,2,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.total", []float64{1, 4, 4, 8}, 1, 1),
					types.MakeMetricData("Server2.aligned.memory.total", []float64{4, 16, 2, 10}, 1, 2),
					types.MakeMetricData("Server3.aligned.memory.total", []float64{4, 16, 2, 10}, 1, 2),
				},
			},
		},
		{
			target: "asPercent(Server{1,2,3}.memory.used,Server{1,2}.memory.total)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2,3}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, 1),
					types.MakeMetricData("Server3.aligned.memory.used", []float64{0, 1, 10, 20}, 1, 1),
				},
				{"Server{1,2}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.total", []float64{1, 4, 4, 8}, 1, 1),
					types.MakeMetricData("Server2.aligned.memory.total", []float64{4, 16, 2, 10}, 1, 2),
				},
			},
		},
		{
			target: "asPercent(Server{1,2}.memory.{used,free},Server{1,2}.memory.total)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.{used,free}", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, 1),
					types.MakeMetricData("Server1.aligned.memory.free", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.free", []float64{0, 1, 10, 20}, 1, 1),
				},
				{"Server{1,2}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.total", []float64{1, 4, 4, 8}, 1, 1),
					types.MakeMetricData("Server2.aligned.memory.total", []float64{4, 16, 2, 10}, 1, 2),
				},
			},
		},
		{
			target: "asPercent(Server{1,2}.aligned.memory.used,Server{1,3}.aligned.memory.total)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.aligned.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20}, 1, 1),
				},
				{"Server{1,3}.aligned.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.total", []float64{1, 4, 4, 8}, 1, 1),
					types.MakeMetricData("Server3.aligned.memory.total", []float64{4, 16, 2, 10}, 1, 2),
				},
			},
		},
		{
			target: "asPercent(Server{1,2}.aligned.memory.used,Server3.aligned.memory.total)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.aligned.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 2, 10, 20}, 1, 1),
				},
				{"Server3.aligned.memory.total", 0, 1}: {
					types.MakeMetricData("Server3.aligned.memory.total", []float64{4, 16, 2, 10, 40}, 1, 1),
				},
			},
		},
		{
			target: `asPercent(Server{1,2}.aligned.memory.used,100)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "Server{1,2}.aligned.memory.used", From: 0, Until: 1}: {
					types.MakeMetricData("Server1.aligned.memory.used", []float64{1, 20, 10, 20, 1, 20, 10, 20, 1, 20, 10, 20}, 1, 2),
					types.MakeMetricData("Server2.aligned.memory.used", []float64{0, 1, 10, 20, 0, 1, 10, 20, 0, 1, 10, 20}, 1, 1),
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

func BenchmarkAsPercentGroup(b *testing.B) {
	NaN := math.NaN()
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, 1),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, 1),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, 1),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, 1),
				},
			},
		},
		{
			target: "asPercent(Server{1,2}.memory.{used,free},None,0)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.{used,free}", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{2, 1, NaN}, 1, 1),
					types.MakeMetricData("Server1.memory.free", []float64{3, NaN, 8}, 1, 1),
					types.MakeMetricData("Server2.memory.used", []float64{4, NaN, 2}, 1, 1),
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
