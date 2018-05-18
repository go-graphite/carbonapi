package timeShift

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

func TestAbsolute(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("timeShift",
				"metric1", parser.ArgValue("0s"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("timeShift(metric1,'0')",
				[]float64{0, 1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			parser.NewExpr("timeShift",
				"metric1", parser.ArgValue("1s"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -1, 0}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, now32-1)},
			},
			[]*types.MetricData{types.MakeMetricData("timeShift(metric1,'-1')",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, now32-1)},
		},
		{
			parser.NewExpr("timeShift",
				"metric1", parser.ArgValue("1h"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", -60 * 60, -60*60 + 1}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, now32-60*60)},
			},
			[]*types.MetricData{types.MakeMetricData("timeShift(metric1,'-3600')",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, now32-60*60)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
