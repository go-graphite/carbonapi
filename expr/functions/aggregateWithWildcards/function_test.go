package aggregateWithWildcards

import (
	"context"
	"math"
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

func TestAggregateWithWildcards(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`aggregateWithWildcards(metric[123],"avg",0)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3.foo.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("foo.bar.baz",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32).SetNameTag(`metric[123]`).SetTag("aggregatedBy", "avg")},
		},
		{
			`aggregateWithWildcards(metric[123],"diff",1)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo2.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.bar.baz", []float64{-1, math.NaN(), -1, 3, -1, -1}, 1, now32).SetNameTag(`metric[123]`).SetTag("aggregatedBy", "diff"),
				types.MakeMetricData("metric2.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32).SetNameTag(`metric[123]`).SetTag("aggregatedBy", "diff"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"max",2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric1.foo.bar3.baz1", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar4.baz2", []float64{4, math.NaN(), 5, 6, 7, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.foo.baz1", []float64{3, math.NaN(), 4, 5, 6, 5}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "max"),
				types.MakeMetricData("metric1.foo.baz2", []float64{4, math.NaN(), 5, 6, 7, 6}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "max"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"min",3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric1.foo.bar.baz2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz4", []float64{4, math.NaN(), 5, 6, 7, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.foo.bar", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "min"),
				types.MakeMetricData("metric2.foo.bar", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "min"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"median",0,3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2.foo.bar1.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3.foo.bar2.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2.foo.bar2.baz", []float64{4, math.NaN(), 5, 6, 7, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("foo.bar1", []float64{1.5, math.NaN(), 2.5, 3, 4.5, 5.5}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "median"),
				types.MakeMetricData("foo.bar2", []float64{3.5, math.NaN(), 4.5, 5.5, 6.5, 8}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "median"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"multiply",1,2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo1.bar.baz", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric1.foo2.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric1.foo3.bar.qux", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo4.bar.qux", []float64{4, math.NaN(), 5, 6, 7, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.baz", []float64{2, math.NaN(), 6, math.NaN(), 20, 30}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "multiply"),
				types.MakeMetricData("metric1.qux", []float64{12, math.NaN(), 20, 30, 42, math.NaN()}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "multiply"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"range",0,2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz.1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3.foo.bar.baz.1", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric4.foo.bar.baz", []float64{4, math.NaN(), 5, 6, 7, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("foo.baz.1", []float64{2, math.NaN(), 2, 2, 2, 0}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "range"),
				types.MakeMetricData("foo.baz", []float64{2, math.NaN(), 2, 0, 2, 3}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "range"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"sum",1,3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo1.bar.baz.qux", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric1.foo2.bar.baz.quux", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric1.foo3.bar.baz.qux", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo4.bar.baz.quux", []float64{4, math.NaN(), 5, 6, 7, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.bar.qux", []float64{4, math.NaN(), 6, 8, 10, 6}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "sum"),
				types.MakeMetricData("metric1.bar.quux", []float64{6, math.NaN(), 8, 6, 12, 13}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "sum"),
			},
		},
		{
			`aggregateWithWildcards(metric[1234],"sum")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo1.bar.baz.qux", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric1.foo2.bar.baz.quux", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.foo1.bar.baz.qux", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "sum"),
				types.MakeMetricData("metric1.foo2.bar.baz.quux", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32).SetNameTag(`metric[1234]`).SetTag("aggregatedBy", "sum"),
			},
		},
		{
			`aggregateWithWildcards(metric[123456],"stddev",0,1,2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123456]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3.foo.bar.baz1", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric4.foo.bar.baz2", []float64{4, math.NaN(), 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metric5.foo.bar.baz1", []float64{5, math.NaN(), 6, 7, 8, 9}, 1, now32),
					types.MakeMetricData("metric6.foo.bar.baz2", []float64{6, math.NaN(), 7, 8, 9, 10}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("baz1", []float64{1.632993161855452, math.NaN(), 1.632993161855452, 1.632993161855452, 1.632993161855452, 1.5}, 1, now32).SetNameTag(`metric[123456]`).SetTag("aggregatedBy", "stddev"),
				types.MakeMetricData("baz2", []float64{1.632993161855452, math.NaN(), 1.632993161855452, 1, 1.632993161855452, 2.0548046676563256}, 1, now32).SetNameTag(`metric[123456]`).SetTag("aggregatedBy", "stddev"),
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
				"metric1.baz": {types.MakeMetricData("metric1.baz", []float64{12, 14, 16, 18, 20}, 1, now32).SetTag("aggregatedBy", "sum").SetNameTag("metric1.foo.*.*")},
				"metric1.qux": {types.MakeMetricData("metric1.qux", []float64{13, 15, 17, 19, 21}, 1, now32).SetTag("aggregatedBy", "sum").SetNameTag("metric1.foo.*.*")},
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

// This return is multireturn
func TestAverageSeriesWithWildcards(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"averageSeriesWithWildcards(metric1.foo.*.*,1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"averageSeriesWithWildcards",
			map[string][]*types.MetricData{
				"metric1.baz": {types.MakeMetricData("metric1.baz", []float64{6, 7, 8, 9, 10}, 1, now32).SetTag("aggregatedBy", "average").SetNameTag("metric1.foo.*.*")},
				"metric1.qux": {types.MakeMetricData("metric1.qux", []float64{6.5, 7.5, 8.5, 9.5, 10.5}, 1, now32).SetTag("aggregatedBy", "average").SetNameTag("metric1.foo.*.*")},
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

func TestFunctionMultiplySeriesWithWildcards(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"multiplySeriesWithWildcards(metric1.foo.*.*,1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 0, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
					types.MakeMetricData("metric1.foo.bar3.baz", []float64{2, 2, 2, 2, 2}, 1, now32),
				},
			},
			"multiplySeriesWithWildcards",
			map[string][]*types.MetricData{
				"metric1.baz": {types.MakeMetricData("metric1.baz", []float64{22, 48, 78, 112, 150}, 1, now32).SetTag("aggregatedBy", "multiply").SetNameTag("metric1.foo.*.*")},
				"metric1.qux": {types.MakeMetricData("metric1.qux", []float64{42, 0, 72, 90, 110}, 1, now32).SetTag("aggregatedBy", "multiply").SetNameTag("metric1.foo.*.*")},
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

func TestEmptyData(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"multiplySeriesWithWildcards(metric1.foo.*.*,1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {},
			},
			[]*types.MetricData{},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func BenchmarkMultiplySeriesWithWildcards(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "multiplySeriesWithWildcards(metric1.foo.bar*.*,1,2)",
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
			target: "multiplySeriesWithWildcards(metric1.foo.*.*,1,2)",
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

func BenchmarkMultiplyAverageSeriesWithWildcards(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "averageSeriesWithWildcards(metric1.foo.bar*.*,1,2)",
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
			target: "averageSeriesWithWildcards(metric1.foo.*.*,1,2)",
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
