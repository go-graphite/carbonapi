package sum

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
			parser.NewExpr("sum",
				"metric1",
				"metric2",
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, -1, 2, -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0, 1, -2, 3, -4, -5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{0, 0, 0, 0, 0, 0}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
