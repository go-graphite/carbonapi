package hitcount

import (
	"math"
	"testing"

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
			Target: "hitcount(metric1,\"30s\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 31*5}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2,
					2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4,
					4, 4, 5, 5, 5, 5,
					math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
					5}, 5, now32)},
			},
			Want:  []float64{5, 40, 75, 110, 120, 25},
			From:  now32,
			Until: now32 + 31*5,
			Name:  "hitcount(metric1,'30s')",
			Step:  30,
			Start: 1410344975,
			Stop:  now32 + 31*5,
		},
		{
			Target: "hitcount(metric1,\"1h\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", tenFiftyNine, tenFiftyNine + 25*5}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			Want:  []float64{375},
			From:  tenFiftyNine,
			Until: tenFiftyNine + 25*5,
			Name:  "hitcount(metric1,'1h')",
			Step:  3600,
			Start: 1410343265,
			Stop:  tenFiftyNine + 25*5,
		},
		{
			Target: "hitcount(metric1,\"1h\",true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 1410343200, 1410350340}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			Want:  []float64{375},
			From:  1410343200,
			Until: 1410350340,
			Name:  "hitcount(metric1,'1h',true)",
			Step:  3600,
			Start: tenFiftyNine,
			Stop:  tenFiftyNine + (((tenFiftyNine + 25*5) - tenFiftyNine) / 3600) + 3600, // The end time is adjusted because of alignToInterval being set to true
		},
		{
			Target: "hitcount(metric1,\"1h\",alignToInterval=true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 1410343200, 1410350340}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3,
					3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5,
					5}, 5, tenFiftyNine)},
			},
			Want:  []float64{375},
			From:  1410343200,
			Until: 1410350340,
			Name:  "hitcount(metric1,'1h',true)",
			Step:  3600,
			Start: tenFiftyNine,
			Stop:  tenFiftyNine + (((tenFiftyNine + 25*5) - tenFiftyNine) / 3600) + 3600, // The end time is adjusted because of alignToInterval being set to true
		},
		{
			Target: "hitcount(metric1,\"15s\")", // Test having a smaller interval than the data's step
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 5*30}: {types.MakeMetricData("metric1", []float64{
					11, 7, 19, 32, 23}, 30, now32)},
			},
			Want:  []float64{165, 165, 105, 105, 285, 285, 480, 480, 345, 345},
			From:  now32,
			Until: now32 + 5*30,
			Name:  "hitcount(metric1,'15s')",
			Step:  15,
			Start: now32,
			Stop:  now32 + 5*30,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestSummarizeEvalExpr(t, &tt)
		})
	}

}
