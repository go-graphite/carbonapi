package multiplySeries

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

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"multiplySeries(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric1,metric2)",
				[]float64{2, math.NaN(), math.NaN(), math.NaN(), 0, 72}, 1, now32)},
		},
		{
			"multiplySeries(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric1,metric2)",
				[]float64{2, math.NaN(), math.NaN(), math.NaN(), 0, 72}, 1, now32)},
		},
		{
			"multiplySeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, math.NaN(), 7, 8}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric1,metric2,metric3)",
				[]float64{6, math.NaN(), math.NaN(), math.NaN(), 0, 576}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

