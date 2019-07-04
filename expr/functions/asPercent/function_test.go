package asPercent

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

func TestAliasByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"asPercent(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("asPercent(metric1,metric2)",
				[]float64{50, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 200}, 1, now32)},
		},
		{
			"asPercent(metricA*,metricB*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metricA*", 0, 1}: {
					types.MakeMetricData("metricA1", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("metricA2", []float64{1, 10, 20}, 1, now32),
				},
				{"metricB*", 0, 1}: {
					types.MakeMetricData("metricB1", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("metricB2", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("asPercent(metricA1,metricB1)",
				[]float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(metricA2,metricB2)",
					[]float64{25, 62.5, 1000}, 1, now32)},
		},
		{
			"asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, now32),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,Server3.memory.total)", []float64{25, 62.5, 1000}, 1, now32),
			},
		},
		{
			"asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)",
			map[parser.MetricRequest][]*types.MetricData{
				{"Server{1,2}.memory.used", 0, 1}: {
					types.MakeMetricData("Server1.memory.used", []float64{1, 20, 10}, 1, now32),
					types.MakeMetricData("Server2.memory.used", []float64{1, 10, 20}, 1, now32),
				},
				{"Server{1,3}.memory.total", 0, 1}: {
					types.MakeMetricData("Server1.memory.total", []float64{4, 4, 8}, 1, now32),
					types.MakeMetricData("Server3.memory.total", []float64{4, 16, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("asPercent(Server1.memory.used,Server1.memory.total)", []float64{25, 500, 125}, 1, now32),
				types.MakeMetricData("asPercent(Server2.memory.used,MISSING)", []float64{math.NaN(), math.NaN(), math.NaN()}, 1, now32),
				types.MakeMetricData("asPercent(MISSING,Server3.memory.total)", []float64{math.NaN(), math.NaN(), math.NaN()}, 1, now32),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
