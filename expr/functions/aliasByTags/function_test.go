package aliasByTags

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

func TestAliasByTags(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("aliasByTags",
				"metric1.foo.bar.baz;foo=bar", parser.ArgValue("foo"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.bar.baz;foo=bar", 0, 1}: {types.MakeMetricData("metric1.foo.bar.baz;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			parser.NewExpr("aliasByTags",
				"metric1;foo=bar", parser.ArgValue("foo"), parser.ArgValue("name"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1;foo=bar", 0, 1}: {types.MakeMetricData("metric1;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
