package aggregateLine

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

func TestConstantLine(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aggregateLine(metric[123])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateLine(metric1, 3)", []float64{3, 3}, 6, now32).SetTag("name", "metric1"),
				types.MakeMetricData("aggregateLine(metric2, 4)", []float64{4, 4}, 6, now32).SetTag("name", "metric2"),
				types.MakeMetricData("aggregateLine(metric3, 4.5)", []float64{4.5, 4.5}, 6, now32).SetTag("name", "metric3"),
			},
		},
		{
			"aggregateLine(metric[12],'avg',true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, 6.0, 3.0, 2.0, 5.0, 6.0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateLine(metric1, None)", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32).SetTag("name", "metric1"),
				types.MakeMetricData("aggregateLine(metric2, 4)", []float64{4, 4, 4, 4, 4, 4}, 1, now32).SetTag("name", "metric2"),
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

func BenchmarkAverageSeries(b *testing.B) {
	target := "aggregateLine(metric[12],'avg',true)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric[12]", 0, 1}: {
			types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 1),
			types.MakeMetricData("metric2", []float64{2.0, 6.0, 3.0, 2.0, 5.0, 6.0}, 1, 1),
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
