package minMax

import (
	"math"
	"testing"
	"time"

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"minMax(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{10, 20, 30, math.NaN(), 40, 50}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("minMax(metric1)",
				[]float64{0.0, 0.25, 0.50, math.NaN(), 0.75, 1.0}, 1, now32)},
		},
		{
			"minMax(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{10, 10, 10, math.NaN(), 10, 10}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("minMax(metric1)",
				[]float64{0, 0, 0, math.NaN(), 0, 0}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
