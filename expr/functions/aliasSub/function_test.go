package aliasSub

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

func TestAliasSub(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aliasSub(metric1.foo.bar.baz, \"foo\", \"replaced\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.bar.baz", 0, 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1.replaced.bar.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"aliasSub(metric1.TCP100,\"^.*TCP(\\d+)\",\"$1\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.TCP100", 0, 1}: {types.MakeMetricData("metric1.TCP100", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("100",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"aliasSub(metric1.TCP100,\"^.*TCP(\\d+)\", \"\\1\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.TCP100", 0, 1}: {types.MakeMetricData("metric1.TCP100", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("100",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"aliasSub(metric1.foo.bar.baz, \"foo\", \"replaced\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.bar.baz", 0, 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1.replaced.bar.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		// #290
		{
			//"aliasSub(*, '.dns.([^.]+).zone.', '\\1 diff to sql')",
			"aliasSub(*, 'dns.([^.]*).zone.', '\\1 diff to sql ')",
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("diffSeries(dns.snake.sql_updated, dns.snake.zone_updated)", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("diffSeries(dns.snake.sql_updated, snake diff to sql updated)",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("diffSeries(dns.snake.sql_updated, snake diff to sql updated)")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func BenchmarkAverageAlias(b *testing.B) {
	target := `aliasSub(metric1.TCP100,"^.*TCP(\\d+)","$1")`
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric1.TCP100", 0, 1}:  {types.MakeMetricData("metric1.TCP100", []float64{1, 2, 3, 4, 5}, 1, 1)},
		{"metric1.TCP1024", 0, 1}: {types.MakeMetricData("metric1.TCP1024", []float64{1, 2, 3, 4, 5}, 1, 1)},
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
