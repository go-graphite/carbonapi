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

// Note: some of these tests are influenced by the testcases for moving* functions
// in Graphite-web. See: https://github.com/graphite-project/graphite-web/blob/master/webapp/tests/test_functions.py
func TestMoving(t *testing.T) {
	tests := []th.EvalTestItemWithRange{
		{
			Target: "movingAverage(metric1,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 20, 25}: {types.MakeMetricData("metric1", generateValues(10, 25, 1), 1, 20)},
				{"metric1", 10, 25}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 10)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,10)`,
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 20).SetTag("movingAverage", "10").SetNameTag(`movingAverage(metric1,10)`)},
			From:  20,
			Until: 25,
		},
		{
			Target: "movingAverage(metric1,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 20, 30}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 20)},
				{"metric1", 10, 30}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 1, 10)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,10)`,
				[]float64{0, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5}, 1, 20).SetTag("movingAverage", "10").SetNameTag(`movingAverage(metric1,10)`)},
			From:  20,
			Until: 30,
		},
		{
			Target: "movingAverage(metric1,60)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 610)},
				{"metric1", 550, 710}: {types.MakeMetricData("metric1", generateValues(0, 100, 1), 1, 600)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,60)`,
				[]float64{30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5, 45.5, 46.5, 47.5, 48.5, 49.5, 50.5, 51.5, 52.5, 53.5, 54.5, 55.5, 56.5, 57.5, 58.5, 59.5, 60.5, 61.5, 62.5, 63.5, 64.5, 65.5, 66.5, 67.5, 68.5, 69.5}, 1, 660).SetTag("movingAverage", "60").SetNameTag(`movingAverage(metric1,60)`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingAverage(metric1,'-1min')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 610)},
				{"metric1", 550, 710}: {types.MakeMetricData("metric1", generateValues(0, 100, 1), 1, 600)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,'-1min')`,
				[]float64{30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5, 45.5, 46.5, 47.5, 48.5, 49.5, 50.5, 51.5, 52.5, 53.5, 54.5, 55.5, 56.5, 57.5, 58.5, 59.5, 60.5, 61.5, 62.5, 63.5, 64.5, 65.5, 66.5, 67.5, 68.5, 69.5}, 1, 660).SetTag("movingAverage", "'-1min'").SetNameTag(`movingAverage(metric1,'-1min')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingMedian(metric1,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 20, 30}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 20)},
				{"metric1", 10, 30}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 1, 10)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMedian(metric1,10)`,
				[]float64{0, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5}, 1, 20).SetTag("movingMedian", "10").SetNameTag(`movingMedian(metric1,10)`)},
			From:  20,
			Until: 30,
		},
		{
			Target: "movingMedian(metric1,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 20, 25}: {types.MakeMetricData("metric1", generateValues(10, 25, 1), 1, 20)},
				{"metric1", 10, 25}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 10)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMedian(metric1,10)`,
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 20).SetTag("movingMedian", "10").SetNameTag(`movingMedian(metric1,10)`)},
			From:  20,
			Until: 25,
		},
		{
			Target: "movingMedian(metric1,60)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 610)},
				{"metric1", 550, 710}: {types.MakeMetricData("metric1", generateValues(0, 100, 1), 1, 600)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMedian(metric1,60)`,
				[]float64{30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5, 45.5, 46.5, 47.5, 48.5, 49.5, 50.5, 51.5, 52.5, 53.5, 54.5, 55.5, 56.5, 57.5, 58.5, 59.5, 60.5, 61.5, 62.5, 63.5, 64.5, 65.5, 66.5, 67.5, 68.5, 69.5}, 1, 660).SetTag("movingMedian", "60").SetNameTag(`movingMedian(metric1,60)`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingMedian(metric1,'1min')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 610)},
				{"metric1", 550, 710}: {types.MakeMetricData("metric1", generateValues(0, 100, 1), 1, 600)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMedian(metric1,'1min')`,
				[]float64{30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5, 45.5, 46.5, 47.5, 48.5, 49.5, 50.5, 51.5, 52.5, 53.5, 54.5, 55.5, 56.5, 57.5, 58.5, 59.5, 60.5, 61.5, 62.5, 63.5, 64.5, 65.5, 66.5, 67.5, 68.5, 69.5}, 1, 660).SetTag("movingMedian", "'1min'").SetNameTag(`movingMedian(metric1,'1min')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingMedian(metric1,'-1min')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", generateValues(10, 110, 1), 1, 610)},
				{"metric1", 550, 710}: {types.MakeMetricData("metric1", generateValues(0, 100, 1), 1, 600)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMedian(metric1,'-1min')`,
				[]float64{30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5, 45.5, 46.5, 47.5, 48.5, 49.5, 50.5, 51.5, 52.5, 53.5, 54.5, 55.5, 56.5, 57.5, 58.5, 59.5, 60.5, 61.5, 62.5, 63.5, 64.5, 65.5, 66.5, 67.5, 68.5, 69.5}, 1, 660).SetTag("movingMedian", "'-1min'").SetNameTag(`movingMedian(metric1,'-1min')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 2, 2}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','avg_zero')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{1.5, 1, 2}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','count')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 1, math.NaN(), 3}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 1, 2}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','diff')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{-1, 3, -5}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','range')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{3, 3, 5}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)},
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','stddev')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{1.247219128924647, 1.5, 2.5}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec','last')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 0, math.NaN(), 5}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{0, math.NaN(), 5}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
			From:  610,
			Until: 710,
		},
		{
			Target: "movingWindow(metric1,'3sec')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 610)},
				{"metric1", 607, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingWindow(metric1,'3sec')`,
				[]float64{2, 2, 2}, 1, 610).SetTag("movingWindow", "'3sec'").SetNameTag(`movingWindow(metric1,'3sec')`)}, // StartTime = from
			From:  610,
			Until: 710,
		},
		{
			Target: "movingAverage(metric1,4)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8}, 1, 610)},
				{"metric1", 606, 710}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8}, 1, 606)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,4)`,
				[]float64{1.25, 1.5, 1.75, 2.5, 3.5, 4.0, 5.0, 6.0}, 1, 610).SetTag("movingAverage", "4").SetNameTag(`movingAverage(metric1,4)`)}, // StartTime = from
			From:  610,
			Until: 710,
		},
		{
			Target: "movingAverage(metric1,'5s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, 610)}, // step > windowSize
				{"metric1", 605, 710}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 10, 605)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,'5s')`,
				[]float64{math.NaN(), math.NaN(), math.NaN()}, 10, 610).SetTag("movingAverage", "'5s'").SetNameTag(`movingAverage(metric1,'5s')`)}, // StartTime = from
			From:  610,
			Until: 710,
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}

func TestMovingXFilesFactor(t *testing.T) {
	tests := []th.EvalTestItemWithRange{
		{
			Target: "movingSum(metric1,'3sec',0.5)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 618}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, math.NaN(), 2, math.NaN(), 3}, 1, 610)},
				{"metric1", 607, 618}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, math.NaN(), 2, math.NaN(), 3}, 1, 607)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingSum(metric1,'3sec')`,
				[]float64{6, 4, 3, math.NaN(), 5}, 1, 610).SetTag("movingSum", "'3sec'").SetNameTag(`movingSum(metric1,'3sec')`)},
			From:  610,
			Until: 618,
		},
		{
			Target: "movingAverage(metric1,4,0.6)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 622}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, math.NaN(), 2, 4, math.NaN(), 4, 6, 8}, 1, 610)},
				{"metric1", 606, 622}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, math.NaN(), 2, 4, math.NaN(), 4, 6, 8}, 1, 606)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,4)`,
				[]float64{1.25, 1.3333333333333333, 1.6666666666666667, 2.6666666666666665, math.NaN(), 3.3333333333333335, 4.666666666666667, 6}, 1, 610).SetTag("movingAverage", "4").SetNameTag(`movingAverage(metric1,4)`)},
			From:  610,
			Until: 622,
		},
		{
			Target: "movingMax(metric1,2,0.5)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 610, 616}: {types.MakeMetricData("metric1", []float64{1, 2, 3, math.NaN(), math.NaN(), 0}, 1, 610)},
				{"metric1", 608, 616}: {types.MakeMetricData("metric1", []float64{1, 2, 3, math.NaN(), math.NaN(), 0}, 1, 608)},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingMax(metric1,2)`,
				[]float64{3, 3, math.NaN(), 0}, 1, 610).SetTag("movingMax", "2").SetNameTag(`movingMax(metric1,2)`)},
			From:  610,
			Until: 616,
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}

func TestMovingError(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			Target: "movingWindow(metric1,'','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 0)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'-','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 0)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'+','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, 0)},
			},
			Error: parser.ErrBadType,
		},
		{
			Target: "movingWindow(metric1,'-s1','average')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 1, 2, 3}, 1, now32)},
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

func generateValues(start, stop, step int64) (values []float64) {
	for i := start; i < stop; i += step {
		values = append(values, float64(i))
	}
	return
}
