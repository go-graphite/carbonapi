package exponentialMovingAverage

import (
	"context"
	"math"
	"testing"

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

func TestExponentialMovingAverage(t *testing.T) {
	const from = 100
	const step = 10

	tests := []th.EvalTestItemWithRange{
		{
			Target: "exponentialMovingAverage(metric1,'30s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", from - 30, from + step*6}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, step, from-30)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(metric1,\"30s\")", []float64{4, 4.258065, 4.757544, 5.353832, 6.040681, 6.81225, 7.663073}, step, from).SetTag("exponentialMovingAverage", `"30s"`),
			},
			From:  from,
			Until: from + step*6,
		},
		{
			Target: "exponentialMovingAverage(empty,3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				// When the window is an integer, the original from-until range is used to get the step.
				// That's why two requests are made.
				{"empty", from, from + step*4}:          {},
				{"empty", from - step*3, from + step*4}: {},
			},
			Want:  []*types.MetricData{},
			From:  from,
			Until: from + step*4,
		},
		{
			Target: "exponentialMovingAverage(metric_changes_rollup,4)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric_changes_rollup", from, from + step*6}: {types.MakeMetricData("metric_changes_rollup", []float64{8, 12, 14, 16, 18, 20}, step, from)},
				// when querying for the preview window, the store changes the rollup and the step changes
				{"metric_changes_rollup", from - step*4, from + step*6}: {types.MakeMetricData("metric_changes_rollup", []float64{10, 20}, step*10, from-step*4)},
			},
			Want: []*types.MetricData{
				// since the input is shorter than the window, the result should be just the average
				types.MakeMetricData("exponentialMovingAverage(metric_changes_rollup,4)", []float64{15}, step*10, from).SetTag("exponentialMovingAverage", "4"),
			},
			From:  from,
			Until: from + step*6,
		},
		{
			// copied from Graphite Web
			Target: "exponentialMovingAverage(halfNone,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"halfNone", from, from + 10}:      {types.MakeMetricData("halfNone", append(append(append(nans(10), rangeFloats(0, 5, 1)...), math.NaN()), rangeFloats(5, 9, 1)...), 1, from)},
				{"halfNone", from - 10, from + 10}: {types.MakeMetricData("halfNone", append(append(append(nans(10), rangeFloats(0, 5, 1)...), math.NaN()), rangeFloats(5, 9, 1)...), 1, from-10)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(halfNone,10)", []float64{0, 0.0, 0.181818, 0.512397, 0.964688, 1.516563, math.NaN(), 2.149915, 2.849931, 3.604489, 4.403673}, 1, from).SetTag("exponentialMovingAverage", `10`),
			},
			From:  from,
			Until: from + 10,
		},
		// copied from Graphite Web
		{
			Target: `exponentialMovingAverage(collectd.test-db0.load.value,"-30s")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db0.load.value", from - 30, from + 30}: {types.MakeMetricData("collectd.test-db0.load.value", rangeFloats(0, 60, 1), 1, from-30)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(collectd.test-db0.load.value,\"-30s\")", []float64{
					14.5, 15.5, 16.5, 17.5, 18.5, 19.5, 20.5, 21.5, 22.5, 23.5, 24.5, 25.5, 26.5, 27.5, 28.5, 29.5, 30.5, 31.5, 32.5, 33.5, 34.5, 35.5, 36.5, 37.5, 38.5, 39.5, 40.5, 41.5, 42.5, 43.5, 44.5,
				}, 1, from).SetTag("exponentialMovingAverage", `"-30s"`),
			},
			From:  from,
			Until: from + 30,
		},
	}

	for _, tt := range tests {
		tt := tt
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}

func nans(n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = math.NaN()
	}
	return res
}

func rangeFloats(start, end, step float64) []float64 {
	res := make([]float64, 0, int((end-start)/step))
	for i := start; i < end; i += step {
		res = append(res, i)
	}
	return res
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
