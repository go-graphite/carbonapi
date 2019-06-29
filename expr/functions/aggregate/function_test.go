package aggregate

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

func TestAverageSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("avg"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avgSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("avg_zero"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 4, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avg_zeroSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 3, 5, 4}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("countSeries(metric[123])",
				[]float64{3, math.NaN(), 3, 2, 3, 2}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("diff"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("diffSeries(metric[123])",
				[]float64{-4, math.NaN(), -5, -2, -7, -1}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("last"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("lastSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("max"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("min"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("minSeries(metric[123])",
				[]float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("median"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("medianSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("multiply"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric[123])",
				[]float64{6, math.NaN(), 24, math.NaN(), 120, math.NaN()}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("range"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("rangeSeries(metric[123])",
				[]float64{2, math.NaN(), 2, 2, 2, 1}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("sum"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric[123])",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("stddev"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.816496580927726, math.NaN(), 0.816496580927726, 1, 0.816496580927726, 0.5}, 1, now32)},
		},
		{
			parser.NewExpr("aggregate",
				"metric[123]",
				parser.ArgValue("stddev"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.4714045207910317, 0.9428090415820634, 1.4142135623730951, 1.8856180831641267, 2.357022603955158}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
