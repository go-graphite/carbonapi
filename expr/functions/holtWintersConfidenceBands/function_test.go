package holtWintersConfidenceBands

import (
	"testing"

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

func TestNoPanicOnBigStep(t *testing.T) {

	// NOTE: the expected values of this test are not meaningful, its purpose is to reproduce a panic
	test := th.EvalTestItem{
		"holtWintersConfidenceBands(metric1,3)",
		map[parser.MetricRequest][]*types.MetricData{
			{"metric1", -604800, 1}: {types.MakeMetricData("metric1", []float64{1}, 604800, -604800)},
		},
		[]*types.MetricData{
			types.MakeMetricData("holtWintersConfidenceLower(metric1)", []float64{}, 604800, 0).SetTag("holtWintersConfidenceLower", "1"),
			types.MakeMetricData("holtWintersConfidenceUpper(metric1)", []float64{}, 604800, 0).SetTag("holtWintersConfidenceUpper", "1"),
		},
	}

	th.TestEvalExpr(t, &test)

}
