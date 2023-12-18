package seriesList

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"diffSeriesLists(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("diffSeries(metric1,metric2)",
				[]float64{-1, math.NaN(), math.NaN(), math.NaN(), 4, 6}, 1, now32)},
		},
		{
			"sumSeriesLists(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{3, math.NaN(), math.NaN(), math.NaN(), 4, 18}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestSeriesListMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"divideSeriesLists(metric[12],metric[12])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"divideSeriesListSameGroups",
			map[string][]*types.MetricData{
				"divideSeries(metric1,metric1)": {types.MakeMetricData("divideSeries(metric1,metric1)", []float64{1, 1, 1, 1, 1}, 1, now32)},
				"divideSeries(metric2,metric2)": {types.MakeMetricData("divideSeries(metric2,metric2)", []float64{1, 1, 1, 1, 1}, 1, now32)},
			},
		},
		{
			"multiplySeriesLists(metric[12],metric[12])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"multiplySeriesListSameGroups",
			map[string][]*types.MetricData{
				"multiplySeries(metric1,metric1)": {types.MakeMetricData("multiplySeries(metric1,metric1)", []float64{1, 4, 9, 16, 25}, 1, now32)},
				"multiplySeries(metric2,metric2)": {types.MakeMetricData("multiplySeries(metric2,metric2)", []float64{4, 16, 36, 64, 100}, 1, now32)},
			},
		},
		{
			"diffSeriesLists(metric[12],metric[12])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"diffSeriesListSameGroups",
			map[string][]*types.MetricData{
				"diffSeries(metric1,metric1)": {types.MakeMetricData("diffSeries(metric1,metric1)", []float64{0, 0, 0, 0, 0}, 1, now32)},
				"diffSeries(metric2,metric2)": {types.MakeMetricData("diffSeries(metric2,metric2)", []float64{0, 0, 0, 0, 0}, 1, now32)},
			},
		},
		{
			"diffSeriesLists(metric[12],metric[134])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
				{"metric[134]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{2, 4, 6, 8, 10}, 1, now32),
					types.MakeMetricData("metric4", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"diffSeriesListSameGroups",
			map[string][]*types.MetricData{
				"diffSeries(metric1,metric1)": {types.MakeMetricData("diffSeries(metric1,metric1)", []float64{0, 0, 0, 0, 0}, 1, now32)},
			},
		},
		{
			"sumSeriesLists(metric[12],metric[12])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"sumSeriesListSameGroups",
			map[string][]*types.MetricData{
				"sumSeries(metric1,metric1)": {types.MakeMetricData("sumSeries(metric1,metric1)", []float64{2, 4, 6, 8, 10}, 1, now32)},
				"sumSeries(metric2,metric2)": {types.MakeMetricData("sumSeries(metric2,metric2)", []float64{4, 8, 12, 16, 20}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}

}

func TestDivideSeriesMismatchedData(t *testing.T) {
	var startTime int64 = 0

	tests := []th.EvalTestItemWithRange{
		{
			Target: `divideSeriesLists(metric1,metric2)`, // Test different step values for metrics
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, startTime)},
				{Metric: "metric2", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 2, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("divideSeries(metric1,metric2)",
				[]float64{1.5, 1.75, 1.6666666666666667, math.NaN(), math.NaN()}, 2, startTime)},
			From:  startTime,
			Until: startTime + 1,
		},
		{
			Target: `divideSeriesLists(metricA,metricB)`, // Test different step values for metrics
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metricA", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricA", []float64{1, 2, 3, 4, 5}, 10, startTime)},
				{Metric: "metricB", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricB", []float64{1, 2, 3, 4, 5}, 5, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("divideSeries(metricA,metricB)",
				[]float64{0.6666666666666666, 0.5714285714285714, 0.6, math.NaN(), math.NaN()}, 10, startTime)},
			From:  startTime,
			Until: startTime + 1,
		},
		{
			Target: `divideSeriesLists(metricC,metricD)`, // Test different number of values for metrics
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metricC", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricC", []float64{1, 2, 3, 4}, 1, startTime)},
				{Metric: "metricD", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricD", []float64{1, 2, 3, 4, 5}, 1, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("divideSeries(metricC,metricD)",
				[]float64{1, 1, 1, 1, math.NaN()}, 1, startTime)},
			From:  startTime,
			Until: startTime + 1,
		},
		{
			Target: `divideSeriesLists(metricE,metricF)`, // Test different number of values and steps in metrics
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metricE", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricE", []float64{1, 2, 3, 4, 5}, 2, startTime)},
				{Metric: "metricF", From: startTime, Until: startTime + 1}: {types.MakeMetricData("metricF", []float64{1, 2, 3, 4}, 1, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("divideSeries(metricE,metricF)",
				[]float64{0.6666666666666666, 0.5714285714285714, math.NaN(), math.NaN(), math.NaN()}, 2, startTime)},
			From:  startTime,
			Until: startTime + 1,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
