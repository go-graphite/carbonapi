package summarize

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

func TestEvalSummarize(t *testing.T) {
	tenThirtyTwo, _, tenThirty := th.InitTestSummarize()
	now32 := tenThirty

	tests := []th.SummarizeEvalTestItem{
		{
			Target: "summarize(metric1,'5s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 35}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1,
					2, 2, 2, 2, 2,
					3, 3, 3, 3, 3,
					4, 4, 4, 4, 4,
					5, 5, 5, 5, 5,
					math.NaN(), 2, 3, 4, 5,
					math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
				}, 1, now32)},
			},
			Want:  []float64{5, 10, 15, 20, 25, 14, math.NaN()},
			From:  now32,
			Until: now32 + 35,
			Name:  "summarize(metric1,'5s')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 35,
		},
		{
			Target: "summarize(metric1,'5s')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 50}: {types.MakeMetricData("metric1", []float64{
					1, 2, 3, 4, 5,
				}, 10, now32)},
			},
			Want:  []float64{1, 2, 3, 4, 5},
			From:  now32,
			Until: now32 + 50,
			Name:  "summarize(metric1,'5s')",
			Step:  10,
			Start: now32,
			Stop:  now32 + 50,
		},
		{
			Target: "summarize(metric1,'5s','avg')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 35}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
			},
			Want:  []float64{1, 2, 3, 4, 5, 2, math.NaN()},
			From:  now32,
			Until: now32 + 35,
			Name:  "summarize(metric1,'5s','avg')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 35,
		},
		{
			Target: "summarize(metric1,'5s','max')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{1, 2, 3, 4.5, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','max')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','min')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{0, 1, 1.5, 2, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','min')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','last')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{1, 2, 3, 4.5, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','last')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','p50')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{0.5, 1.5, 2, 3, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','p50')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','p25')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{0, 1, 2, 3, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','p25')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','p99.9')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{1, 2, 3, 4.498, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','p99.9')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'5s','p100.1')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'5s','p100.1')",
			Step:  5,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'1s','p50')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", now32, now32 + 25*1}: {types.MakeMetricData("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			Want:  []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5},
			From:  now32,
			Until: now32 + 25*1,
			Name:  "summarize(metric1,'1s','p50')",
			Step:  1,
			Start: now32,
			Stop:  now32 + 25*1,
		},
		{
			Target: "summarize(metric1,'10min')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", tenThirtyTwo, tenThirty + 30*60}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2,
					3, 3, 3, 3, 3, 4, 4, 4, 4, 4,
					5, 5, 5, 5, 5}, 60, tenThirtyTwo)},
			},
			Want:  []float64{11, 31, 33},
			From:  tenThirtyTwo,
			Until: tenThirty + 30*60,
			Name:  "summarize(metric1,'10min')",
			Step:  600,
			Start: tenThirty,
			Stop:  tenThirty + 30*60,
		},
		{
			Target: "summarize(metric1,'10min','sum',true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", tenThirtyTwo, tenThirtyTwo + 25*60}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2,
					3, 3, 3, 3, 3, 4, 4, 4, 4, 4,
					5, 5, 5, 5, 5}, 60, tenThirtyTwo)},
			},
			Want:  []float64{15, 35, 25},
			From:  tenThirtyTwo,
			Until: tenThirtyTwo + 25*60,
			Name:  "summarize(metric1,'10min','sum',true)",
			Step:  600,
			Start: tenThirtyTwo,
			Stop:  tenThirtyTwo + 25*60,
		},
		{
			Target: "summarize(metric1,'10min','sum',true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", tenThirtyTwo, tenThirtyTwo + 25*60}: {types.MakeMetricData("metric1", []float64{
					1, 1, 1, 1, 1, 2, 2, 2, 2, 2,
					3, 3, 3, 3, 3, 4, 4, 4, 4, 4,
					5, 5, 5, 5, 5}, 60, tenThirtyTwo)},
			},
			Want:  []float64{15, 35, 25},
			From:  tenThirtyTwo,
			Until: tenThirtyTwo + 25*60,
			Name:  "summarize(metric1,'10min','sum',true)",
			Step:  600,
			Start: tenThirtyTwo,
			Stop:  tenThirtyTwo + 25*60,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}
