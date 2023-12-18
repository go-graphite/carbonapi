package fallbackSeries

import (
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

func TestFallbackSeries(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"fallbackSeries(metric*,fallbackmetric)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}:        {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"fallbackmetric", 0, 1}: {types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32),
			},
		},
		{
			"fallbackSeries(metric1,metrc2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32),
			},
		},
		{
			"fallbackSeries(absentmetric,fallbackmetric)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}:        {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
				{"fallbackmetric", 0, 1}: {types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("fallbackmetric", []float64{0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7}, 1, now32),
			},
		},
		{
			"fallbackSeries(metric1,metrc2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32)},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}, 1, now32),
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

func TestErrorMissingTimeSeriesFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			"fallbackSeries(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
			},
			nil,
			parser.ErrMissingTimeseries,
		},
	}

	for _, testCase := range tests {
		testName := testCase.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithError(t, &testCase)
		})
	}
}
