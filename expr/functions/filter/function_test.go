package filter

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

func TestConstantLine(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"filterSeries(metric[123], 'max', '>', 5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
				types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
			},
		},
		{
			"filterSeries(metric[123], 'max', '=', 5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
			},
		},
		{
			"filterSeries(metric[123], 'max', '!=', 6)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
			},
		},
		{
			"filterSeries(metric[123], 'max', '<', 6)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
			},
		},
		{
			"filterSeries(metric[123], 'max', '>=', 5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
				types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
				types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
			},
		},
		{
			"filterSeries(metric[123], 'max', '<=', 5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
					types.MakeMetricData("metric2", []float64{2.0, math.NaN(), 3.0, math.NaN(), 5.0, 6.0}, 1, now32),
					types.MakeMetricData("metric3", []float64{3.0, math.NaN(), 4.0, 5.0, 6.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1.0, math.NaN(), 2.0, 3.0, 4.0, 5.0}, 1, now32),
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
