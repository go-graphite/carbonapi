package aliasSub

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
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

func TestAliasByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("aliasSub",
				"metric1.foo.bar.baz", parser.ArgValue("foo"), parser.ArgValue("replaced"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.bar.baz", 0, 1}: {types.MakeMetricData("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1.replaced.bar.baz",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			parser.NewExpr("aliasSub",
				"metric1.TCP100", parser.ArgValue("^.*TCP(\\d+)"), parser.ArgValue("$1"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.TCP100", 0, 1}: {types.MakeMetricData("metric1.TCP100", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("100",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			parser.NewExpr("aliasSub",
				"metric1.TCP100", parser.ArgValue("^.*TCP(\\d+)"), parser.ArgValue("\\1"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.TCP100", 0, 1}: {types.MakeMetricData("metric1.TCP100", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("100",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
