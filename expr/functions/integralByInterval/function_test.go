package integralByInterval

import (
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

func TestFunction(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"integralByInterval(10s,'10s')",
			map[parser.MetricRequest][]*types.MetricData{
				{"10s", 0, 1}: {
					types.MakeMetricData("10s", []float64{1, 0, 2, 3, 4, 5, 0, 7, 8, 9, 10}, 2, 0),
				},
			},
			[]*types.MetricData{types.MakeMetricData(
				"integralByInterval(10s,'10s')",
				[]float64{1, 1, 3, 6, 10, 5, 5, 12, 20, 29, 10}, 2, 0).SetTag("integralByInterval", "10s"),
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
