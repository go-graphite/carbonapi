package hitcount

import (
	"math"
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
func TestHitcountEmptyData(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"hitcount(foo.bar, '1min')",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo.bar", 0, 1}: {},
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

func TestHitcount(t *testing.T) {
	_, tenFiftyNine, tenThirty := th.InitTestSummarize()
	now32 := tenThirty

	tests := []th.SummarizeEvalTestItem{
		{
			"hitcount(metric1,\"30s\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2,
					2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4,
					4, 4, 5, 5, 5, 5,
					math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
					5}, 5, now32)},
			},
			[]float64{5, 40, 75, 110, 120, 25},

			"hitcount(metric1,'30s')",
			30,
			1410344975,
			now32 + 31*5,
		},
		{
			"hitcount(metric1,\"1h\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			[]float64{375},
			"hitcount(metric1,'1h')",
			3600,
			1410343265,
			tenFiftyNine + 25*5,
		},
		{
			"hitcount(metric1,\"1h\",true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			[]float64{375},
			"hitcount(metric1,'1h',true)",
			3600,
			tenFiftyNine,
			tenFiftyNine + (((tenFiftyNine + 25*5) - tenFiftyNine) / 3600) + 3600, // The end time is adjusted because of alignToInterval being set to true
		},
		{
			"hitcount(metric1,\"1h\",alignToInterval=true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			[]float64{375},
			"hitcount(metric1,'1h',true)",
			3600,
			tenFiftyNine,
			tenFiftyNine + (((tenFiftyNine + 25*5) - tenFiftyNine) / 3600) + 3600, // The end time is adjusted because of alignToInterval being set to true
		},
		{
			"hitcount(metric1,\"15s\")", // Test having a smaller interval than the data's step
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					11, 7, 19, 32, 23}, 30, now32)},
			},
			[]float64{165, 165, 105, 105, 285, 285, 480, 480, 345, 345},

			"hitcount(metric1,'15s')",
			15,
			now32,
			now32 + 5*30,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestSummarizeEvalExpr(t, &tt)
		})
	}

}
