package sortBy

import (
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
	md []interfaces.FunctionMetadata = New("")
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
<<<<<<< HEAD
			"sortByTotal(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortByTotal(metric1)",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortByMaxima(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortByMaxima(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortByMinima(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortByMinima(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortBy(metric*)",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortBy(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortBy(metric*, 'median')",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortBy(metric*, 'median')",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortBy(metric*, 'max', true)",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortBy(metric*, 'max', true)",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
<<<<<<< HEAD
			"sortBy(metric*, 'test')",
			map[parser.MetricRequest][]*types.MetricData{
=======
			Target: "sortBy(metric*, 'test')",
			M: map[parser.MetricRequest][]*types.MetricData{
>>>>>>> 6447e792 (Add field names to struct literal)
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
