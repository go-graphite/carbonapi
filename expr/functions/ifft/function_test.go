package ifft

import (
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

func TestFunction(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"ifft(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, 6, 7, 8}, 1, 0)},
			},
			[]*types.MetricData{types.MakeMetricData("ifft(metric1)",
				[]float64{4.5, 1.3065629648763766, 0.7071067811865476, 0.5411961001461969, 0.5, 0.5411961001461969, 0.7071067811865476, 1.3065629648763766}, 1, 0)},
		},
		{
			"ifft(metric1,None)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, 6, 7, 8}, 1, 0)},
			},
			[]*types.MetricData{types.MakeMetricData("ifft(metric1)",
				[]float64{4.5, 1.3065629648763766, 0.7071067811865476, 0.5411961001461969, 0.5, 0.5411961001461969, 0.7071067811865476, 1.3065629648763766}, 1, 0)},
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
