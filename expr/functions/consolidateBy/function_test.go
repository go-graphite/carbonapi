package consolidateBy

import (
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

func TestConsolidateBy(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"consolidateBy(metric1,\"sum\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"sum\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "sum").SetConsolidationFunc("sum")},
		},
		{
			"consolidateBy(metric1,\"avg\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"avg\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "avg").SetConsolidationFunc("avg")},
		},
		{
			"consolidateBy(metric1,\"min\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"min\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "min").SetConsolidationFunc("min")},
		},
		{
			"consolidateBy(metric1,\"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"max\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "max").SetConsolidationFunc("max")},
		},
		{
			"consolidateBy(metric1,\"first\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"first\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "first").SetConsolidationFunc("first")},
		},
		{
			"consolidateBy(metric1,\"last\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"last\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "last").SetConsolidationFunc("last")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
