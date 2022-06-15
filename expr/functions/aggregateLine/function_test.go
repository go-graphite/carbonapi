package aggregateLine

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

func TestConstantLine(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aggregateLine(metric[123])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateLine(metric1, 3)", []float64{3, 3}, 6, now32),
				types.MakeMetricData("aggregateLine(metric2, 4)", []float64{4, 4}, 6, now32),
				types.MakeMetricData("aggregateLine(metric3, 4.5)", []float64{4.5, 4.5}, 6, now32),
			},
		},
		{
			"aggregateLine(metric[12],'avg',true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, 6.0, 3.0, 2.0, 5.0, 6.0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateLine(metric1, None)", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
				types.MakeMetricData("aggregateLine(metric2, 4)", []float64{4, 4, 4, 4, 4, 4}, 1, now32),
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
