package exponentialMovingAverage

import (
	"context"
	"testing"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
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

func TestExponentialMovingAverage(t *testing.T) {
	startTime := int64(0)

	tests := []th.EvalTestItem{
		{
			"exponentialMovingAverage(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, 1, startTime)},
			},
			[]*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(metric1,3)", []float64{4, 6, 9, 11.5, 13.75, 15.875, 17.9375}, 1, 0).SetTag("exponentialMovingAverage", "3"),
			},
		},
		{
			"exponentialMovingAverage(metric1,'3s')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, 1, startTime)},
			},
			[]*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(metric1,\"3s\")", []float64{4, 6, 9, 11.5, 13.75, 15.875, 17.9375}, 1, 0).SetTag("exponentialMovingAverage", "3"),
			},
		},
		{
			// if the window is larger than the length of the values, the result should just be the average.
			// this matches graphiteweb's behavior
			"exponentialMovingAverage(metric1,100)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, startTime)},
			},
			[]*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(metric1,100)", []float64{2}, 1, 0).SetTag("exponentialMovingAverage", "100"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func BenchmarkExponentialMovingAverage(b *testing.B) {
	target := "exponentialMovingAverage(metric1,3)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric[1234]", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, 1, 0)},
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

func BenchmarkExponentialMovingAverageStr(b *testing.B) {
	target := "exponentialMovingAverage(metric1,'3s')"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric[1234]", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, 1, 0)},
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
