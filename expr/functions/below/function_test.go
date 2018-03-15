package below

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

func TestBelow(t *testing.T) {
	now32 := int32(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("currentAbove",
				"metric1",
				7,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB",
				[]float64{3, 4, 5, 6, 7, 8}, 1, now32)},
		},
		{
			parser.NewExpr("currentBelow",
				"metric1",
				0,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, math.NaN()}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{0, 4, 4, 5, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA",
				[]float64{0, 0, 0, 0, 0, math.NaN()}, 1, now32)},
		},
		{
			parser.NewExpr("averageAbove",
				"metric1",
				5,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
			},
		},
		{
			parser.NewExpr("averageBelow",
				"metric1",
				0,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{0, 4, 4, 5, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA",
				[]float64{0, 0, 0, 0, 0, 0}, 1, now32)},
		},
		{
			parser.NewExpr("maximumAbove",
				"metric1",
				6,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB",
				[]float64{3, 4, 5, 6, 7, 8}, 1, now32)},
		},
		{
			parser.NewExpr("maximumBelow",
				"metric1",
				5,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA",
				[]float64{0, 0, 0, 0, 0, 0}, 1, now32)},
		},
		{
			parser.NewExpr("minimumAbove",
				"metric1",
				1,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{2, 4, 4, 5, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricC",
				[]float64{2, 4, 4, 5, 5, 6}, 1, now32)},
		},
		{
			parser.NewExpr("minimumBelow",
				"metric1",
				-2,
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{-1, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{-2, 4, 4, 5, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricC",
				[]float64{-2, 4, 4, 5, 5, 6}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
