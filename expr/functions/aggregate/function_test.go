package aggregate

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
	"github.com/go-graphite/carbonapi/tests/compare"
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

func TestAverageSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`aggregate(metric[123], "avg")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avgSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
		},
		{
			`aggregate(metric[123], "avg_zero")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 4, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avg_zeroSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 3, 5, 4}, 1, now32)},
		},
		{
			`aggregate(metric[123], "count")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("countSeries(metric[123])",
				[]float64{3, math.NaN(), 3, 2, 3, 2}, 1, now32)},
		},
		{
			`aggregate(metric[123], "diff")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("diffSeries(metric[123])",
				[]float64{-4, math.NaN(), -5, -2, -7, -1}, 1, now32)},
		},
		{
			`aggregate(metric[123], "last")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("lastSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32)},
		},
		{
			`aggregate(metric[123], "max")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32)},
		},
		{
			`aggregate(metric[123], "min")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("minSeries(metric[123])",
				[]float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
		},
		{
			`aggregate(metric[123], "median")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("medianSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
		},
		{
			`aggregate(metric[123], "multiply")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric[123])",
				[]float64{6, math.NaN(), 24, 15, 120, 30}, 1, now32)},
		},
		{
			`aggregate(metric[123], "range")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("rangeSeries(metric[123])",
				[]float64{2, math.NaN(), 2, 2, 2, 1}, 1, now32)},
		},
		{
			`aggregate(metric[123], "sum")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric[123])",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32)},
		},
		{
			`stddevSeries(metric[123])`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.816496580927726, math.NaN(), 0.816496580927726, 1, 0.816496580927726, 0.5}, 1, now32)},
		},
		{
			`stddevSeries(metric1,metric2,metric3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric1,metric2,metric3)",
				[]float64{0.4714045207910317, 0.9428090415820634, 1.4142135623730951, 1.8856180831641267, 2.357022603955158}, 1, now32)},
		},
		{
			`aggregate(metric[123], "stddev")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.816496580927726, math.NaN(), 0.816496580927726, 1, 0.816496580927726, 0.5}, 1, now32)},
		},
		{
			`aggregate(metric[123], "stddev")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.4714045207910317, 0.9428090415820634, 1.4142135623730951, 1.8856180831641267, 2.357022603955158}, 1, now32)},
		},

		// sum
		{
			`sum(metric1,metric2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, -1, 2, -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0, 1, -2, 3, -4, -5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{0, 0, 0, 0, 0, 0}, 1, now32)},
		},
		{
			"sum(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, math.NaN()}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 3, math.NaN(), 5, 6, math.NaN()}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, 4, 5, 6, math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2,metric3)", []float64{6, 9, 8, 15, 11, math.NaN()}, 1, now32)},
		},
		{
			"sum(metric1,metric2,metric3,metric4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, math.NaN()}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 3, math.NaN(), 5, 6, math.NaN()}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, 4, 5, 6, math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2,metric3)", []float64{6, 9, 8, 15, 11, math.NaN()}, 1, now32)},
		},

		// minMax
		{
			"maxSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric1,metric2,metric3)",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32)},
		},
		{
			"minSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("minSeries(metric1,metric2,metric3)",
				[]float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
		},

		// avg
		{
			"averageSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("averageSeries(metric1,metric2,metric3)",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func TestAverageSeriesAlign(t *testing.T) {
	tests := []th.EvalTestItem{
		// Indx      |   0  |   1  |   2  |   3  |   4  |
		// commonStep  2
		// Start  0 (1 - 1 % 2)
		// metric1_2 |      |   1  |   3  |   5  |      |
		// metric1_2 |   1  |      |   4  |      |      |
		//
		// metric2_1 |      |   1  |      |   5  |      |
		// metric2_1 |   1  |      |   5  |      |      |

		// sum       |   2  |      |   9  |      |      |
		{
			// timeseries with different length
			Target: "sum(metric1_2,metric2_1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1_2", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 3, 5}, 1, 1)},
				{"metric2_1", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 5}, 2, 1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("sumSeries(metric1_2,metric2_1)",
				[]float64{2, 9}, 2, 0)},
		},
		{
			// First timeseries with broker StopTime
			Target: "sum(metric1,metric2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 3, 5, 8}, 1, 1).AppendStopTime(-1)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 5, 7}, 1, 1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{2, 8, 12, 8}, 1, 1)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprResult(t, &tt)
		})
	}

}

func BenchmarkAverageSeries(b *testing.B) {
	target := "sum(metric*)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric*", 0, 1}: {
			types.MakeMetricData("metric2", compare.GenerateMetrics(2046, 1, 10, 1), 2, 1),
			types.MakeMetricData("metric1", compare.GenerateMetrics(4096, 1, 10, 1), 1, 1),
			types.MakeMetricData("metric3", compare.GenerateMetrics(1360, 1, 10, 1), 3, 1),
		},
	}

	evaluator := metadata.GetEvaluator()
	exp, _, err := parser.ParseExpr(target)
	if err != nil {
		b.Fatalf("failed to parse %s: %+v", target, err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g, err := evaluator.Eval(context.Background(), exp, 0, 1, metrics)
		if err != nil {
			b.Fatalf("failed to eval %s: %+v", target, err)
		}
		_ = g
	}
}
