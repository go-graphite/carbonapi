// +build cairo

package cairo

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	md := New("")
	metadata.SetEvaluator(th.EvaluatorFromFunc(md[0].F))
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestEvalExpressionGraph(t *testing.T) {

	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("threshold",
				42.42,
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("42.42",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			parser.NewExpr("threshold",
				42.42,
				parser.ArgValue("fourty-two"),
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("fourty-two",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			parser.NewExpr("threshold",
				42.42,
				parser.ArgValue("fourty-two"),
				parser.ArgValue("blue"),
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("fourty-two",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			parser.NewExpr("threshold",
				42.42,
				parser.NamedArgs{"label": parser.ArgValue("fourty-two")},
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("fourty-two",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			//TODO(nnuss): test blue is being set rather than just not causing expression to parse/fail
			parser.NewExpr("threshold",
				42.42,
				parser.NamedArgs{"color": parser.ArgValue("blue")},
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("42.42",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			//TODO(nnuss): test blue is being set rather than just not causing expression to parse/fail
			parser.NewExpr("threshold",
				42.42,
				parser.NamedArgs{
					"label": parser.ArgValue("fourty-two-blue"),
					"color": parser.ArgValue("blue"),
				},
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("fourty-two-blue",
				[]float64{42.42, 42.42}, 1, now32)},
		},
		{
			// BUG(nnuss): This test actually fails with color = "" because of
			// how getStringNamedOrPosArgDefault works but we don't notice
			// because we're not testing color is set.
			// You may manually verify with this request URI: /render/?format=png&target=threshold(42.42,"gold",label="fourty-two-aurum")
			parser.NewExpr("threshold",
				42.42,
				parser.ArgValue("gold"),
				parser.NamedArgs{"label": parser.ArgValue("fourty-two-aurum")},
			),
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("fourty-two-aurum",
				[]float64{42.42, 42.42}, 1, now32)},
		},
	}

	for _, tt := range tests {
		th.TestEvalExpr(t, &tt)
	}
}
