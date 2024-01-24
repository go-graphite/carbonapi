package identity

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
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExprWithRange(t, eval, &tt)
		})
	}
}
