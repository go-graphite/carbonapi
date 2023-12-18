package slo

import (
	"math"
	"testing"

	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestSlo(t *testing.T) {
	nan := math.NaN()
	now32 := int64(1615737710)

	testCases := []th.EvalTestItem{
		{
			"slo(x.y.z, \"10sec\", \"above\", 2)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, 2, 3, 4, 5, nan, nan, 6, 7, 0, 8},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"slo(x.y.z, 10sec, above, 2)",
					// (1, 2) -> 0
					// (3, 4) -> 1
					// (5, nan) -> 1: all not-null elements are above 2
					// (nan, 6) -> 1: the same
					// (7, 0) -> 0.5: only 1 element of 2 is above 2
					// (8) -> 1
					[]float64{0, 1, 1, 1, 0.5, 1},
					10,
					now32,
				),
			},
		},
		{
			"slo(x.y.z, \"4sec\", \"below\", 6)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"slo(x.y.z, 4sec, below, 6)",
					// all data points are nan because interval (4 sec) is less than step time (5 sec)
					[]float64{nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan},
					4,
					now32,
				),
			},
		},
	}

	for _, testCase := range testCases {
		testName := testCase.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase)
		})
	}
}

func TestSloErrorBudget(t *testing.T) {
	nan := math.NaN()
	now32 := int64(1615737710)

	testCases := []th.EvalTestItem{
		{
			"sloErrorBudget(some.data.series, \"5sec\", \"aboveOrEqual\", 2, 0.6)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "some.data.series",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"some.data.series",
						[]float64{
							1, 1.5, 2, 3, 4, // 3 of 5 points are greater or equal than 2
							nan, 0, 1, 1.5, 2.1, // 1 of 4 points is greater or equal than 2
							1, 2, 3, 4, 5, // 4 of 5 points are greater or equal than 2
							1, 2, 3, 4, // 3 of 4 points are greater or equal than 2
						},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"sloErrorBudget(some.data.series, 5sec, aboveOrEqual, 2, 0.6)",
					[]float64{
						0,     // 3 of 5 points match, slo is 0.6, no error budget remains
						-1.75, // 1 of 4 points match, slo is 0.6, error budget is exceeded by (0.25 - 0.6) * 5 = -1.75
						1,     // 4 of 5 points match, slo is 0.6, amount of remained budget is (0.8 - 0.6) * 5 = 1
						0.6,   // 3 of 4 points match, slo is 0.6, amount of remained budget is (0.75 - 0.6) * 4 = 0.6
					},
					5,
					now32,
				),
			},
		},
		{
			"sloErrorBudget(some.data.series, \"4sec\", \"aboveOrEqual\", 2, 0.6)",
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "some.data.series",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"some.data.series",
						[]float64{
							1, 1.5, 2, 3, 4,
							nan, 0, 1, 1.5, 2.1,
							1, 2, 3, 4, 5,
							1, 2, 3, 4,
						},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"sloErrorBudget(some.data.series, 4sec, aboveOrEqual, 2, 0.6)",
					[]float64{
						// all data points are nan because interval (4 sec) is less than step time (5 sec)
						nan, nan, nan, nan, nan, nan, nan, nan,
						nan, nan, nan, nan, nan, nan, nan, nan,
						nan, nan, nan, nan, nan, nan, nan, nan,
					},
					4,
					now32,
				),
			},
		},
	}

	for _, testCase := range testCases {
		testName := testCase.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &testCase)
		})
	}
}
