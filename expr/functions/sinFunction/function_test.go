package sinFunction

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

func TestSinFunction(t *testing.T) {
	var startTime int64 = 1

	tests := []th.EvalTestItemWithRange{
		{
			Target: `sinFunction("The.time.series")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{From: startTime, Until: startTime + 240}: {},
			},
			Want: []*types.MetricData{types.MakeMetricData("The.time.series",
				[]float64{0.8414709848078965, -0.9661177700083929, 0.9988152247235795, -0.936451400117644}, 60, startTime)},
			From:  startTime,
			Until: startTime + 240,
		},
		{
			Target: `sinFunction("The.time.series.2", 5.0, 10)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{From: startTime, Until: startTime + 60}: {},
			},
			Want: []*types.MetricData{types.MakeMetricData("The.time.series.2",
				[]float64{4.207354924039483, -4.9999510327535175, 4.18327819268028, -2.0201882266153253, -0.7931133440235449, 3.3511458792168733}, 10, startTime)},
			From:  startTime,
			Until: startTime + 60,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
