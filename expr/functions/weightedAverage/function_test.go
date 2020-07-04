package weightedAverage

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var None = math.NaN()

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
			"weightedAverage(metric*, metric*, 0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32),
					types.MakeMetricData("metric2", []float64{None, 2, None, 4, None, 6, None, 8, None, 10, None, 12, None, 14, None, 16, None, 18, None, 20}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, None, None, None}, 1, now32),
					types.MakeMetricData("metric4", []float64{1, 2, 3, 4, None, 6, None, None, 9, 10, 11, None, 13, None, None, None, None, 18, 19, 20}, 1, now32),
					types.MakeMetricData("metric5", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, None, None}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData(
				"weightedAverage(metric1,metric2,metric3,metric4,metric5, metric1,metric2,metric3,metric4,metric5, 0)",
				[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32,
			),
			},
		},
		{
			"weightedAverage(metric*.dividend, metric*.divisor, 0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*.dividend", 0, 1}: {
					types.MakeMetricData("metric1.dividend", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32),
					types.MakeMetricData("metric2.dividend", []float64{None, 2, None, 4, None, 6, None, 8, None, 10, None, 12, None, 14, None, 16, None, 18, None, 20}, 1, now32),
					types.MakeMetricData("metric3.dividend", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, None, None, None}, 1, now32),
					types.MakeMetricData("metric5.dividend", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, None, None}, 1, now32),
				},
				{"metric*.divisor", 0, 1}: {
					types.MakeMetricData("metric1.divisor", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32),
					types.MakeMetricData("metric3.divisor", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, None, None, None}, 1, now32),
					types.MakeMetricData("metric4.divisor", []float64{1, 2, 3, 4, None, 6, None, None, 9, 10, 11, None, 13, None, None, None, None, 18, 19, 20}, 1, now32),
					types.MakeMetricData("metric5.divisor", []float64{1, 2, None, None, None, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, None, None}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData(
				"weightedAverage(metric1.dividend,metric2.dividend,metric3.dividend,metric5.dividend, metric1.divisor,metric3.divisor,metric4.divisor,metric5.divisor, 0)",
				[]float64{0.75, 1.5, 1.5, 2.0, 5.0, 4.5, 7.0, 8.0, 6.75, 7.5, 8.25, 12.0, 9.75, 14.0, 15.0, 16.0, 17.0, 12.0, 9.5, 10.0}, 1, now32,
			),
			},
		},
		{
			"weightedAverage(metric*.dividend, metric*.divisor, 0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*.dividend", 0, 1}: {
					types.MakeMetricData("metric1.dividend", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32),
				},
				{"metric*.divisor", 0, 1}: {
					types.MakeMetricData("metric2.divisor", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now32),
				},
			},
			[]*types.MetricData{},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
