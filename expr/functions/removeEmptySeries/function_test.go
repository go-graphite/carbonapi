package removeEmptySeries

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
			"removeEmptySeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32).SetTag("removeEmptySeries", "0"),
				types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32).SetTag("removeEmptySeries", "0"),
			},
		},
		{
			"removeZeroSeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32).SetTag("removeZeroSeries", "0"),
			},
		},
		{
			"removeEmptySeries(metric*,0.00001)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32).SetTag("removeEmptySeries", "0.00001"),
				types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32).SetTag("removeEmptySeries", "0.00001"),
			},
		},
		{
			"removeZeroSeries(metric*,0.000001)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 30, math.NaN()}, 1, now32).SetTag("removeZeroSeries", "0.000001"),
			},
		},
		{
			"removeEmptySeries(metric*,0.8)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, -1, 7, 8, 20, 23, math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric4", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric5", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32).SetTag("removeEmptySeries", "0.8"),
				types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32).SetTag("removeEmptySeries", "0.8"),
				types.MakeMetricData("metric5", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32).SetTag("removeEmptySeries", "0.8"),
			},
		},
		{
			"removeZeroSeries(metric*,0.8)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, -1, 7, 8, 20, 23, math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric4", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric5", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32).SetTag("removeZeroSeries", "0.8"),
				types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, math.NaN(), math.NaN()}, 1, now32).SetTag("removeZeroSeries", "0.8"),
			},
		},
		{
			"removeEmptySeries(metric*,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32).SetTag("removeEmptySeries", "1"),
			},
		},
		{
			"removeZeroSeries(metric*,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, 0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, -1, 7, 8, 20, 23, 12, 8, -2.3}, 1, now32).SetTag("removeZeroSeries", "1"),
			},
		},
		{
			"removeEmptySeries(metric*,0.5)", // Verify that passing in empty series with an xFilesFactor does not result in an error
			nil,
			[]*types.MetricData{},
		},
		{
			"removeZeroSeries(metric*,0.5)", // Verify that passing in empty series with an xFilesFactor does not result in an error
			nil,
			[]*types.MetricData{},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
