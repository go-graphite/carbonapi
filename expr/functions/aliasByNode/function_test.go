package aliasByNode

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/functions/aggregate"
	"github.com/go-graphite/carbonapi/expr/functions/aliasSub"
	"github.com/go-graphite/carbonapi/expr/functions/perSecond"
	"github.com/go-graphite/carbonapi/expr/functions/transformNull"
)

func init() {
	md := New("")
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
	asFunc := aliasSub.New("")
	for _, m := range asFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}
	tnFunc := transformNull.New("")
	for _, m := range tnFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}
	psFunc := perSecond.New("")
	for _, m := range psFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}
	aggFunc := aggregate.New("")
	for _, m := range aggFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}

	evaluator := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
	metadata.SetEvaluator(evaluator)
}

func TestAliasByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		// issue 517
		{
			"aliasByNode(aliasSub(a.b.c.d.e, '(.*)', '0.1.2.@.4'), 2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"a.b.c.d.e", 0, 1}: {
					types.MakeMetricData("a.b.c.d.e", []float64{8, 2, 4}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("2", []float64{8, 2, 4}, 1, now32)},
		},
		{
			Target: "aliasByNode(aliasSub(a.b.c.d.e, '(.*)', '0.1.2.@.4'), 2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "a.b.c.d.e", From: 0, Until: 1}: {types.MakeMetricData("a.b.c.d.e", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("2", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(aliasSub(transformNull(metric1.foo.bar.ba*, 0), 'baz', 'word'), 2, 3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.ba*", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar.word", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("transformNull", "0")},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1,3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: "aliasByNode(metric1.foo.bar.baz,1,-2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo.bar",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, "foo")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("foo", "bar").SetTag("baz", "bam")},
		},
		{
			Target: `aliasByTags(*, "foo", "name")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("metric1;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("foo", "bar")},
		},
		{
			Target: `aliasByTags(*, 2, "blah", "foo", 1)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData(".bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("foo", "bar").SetTag("baz", "bam")},
		},
		{
			Target: `aliasByTags(*, 2, "baz", "foo", 1)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bam.bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("foo", "bar").SetTag("baz", "bam")},
		},
		{
			Target: `aliasByTags(perSecond(*), 'name')`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("base.metric1", []float64{math.NaN(), 1, 1, 1, 1}, 1, now32).SetTag("foo", "bar").SetTag("baz", "bam").SetTag("perSecond", "1")},
		},
		{
			Target: "aliasByNode(metric1.fo*.bar.baz,1,3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.fo*.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo==.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("foo==.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			Target: `aliasByTags(*, 2, "baz", "foo", 1)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "*", From: 0, Until: 1}: {types.MakeMetricData("base.metric1;foo=bar=;baz=bam==", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			Want: []*types.MetricData{types.MakeMetricData("bam==.bar=.metric1", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("foo", "bar=").SetTag("baz", "bam==")},
		},
		// extract nodes with sumSeries
		{
			Target: `aliasByNode(sumSeries(metric.{a,b}*.b), 1, 2)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric.{a,b}*.b", 0, 1}: {
					types.MakeMetricData("metric.a1.b", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric.b2.b", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric.c2.b", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			Want: []*types.MetricData{types.MakeMetricData("{a,b}*.b", []float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTag("aggregatedBy", "sum")},
		},
		// extract tags from seriesByTag
		{
			Target: `aliasByTags(sumSeries(seriesByTag('tag2=value*', 'name=metric')), 'tag2', 'name')`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value*', 'name=metric')", 0, 1}: {
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric;tag2=value21;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric;tag2=value21;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			// Want: []*types.MetricData{types.MakeMetricData("value____.metric", []float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32)},
			Want: []*types.MetricData{
				types.MakeMetricData("value21.metric", []float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTag("aggregatedBy", "sum").SetTag("tag2", "value21")},
		},
		// TODO msaf1980: tests with extractTagsFromArgs = true
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func BenchmarkAliasByNode(b *testing.B) {
	target := "aliasByNode(metric1.foo.bar.baz,1,3)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {
			types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
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
