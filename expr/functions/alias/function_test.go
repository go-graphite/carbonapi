package alias

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

func TestAlias(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"alias(metric1,\"renamed\")",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric1",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{types.MakeMetricData("renamed",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"alias(metric2, \"some format ${expr} str ${expr} and another ${expr\", true)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric2",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"some format metric2 str metric2 and another ${expr",
					[]float64{1, 2, 3, 4, 5},
					1,
					now32,
				),
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
