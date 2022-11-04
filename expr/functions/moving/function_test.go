package moving

import (
	"context"
	"math"
	"strconv"
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

func TestMoving(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"movingWindow(metric1,'3sec','average')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 2, 2}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','avg_zero')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{1, 1, 0.3333333333333333}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','count')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 2, 1}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','diff')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{-4, -1, 3}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','range')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 3, 3}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','stddev')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 0, 3, math.NaN(), 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{0.8164965809277259, 1.247219128924647, 1.5}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingWindow(metric1,'3sec','last')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{3, 0, math.NaN()}, 1, 0).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingAverage(metric1,'3sec')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingAverage(metric1,'3sec')`,
				[]float64{2, 2, 2}, 1, 0).SetTag("movingAverage", "'3sec'").SetNameTag(`movingAverage(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingAverage(metric1,4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingAverage(metric1,4)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 1, 1.25, 1.5, 1.75, 2.5, 3.5, 4, 5}, 1, 0).SetTag("movingAverage", "4").SetNameTag("movingAverage(metric1,4)")}, // StartTime = from
		},
		{
			"movingAverage(metric1,'5s')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -5, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, now32)}, // step > windowSize
			},
			[]*types.MetricData{types.MakeMetricData(`movingAverage(metric1,'5s')`,
				[]float64{math.NaN(), math.NaN(), math.NaN()}, 10, now32).SetTag("movingAverage", "'5s'").SetNameTag(`movingAverage(metric1,'5s')`)}, // StartTime = from
		},
		{
			"movingSum(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingSum(metric1,2)",
				[]float64{math.NaN(), math.NaN(), 3, 5, 7, 9}, 1, 0).SetTag("movingSum", "2").SetNameTag("movingSum(metric1,2)")}, // StartTime = from
		},
		{
			"movingMin(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 2, 1, 0}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMin(metric1,2)",
				[]float64{math.NaN(), math.NaN(), 1, 2, 2, 1}, 1, 0).SetTag("movingMin", "2").SetNameTag("movingMin(metric1,2)")}, // StartTime = from
		},
		{
			"movingMax(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 2, 1, 0}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMax(metric1,2)",
				[]float64{math.NaN(), math.NaN(), 2, 3, 3, 2}, 1, 0).SetTag("movingMax", "2").SetNameTag("movingMax(metric1,2)")}, // StartTime = from
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestMovingXFilesFactor(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"movingSum(metric1,'3sec',0.5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, math.NaN(), 2, math.NaN(), 3}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(`movingSum(metric1,'3sec')`,
				[]float64{6, 6, 4, 3, math.NaN()}, 1, 0).SetTag("movingSum", "'3sec'").SetNameTag(`movingSum(metric1,'3sec')`)}, // StartTime = from
		},
		{
			"movingAverage(metric1,4,0.6)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, math.NaN(), 2, 4, math.NaN(), 4, 6, 8}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingAverage(metric1,4)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 1, 1.25,
					1.3333333333333333, 1.6666666666666667, 2.666666666666666, math.NaN(), 3.3333333333333335, 4.666666666666667}, 1, 0).SetTag(
				"movingAverage", "4").SetNameTag("movingAverage(metric1,4)")}, // StartTime = from
		},
		{
			"movingMax(metric1,2,0.5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, math.NaN(), math.NaN(), 0}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("movingMax(metric1,2)",
				[]float64{math.NaN(), math.NaN(), 2, 3, 3, math.NaN()}, 1, 0).SetTag("movingMax", "2").SetNameTag("movingMax(metric1,2)")}, // StartTime = from
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestMovingError(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			Target: "movingWindow(metric1,'','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'-','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'+','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'-s1','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -3, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
			},
			Error: parser.ErrBadType,
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExprWithError(t, &tt)
		})
	}

}

func BenchmarkMoving(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "movingAverage(metric1,'5s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -5, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, 1)}, // step > windowSize
			},
		},
		{
			target: "movingAverage(metric1,4)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingAverage(metric1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingSum(metric1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMin(metric1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMax(metric1,2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingAverage(metric1,600)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingSum(metric1,600)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMin(metric1,600)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
			},
		},
		{
			target: "movingMax(metric1,600)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", compare.GenerateMetrics(1024, 1.0, 9.0, 1.0), 1, 1)},
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
