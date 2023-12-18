package sortBy

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/consolidations"

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
			"sortByTotal(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 4, 4}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 4, 4}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			"sortByMaxima(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{2, 2, 10, 5, 2, 2}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricC", []float64{2, 2, 10, 5, 2, 2}, 1, now32),
				types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			"sortByMinima(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
			},
		},
		{
			"sortBy(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 2, 3, 4, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricC", []float64{1, 2, 3, 4, 5, 6}, 1, now32),
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
			},
		},
		{
			"sortBy(metric*, 'median')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
			},
		},

		{
			"sortBy(metric*, 'max', true)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
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

func TestErrorInvalidConsolidationFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			"sortBy(metric*, 'test')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
			},
			nil,
			consolidations.ErrInvalidConsolidationFunc,
		},
	}

	for _, testCase := range tests {
		testName := testCase.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithError(t, &testCase)
		})
	}
}
