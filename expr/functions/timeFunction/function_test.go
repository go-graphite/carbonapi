package timeFunction

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

func TestTimeFunction(t *testing.T) {
	var startTime int64 = 1

	tests := []th.EvalTestItemWithRange{
		{
			Target: `timeFunction("The.time.series")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{From: startTime, Until: startTime + 240}: {},
			},
			Want: []*types.MetricData{types.MakeMetricData("The.time.series",
				[]float64{1, 61, 121, 181}, 60, startTime)},
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
