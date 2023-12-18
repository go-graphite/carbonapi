package percentileOfSeries

import (
	"math"
	"testing"
	"time"

	fconfig "github.com/go-graphite/carbonapi/expr/functions/config"
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

func TestPercentileOfSeries(t *testing.T) {
	fconfig.Config.ExtractTagsFromArgs = false

	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`percentileOfSeries(metric1.empty,4)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.empty", 0, 1}: {},
			},
			[]*types.MetricData{},
		},
		{
			`percentileOfSeries(metric1,4)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("percentileOfSeries(metric1,4)", []float64{1, 1, 1, 1, 2, 2, 2, 4, 6, 4, 6, 8, math.NaN()}, 1, now32)},
		},
		{
			`percentileOfSeries(metric1.foo.*.*,50)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("percentileOfSeries(metric1.foo.*.*,50)", []float64{7, 8, 9, 10, 11, math.NaN()}, 1, now32)},
		},
		{
			`percentileOfSeries(metric1.foo.*.*,50,interpolate=true)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("percentileOfSeries(metric1.foo.*.*,50,interpolate=true)", []float64{6.5, 7.5, 8.5, 9.5, 11, math.NaN()}, 1, now32)},
		},
		{
			`percentileOfSeries(metric1.foo.*.*,95,false)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric1.foo.bar3.qux", []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32),
					types.MakeMetricData("metric1.foo.bar4.qux", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric1.foo.bar5.qux", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("percentileOfSeries(metric1.foo.*.*,95,false)", []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32)},
		},
		// seriesByTag
		{
			`percentileOfSeries(seriesByTag('tag2=value*', 'name=metric'),95,false)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value*', 'name=metric')", 0, 1}: {
					types.MakeMetricData("metric;tag1=value1;tag2=value21;tag3=value3", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric;tag2=value21", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData(`percentileOfSeries(seriesByTag('tag2=value*', 'name=metric'),95,false)`, []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32).
				SetTags(map[string]string{"name": "metric", "tag2": "value21"})},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func TestPercentileOfSeriesExtractSeriesByTag(t *testing.T) {
	fconfig.Config.ExtractTagsFromArgs = true

	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`percentileOfSeries(seriesByTag('tag2=value*', 'name=metric'),95,false)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value*', 'name=metric')", 0, 1}: {
					types.MakeMetricData("metric;tag1=value1;tag2=value21;tag3=value3", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
					types.MakeMetricData("metric;tag2=value22", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value23", []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value24", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
					types.MakeMetricData("metric;tag1=value1;tag2=value25", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 0}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData(`percentileOfSeries(seriesByTag('tag2=value*', 'name=metric'),95,false)`, []float64{0, 0, 0, 100500, 100501, 1005002}, 1, now32).
				SetTags(map[string]string{"name": "metric", "tag2": "value*"})},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
