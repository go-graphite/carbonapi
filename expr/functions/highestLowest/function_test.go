package highestLowest

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

func TestHighestMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"highestCurrent(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			"highestCurrent",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32)},
			},
		},
		{
			"highestCurrent(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			"highestCurrent",
			map[string][]*types.MetricData{
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32)},
			},
		},
		{
			"highestMax(metric1, 2)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12, 9}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 5, 5, 5, 5, 5, 3}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10, 10}, 1, now32),
				},
			},
			"highestMax",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12, 9}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10, 10}, 1, now32)},
			},
		},
		{
			"highestMin(metric1, 2)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{6, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 2, 3, 3, 4, 10}, 1, now32),
				},
			},
			"highestMin",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{3, 2, 3, 3, 4, 10}, 1, now32)},
			},
		},
		{
			"highest(metric1, 2, \"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			"highest",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32)},
			},
		},
		{
			"lowest(metric1, 2, \"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			"lowest",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32)},
				"metricC": {types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32)},
			},
		},

		{
			"lowestCurrent(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
					types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				},
			},
			"lowestCurrent",
			map[string][]*types.MetricData{
				"metricA": {types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32)},
				"metricB": {types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32)},
				"metricD": {types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32)},
			},
		},
		{
			"lowestCurrent(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{"metric1", 0, 1}: {
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
					types.MakeMetricData("metricD", []float64{1, 1, 3, 3, 4, 3}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				},
			},
			"lowestCurrent",
			map[string][]*types.MetricData{
				"metricB": {types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32)},
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

func TestHighest(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"highestCurrent(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricC", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 1, 3, 3, 4, 15}, 1, now32)},
		},
		{
			"highestCurrent(metric1,4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
				types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
				//NOTE(nnuss): highest* functions filter null-valued series as a side-effect when `n` >= number of series
				//TODO(nnuss): bring lowest* functions into harmony with this side effect or get rid of it
				//types.MakeMetricData("metric0", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
			},
		},
		{
			"highestAverage(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 5, 5, 5, 5, 5}, 1, now32)},
		},
		{
			"highestMax(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 1, 3, 3, 12, 11}, 1, now32)},
		},
		{
			"highestMin(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{6, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{2, 5, 5, 5, 5, 5}, 1, now32),
					types.MakeMetricData("metricC", []float64{3, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB", // NOTE(dgryski): not sure if this matches graphite
				[]float64{2, 5, 5, 5, 5, 5}, 1, now32)},
		},
		{
			"highest(metric1,\"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricA",
				[]float64{1, 1, 3, 3, 12, 11}, 1, now32)},
		},
		{
			"lowest(metric1,\"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 12, 11}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 10}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
			},
		},
		{
			"lowestCurrent(metric1,1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metricA", []float64{1, 1, 3, 3, 4, 12}, 1, now32),
					types.MakeMetricData("metricB", []float64{1, 1, 3, 3, 4, 1}, 1, now32),
					types.MakeMetricData("metricC", []float64{1, 1, 3, 3, 4, 15}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("metricB", // NOTE(dgryski): not sure if this matches graphite
				[]float64{1, 1, 3, 3, 4, 1}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
