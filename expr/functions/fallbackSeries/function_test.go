package fallbackSeries

import (
	"testing"
	"time"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
	th "github.com/grafana/carbonapi/tests"
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

func TestFallbackSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"fallbackSeries(metric*,fallbackmetric)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}:        {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"fallbackmetric", 0, 1}: {types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32),
			},
		},
		{
			"fallbackSeries(metric1,metrc2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32),
			},
		},
		{
			"fallbackSeries(absentmetric,fallbackmetric)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}:        {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"fallbackmetric", 0, 1}: {types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32),
			},
		},
		{
			"fallbackSeries(metric1,metrc2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32),
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
