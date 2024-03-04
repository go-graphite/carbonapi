package delay

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"github.com/go-graphite/carbonapi/tests/compare"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestDelay(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"delay(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("delay(metric1,3)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 1, 2, 3, math.NaN()}, 1, now32).SetTag("delay", "3").SetNameTag("delay(metric1,3)")},
		},
		{
			"delay(metric1,-3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), 1, 2, 3, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("delay(metric1,-3)",
				[]float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32).SetTag("delay", "-3").SetNameTag("delay(metric1,-3)")},
		},
		{
			"delay(metric1,0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("delay(metric1,0)",
				[]float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32).SetTag("delay", "0").SetNameTag("delay(metric1,0)")},
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

func BenchmarkDelay(b *testing.B) {
	target := "delay(metric*,3)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric*", From: 0, Until: 1}: {
			types.MakeMetricData("metric1", compare.GenerateMetrics(2046, 1, 10, 1), 1, 1),
			types.MakeMetricData("metric2", compare.GenerateMetrics(2046, 1, 10, 1), 1, 1),
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

func BenchmarkDelayReverse(b *testing.B) {
	target := "delay(metric*,-3)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric*", From: 0, Until: 1}: {
			types.MakeMetricData("metric1", compare.GenerateMetrics(2046, 1, 10, 1), 1, 1),
			types.MakeMetricData("metric2", compare.GenerateMetrics(2046, 1, 10, 1), 1, 1),
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
