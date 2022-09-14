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
			[]float64{35, 70, 105, 140, math.NaN(), 25},
			"hitcount(metric1,'30s')",
			30,
			now32,
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
			tenFiftyNine,
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
			[]float64{105, 270},
			"hitcount(metric1,'1h',true)",
			3600,
			tenFiftyNine - (59 * 60),
			tenFiftyNine + 25*5,
		},
		{
			"hitcount(metric1,\"1h\",alignToInterval=true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			[]float64{105, 270},
			"hitcount(metric1,'1h',true)",
			3600,
			tenFiftyNine - (59 * 60),
			tenFiftyNine + 25*5,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestSummarizeEvalExpr(t, &tt)
		})
	}

}
