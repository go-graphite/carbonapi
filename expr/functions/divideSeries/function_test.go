package divideSeries

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

func TestDivideSeriesMultiReturn(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []th.MultiReturnEvalTestItem{
		{
			"divideSeries(metric[12],metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
				{"metric2", 0, 1}: {
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
				},
			},
			"divideSeries",
			map[string][]*types.MetricData{
				"divideSeries(metric1,metric2)": {types.MakeMetricData("divideSeries(metric1,metric2)", []float64{0.5, 0.5, 0.5, 0.5, 0.5}, 1, now32)},
				"divideSeries(metric2,metric2)": {types.MakeMetricData("divideSeries(metric2,metric2)", []float64{1, 1, 1, 1, 1}, 1, now32)},
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

func TestDivideSeries(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"divideSeries(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric1,metric2)",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2}, 1, now32)},
		},
		{
			"divideSeries(metric[12])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[12]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric[12])",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2}, 1, now32).SetNameTag("metric1")},
		},
		{
			"divideSeries(testMetric,metric)", // verify that a non-existant denominator will not error out and instead will return a list of math.NaN() values
			map[parser.MetricRequest][]*types.MetricData{
				{"testMetric", 0, 1}: {types.MakeMetricData("testMetric", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(testMetric,MISSING)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestDivideSeriesAligned(t *testing.T) {
	startTime := int64(0)

	tests := []th.EvalTestItemWithRange{
		{
			"divideSeries(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12, 2}, 1, startTime),
				},
				{"metric2", 0, 1}: {
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, startTime),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric1,metric2)",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2, math.NaN()}, 1, startTime)},
			startTime,
			startTime + 1,
		},
		{
			"divideSeries(metric[23])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[23]", 0, 1}: {
					types.MakeMetricData("metric2", []float64{1, math.NaN(), math.NaN(), 3, 4, 12, 2}, 1, startTime),
					types.MakeMetricData("metric3", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, startTime),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric[23])",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2, math.NaN()}, 1, startTime).SetNameTag("metric2")},
			startTime,
			startTime + 1,
		},
		{
			"divideSeries(metric3,metric4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric3", 0, 1}: {
					types.MakeMetricData("metric3", []float64{1, math.NaN(), math.NaN(), 3, 4, 8, 2, math.NaN(), 3, math.NaN(), 0, 6}, 5, startTime),
				},
				{"metric4", 0, 1}: {
					types.MakeMetricData("metric4", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 10, startTime),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric3,metric4)",
				[]float64{0.5, math.NaN(), 2, math.NaN(), math.NaN(), 0.5}, 10, startTime)},
			startTime,
			startTime + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Target, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
