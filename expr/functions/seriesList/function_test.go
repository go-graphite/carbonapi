package seriesList

import (
	"math"
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
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}

}
