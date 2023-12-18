package alias

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
			"alias(metric1,\"renamed\")",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric1",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{types.MakeMetricData("renamed",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"alias(metric2, \"some format ${expr} str ${expr} and another ${expr\", true)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric2",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"some format metric2 str metric2 and another ${expr",
					[]float64{1, 2, 3, 4, 5},
					1,
					now32,
				),
			},
		},
		{
			"alias(metric2, 'Метрика 2')",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric2",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"Метрика 2",
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

func BenchmarkAverageAlias(b *testing.B) {
	target := `alias(metric1, "renamed")`
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, 1)},
		{Metric: "metric2", From: 0, Until: 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 1, 1)},
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
