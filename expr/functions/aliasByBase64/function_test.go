package aliasByBase64

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

func TestAlias(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aliasByBase64(bWV0cmljLm5hbWU=)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "bWV0cmljLm5hbWU=",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"bWV0cmljLm5hbWU=",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metric.name",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"alias(metric.bmFtZQ==, 2)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric.bmFtZQ==",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric.bmFtZQ==",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"metric.name",
					[]float64{1, 2, 3, 4, 5},
					1,
					now32,
				),
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

func BenchmarkAliasByMetric(b *testing.B) {
	benchmarks := []struct {
		target string
		M      map[parser.MetricRequest][]*types.MetricData
	}{
		{
			target: "aliasByBase64(bWV0cmljLm5hbWU=)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "bWV0cmljLm5hbWU=", From: 0, Until: 1}: {
					types.MakeMetricData("bWV0cmljLm5hbWU=", []float64{1, 2, 3, 4, 5}, 1, 1),
				},
			},
		},
		{
			target: "aliasByBase64(metric.bmFtZQ==, 2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric.bmFtZQ==", From: 0, Until: 1}: {
					types.MakeMetricData("metric.bmFtZQ==", []float64{1, 2, 3, 4, 5}, 1, 1),
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
