package linearRegression

import (
	"context"
	"math"
	"testing"

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

func TestLinearRegression(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"linearRegression(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{1, 2, math.NaN(), math.NaN(), 5, 6}, 1, 123),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("linearRegression(metric1)",
					[]float64{1, 2, 3, 4, 5, 6}, 1, 123).SetTag("linearRegressions", "123, 129"),
			},
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

func BenchmarkLinearRegression(b *testing.B) {
	target := "linearRegression(metric1)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), math.NaN(), 5, 6}, 1, 1)},
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
