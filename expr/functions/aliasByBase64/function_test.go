package aliasByBase64

import (
	"testing"
	"time"

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

func TestAlias(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"aliasByBase64(bWV0cmljLm5hbWU=)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "bWV0cmljLm5hbWU=",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"bWV0cmljLm5hbWU=",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metric.name",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"alias(metric.bmFtZQ==, 2)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric.bmFtZQ==",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"metric.bmFtZQ==",
						[]float64{1, 2, 3, 4, 5},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"metric.name",
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
