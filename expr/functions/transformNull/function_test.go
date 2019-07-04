package transformNull

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"math"
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

func TestDerivative(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`transformNull(metric1)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1)",
				[]float64{1, 0, 0, 3, 4, 12}, 1, now32)},
		},
		{
			`transformNull(metric1, default=5)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1,5)",
				[]float64{1, 5, 5, 3, 4, 12}, 1, now32)},
		},
		{
			`transformNull(metric1, default=5, referenceSeries=metric2.*)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), math.NaN(), 4, 12}, 1, now32)},
				{"metric2.*", 0, 1}: {
					types.MakeMetricData("metric2.foo", []float64{math.NaN(), 3, math.NaN(), 3, math.NaN(), 12}, 1, now32),
					types.MakeMetricData("metric2.bar", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1,5)",
				[]float64{1, 5, math.NaN(), 5, 4, 12}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
