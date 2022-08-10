package divideSeries

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

func TestDivideSeriesMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

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
	now32 := int64(time.Now().Unix())

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
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestDivideSeriesAligned(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"divideSeries(metric1,metric2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12, 2}, 1, now32),
				},
				{"metric2", 0, 1}: {
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric1,metric2)",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2, math.NaN()}, 1, now32)},
		},
		{
			"divideSeries(metric[23])",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[23]", 0, 1}: {
					types.MakeMetricData("metric2", []float64{1, math.NaN(), math.NaN(), 3, 4, 12, 2}, 1, now32),
					types.MakeMetricData("metric3", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("divideSeries(metric[23])",
				[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2, math.NaN()}, 1, now32).SetNameTag("metric2")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
