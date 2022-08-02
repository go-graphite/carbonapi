package identity

import (
	"testing"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
	th "github.com/grafana/carbonapi/tests"
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

func TestIdentityFunction(t *testing.T) {
	var startTime int64 = 0

	tests := []th.EvalTestItemWithRange{
		{
			Target: `identity("The.time.series")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{From: startTime, Until: startTime + 240}: {},
			},
			Want: []*types.MetricData{types.MakeMetricData("identity(The.time.series)",
				[]float64{0, 60, 120, 180}, 60, startTime)},
			From:  startTime,
			Until: startTime + 240,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
