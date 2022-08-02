package removeBetweenPercentile

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`removeBetweenPercentile(metric[1234], 30)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{7, 7, 7, 7, 7, 7}, 1, now32),
					types.MakeMetricData("metric2", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{10, 10, 10, 10, 10, 10}, 1, now32),
					types.MakeMetricData("metric4", []float64{1, 1, 1, 1, 1, 1}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("removeBetweenPercentile(metric2, 30)", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				types.MakeMetricData("removeBetweenPercentile(metric3, 30)", []float64{10, 10, 10, 10, 10, 10}, 1, now32),
				types.MakeMetricData("removeBetweenPercentile(metric4, 30)", []float64{1, 1, 1, 1, 1, 1}, 1, now32),
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
