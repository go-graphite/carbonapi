package aliasByMetric

import (
	"context"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestAliasByMetric(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aliasByMetric(metric1.foo.bar.baz)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("baz", []float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("metric1.foo.bar.baz")},
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

func BenchmarkAliasByMetric(b *testing.B) {
	target := "aliasByMetric(metric1.foo.bar.baz)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric1.foo.bar.baz", From: 0, Until: 1}: {
			types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, 1),
		},
	}

	eval := th.EvaluatorFromFunc(md[0].F)
	exp, _, err := parser.ParseExpr(target)
	if err != nil {
		b.Fatalf("failed to parse %s: %+v", target, err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g, err := eval.Eval(context.Background(), exp, 0, 1, metrics)
		if err != nil {
			b.Fatalf("failed to eval %s: %+v", target, err)
		}
		_ = g
	}
}
