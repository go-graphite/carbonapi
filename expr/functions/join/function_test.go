package join

import (
	"math"
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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"join(metric1, metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{},
		},
		{
			"join(metric1, metric2, \"OR\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32),
				types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32),
			},
		},
		{
			"join(metric1, metric2, \"XOR\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32),
				types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32),
			},
		},
		{
			"join(metric1, metric2, \"SUB\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, -3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{math.NaN(), -1, math.NaN(), -3, 4, 5}, 1, now32),
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