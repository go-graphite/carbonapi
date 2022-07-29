package aggregateWithWildcards

import (
	"math"
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
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
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
				types.MakeMetricData("metric1.bar.baz", []float64{-1, math.NaN(), -1, 3, -1, -1}, 1, now32),
				types.MakeMetricData("metric2.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
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
				types.MakeMetricData("metric1.foo.baz1", []float64{3, math.NaN(), 4, 5, 6, 5}, 1, now32),
				types.MakeMetricData("metric1.foo.baz2", []float64{4, math.NaN(), 5, 6, 7, 6}, 1, now32),
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
				types.MakeMetricData("metric1.foo.bar", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
				types.MakeMetricData("metric2.foo.bar", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
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
				types.MakeMetricData("foo.bar1", []float64{1.5, math.NaN(), 2.5, 3, 4.5, 5.5}, 1, now32),
				types.MakeMetricData("foo.bar2", []float64{3.5, math.NaN(), 4.5, 5.5, 6.5, 8}, 1, now32),
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
				types.MakeMetricData("metric1.baz", []float64{2, math.NaN(), 6, 3, 20, 30}, 1, now32),
				types.MakeMetricData("metric1.qux", []float64{12, math.NaN(), 20, 30, 42, math.NaN()}, 1, now32),
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
				types.MakeMetricData("foo.baz.1", []float64{2, math.NaN(), 2, 2, 2, 0}, 1, now32),
				types.MakeMetricData("foo.baz", []float64{2, math.NaN(), 2, 0, 2, 3}, 1, now32),
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
				types.MakeMetricData("metric1.bar.qux", []float64{4, math.NaN(), 6, 8, 10, 6}, 1, now32),
				types.MakeMetricData("metric1.bar.quux", []float64{6, math.NaN(), 8, 6, 12, 13}, 1, now32),
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
				types.MakeMetricData("baz1", []float64{1.632993161855452, math.NaN(), 1.632993161855452, 1.632993161855452, 1.632993161855452, 1.5}, 1, now32),
				types.MakeMetricData("baz2", []float64{1.632993161855452, math.NaN(), 1.632993161855452, 1, 1.632993161855452, 2.0548046676563256}, 1, now32),
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
