package toLowerCase

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

func TestToLowerCaseFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"lower(METRIC.TEST.FOO)",
			map[parser.MetricRequest][]*types.MetricData{
				{"METRIC.TEST.FOO", 0, 1}: {types.MakeMetricData("METRIC.TEST.FOO", []float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric.test.foo",
				[]float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"lower(METRIC.TEST.FOO,7)",
			map[parser.MetricRequest][]*types.MetricData{
				{"METRIC.TEST.FOO", 0, 1}: {types.MakeMetricData("METRIC.TEST.FOO", []float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("METRIC.tEST.FOO",
				[]float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"lower(METRIC.TEST.FOO,-3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"METRIC.TEST.FOO", 0, 1}: {types.MakeMetricData("METRIC.TEST.FOO", []float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("METRIC.TEST.fOO",
				[]float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"lower(METRIC.TEST.FOO,0,7,12)",
			map[parser.MetricRequest][]*types.MetricData{
				{"METRIC.TEST.FOO", 0, 1}: {types.MakeMetricData("METRIC.TEST.FOO", []float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("mETRIC.tEST.fOO",
				[]float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
		{
			"toLowerCase(METRIC.TEST.FOO)",
			map[parser.MetricRequest][]*types.MetricData{
				{"METRIC.TEST.FOO", 0, 1}: {types.MakeMetricData("METRIC.TEST.FOO", []float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric.test.foo",
				[]float64{1, 2, 0, 7, 8, 20, 30, math.NaN()}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
