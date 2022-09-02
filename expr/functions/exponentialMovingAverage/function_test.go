package exponentialMovingAverage

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

func TestExponentialMovingAverage(t *testing.T) {
	startTime := int64(0)

	tests := []th.EvalTestItem{
		{
			"exponentialMovingAverage(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{2, 4, 6, 8, 12, 14, 16, 18, 20}, 1, startTime)},
			},
			[]*types.MetricData{
				types.MakeMetricData("exponentialMovingAverage(metric1,3)", []float64{4, 6, 9, 11.5, 13.75, 15.875, 17.9375}, 1, 0),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
