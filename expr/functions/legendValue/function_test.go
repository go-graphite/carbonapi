package legendValue

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"legendValue(metric1,\"avg\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (avg: 3)",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"sum\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (sum: 15)",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"total\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (total: 15)",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"sum\",\"avg\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (sum: 15) (avg: 3)",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"sum\",\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 10000, 20000, -30000, -40000}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (sum: -40.00K )",
				[]float64{0, 10000, 20000, -30000, -40000}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"avg\",\"total\",\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 10000, 20000, -30000, -40000}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (avg: -8.00K ) (total: -40.00K )",
				[]float64{0, 10000, 20000, -30000, -40000}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"sum\",\"binary\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 10000, 20000, -30000, -40000}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (sum: -39.06Ki )",
				[]float64{0, 10000, 20000, -30000, -40000}, 1, now32).SetNameTag("metric1")},
		},
		{
			"legendValue(metric1,\"avg\",\"total\",\"binary\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, 10000, 20000, -30000, -40000}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1 (avg: -7.81Ki ) (total: -39.06Ki )",
				[]float64{0, 10000, 20000, -30000, -40000}, 1, now32).SetNameTag("metric1")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
