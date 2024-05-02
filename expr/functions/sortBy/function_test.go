package sortBy

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/interfaces"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var (
	md  []interfaces.FunctionMetadata = New("")
	nan                               = math.NaN()
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			Target: "sortByTotal(metric1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 4, 4}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 4, 4}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			Target: "sortByMaxima(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{2, 2, 10, 5, 2, 2}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricC", []float64{2, 2, 10, 5, 2, 2}, 1, now32),
				types.MakeMetricData("metricB", []float64{5, 5, 5, 5, 5, 5}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			Target: "sortByMinima(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
			},
		},
		{
			Target: "sortBy(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 2, 3, 4, 5, 6}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricC", []float64{1, 2, 3, 4, 5, 6}, 1, now32),
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
			},
		},
		{
			Target: "sortBy(metric*, 'median')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
			},
		},
		{
			Target: "sortBy(metric*, 'max', true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
			},
		},
		{
			Target: "sortBy(metric*, 'max', true)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{nan, nan, nan, nan, nan, nan}, 1, now32),
					types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
					types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("metricB", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				types.MakeMetricData("metricC", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
				types.MakeMetricData("metricA", []float64{nan, nan, nan, nan, nan, nan}, 1, now32),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExpr(t, eval, &tt)
		})
	}

}

func TestErrorInvalidConsolidationFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			Target: "sortBy(metric*, 'test')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: 0, Until: 1}: {
					types.MakeMetricData("metricA", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("metricB", []float64{4, 4, 5, 5, 6, 6}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 4, 5, 6, 7, 8}, 1, now32),
				},
			},
			Want:  nil,
			Error: consolidations.ErrInvalidConsolidationFunc,
		},
	}

	for _, testCase := range tests {
		testName := testCase.Target
		t.Run(testName, func(t *testing.T) {
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExprWithError(t, eval, &testCase)
		})
	}
}
