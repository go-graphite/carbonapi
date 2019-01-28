package aboveSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"testing"
	"time"
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

func TestDiffSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("aboveSeries",
				"metric1", 7, parser.ArgValue("Kotik"), parser.ArgValue("Bog"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricSobaka", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricKotik", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricHomyak", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricBog",
				[]float64{3, 4, 5, 6, 7, 8}, 1, now32)},
		},
		{
			parser.NewExpr("aboveSeries",
				"metric1", 7, parser.ArgValue(".*Ko.ik$"), parser.ArgValue("Bog"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricSobaka", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricKotik", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricHomyak", []float64{4, 4, 5, 5, 6, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("Bog", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricHomyak", []float64{4, 4, 5, 5, 6, 8}, 1, now32),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
