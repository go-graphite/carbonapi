package rangeOfSeries

import (
	"math"
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

func TestRangeOfSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"rangeOfSeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {},
			},
			[]*types.MetricData{},
		},
		{
			"rangeOfSeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), 3, 4, 12, -10}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), math.NaN(), 15, 0, 6, 10}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, math.NaN(), 4, 5, 6, 7}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("rangeOfSeries(metric*)",
				[]float64{1, math.NaN(), math.NaN(), 12, 5, 6, 20}, 1, now32).SetNameTag("rangeOfSeries(metric*)")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
