package interpolate

import (
	"math"
	"testing"
	"time"

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

func TestInterpolate_Do(t *testing.T) {
	nan := math.NaN()
	now32 := time.Now().Unix()

	testCases := []th.EvalTestItem{
		{
			"interpolate(x1.y1.z1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x1.y1.z1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x1.y1.z1",
						[]float64{1, 2, 3, 4, nan, nan, nan, 6, 7, 8},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"interpolate(x1.y1.z1)",
					[]float64{1, 2, 3, 4, 4.5, 5, 5.5, 6, 7, 8},
					1,
					now32,
				),
			},
		},
		{
			"interpolate(x1.y1.z1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x1.y1.z1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x1.y1.z1",
						[]float64{1, 2, 3, 4, 5, nan, nan, 8, 9, 10},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"interpolate(x1.y1.z1)",
					[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
					1,
					now32,
				),
			},
		},
		{
			"interpolate(x1.y1.z1, 2)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x1.y1.z1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x1.y1.z1",
						[]float64{1, 2, 3, 4, nan, nan, nan, 6, 7, 8},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"interpolate(x1.y1.z1)",
					[]float64{1, 2, 3, 4, nan, nan, nan, 6, 7, 8}, // limit is 2 gaps, have 3, cannot interpolate
					1,
					now32,
				),
			},
		},
		{
			"interpolate(x1.y1.z1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x1.y1.z1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x1.y1.z1",
						[]float64{nan, nan, nan, 1, 2, 3},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"interpolate(x1.y1.z1)",
					[]float64{nan, nan, nan, 1, 2, 3}, // there are not values before the gap
					1,
					now32,
				),
			},
		},
		{
			"interpolate(x1.y1.z1, inf)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x1.y1.z1",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x1.y1.z1",
						[]float64{1, 2, 3, 4, nan, nan, nan, 6, 7, 8},
						1,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					"interpolate(x1.y1.z1)",
					[]float64{1, 2, 3, 4, 4.5, 5, 5.5, 6, 7, 8},
					1,
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
