package aggregate

import (
	"context"
	"math"
	"testing"
	"time"

	fconfig "github.com/go-graphite/carbonapi/expr/functions/config"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"github.com/go-graphite/carbonapi/tests/compare"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestAverageSeries(t *testing.T) {
	fconfig.Config.ExtractTagsFromArgs = false

	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`aggregate(metric[123], "avg")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {},
			},
			[]*types.MetricData{},
		},
		{
			`aggregate(metric[123], "avg")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avgSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32).SetTag("aggregatedBy", "avg")},
		},
		{
			`aggregate(metric[123], "avg_zero")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 4, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avg_zeroSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 3, 5, 4}, 1, now32).SetTag("aggregatedBy", "avg_zero")},
		},
		{
			`aggregate(metric[123], "count")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("countSeries(metric[123])",
				[]float64{3, math.NaN(), 3, 2, 3, 2}, 1, now32).SetTag("aggregatedBy", "count")},
		},
		{
			`aggregate(metric[123], "diff")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("diffSeries(metric[123])",
				[]float64{-4, math.NaN(), -5, -2, -7, -1}, 1, now32).SetTag("aggregatedBy", "diff")},
		},
		{
			`aggregate(metric[123], "last")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("lastSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32).SetTag("aggregatedBy", "last")},
		},
		{
			`aggregate(metric[123], "current")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("currentSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32).SetTag("aggregatedBy", "current")},
		},
		{
			`aggregate(metric[123], "max")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32).SetTag("aggregatedBy", "max")},
		},
		{
			`aggregate(metric[123], "min")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("minSeries(metric[123])",
				[]float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32).SetTag("aggregatedBy", "min")},
		},
		{
			`aggregate(metric[123], "median")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("medianSeries(metric[123])",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32).SetTag("aggregatedBy", "median")},
		},
		{
			`aggregate(metric[123], "multiply")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("multiplySeries(metric[123])",
				[]float64{6, math.NaN(), 24, math.NaN(), 120, math.NaN()}, 1, now32).SetTag("aggregatedBy", "multiply")},
		},
		{
			`aggregate(metric[123], "range")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("rangeSeries(metric[123])",
				[]float64{2, math.NaN(), 2, 2, 2, 1}, 1, now32).SetTag("aggregatedBy", "range")},
		},
		{
			`aggregate(metric[123], "rangeOf")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("rangeOfSeries(metric[123])",
				[]float64{2, math.NaN(), 2, 2, 2, 1}, 1, now32).SetTag("aggregatedBy", "rangeOf")},
		},
		{
			`aggregate(metric[123], "sum")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric[123])",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTag("aggregatedBy", "sum")},
		},
		{
			`aggregate(metric[123], "total")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("totalSeries(metric[123])",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTag("aggregatedBy", "total")},
		},
		{
			`aggregate(metric[123], "avg", 0.7)`, // Test with xFilesFactor
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, math.NaN(), 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("avgSeries(metric[123])",
				[]float64{2, math.NaN(), 3, math.NaN(), 5, math.NaN()}, 1, now32).SetTag("aggregatedBy", "avg")},
		},
		{
			`aggregate(metric[123], "sum", 0.5)`, // Test with xFilesFactor
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, math.NaN()}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric[123])",
				[]float64{6, math.NaN(), 9, 8, 15, math.NaN()}, 1, now32).SetTag("aggregatedBy", "sum")},
		},
		{
			`aggregate(metric[123], "max", 0.3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, math.NaN(), 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric[123])",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32).SetTag("aggregatedBy", "max")},
		},
		{
			`stddevSeries(metric[123])`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.816496580927726, math.NaN(), 0.816496580927726, 1, 0.816496580927726, 0.5}, 1, now32).SetTag("aggregatedBy", "stddev")},
		},
		{
			`stddevSeries(metric1,metric2,metric3)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric1,metric2,metric3)",
				[]float64{0.4714045207910317, 0.9428090415820634, 1.4142135623730951, 1.8856180831641267, 2.357022603955158}, 1, now32).SetTag("aggregatedBy", "stddev")},
		},
		{
			`aggregate(metric[123], "stddev")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 6}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 5}, 1, now32),
					types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.816496580927726, math.NaN(), 0.816496580927726, 1, 0.816496580927726, 0.5}, 1, now32).SetTag("aggregatedBy", "stddev")},
		},
		{
			`aggregate(metric[123], "stddev")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[123]", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 4, 6, 8, 10}, 1, now32),
					types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("stddevSeries(metric[123])",
				[]float64{0.4714045207910317, 0.9428090415820634, 1.4142135623730951, 1.8856180831641267, 2.357022603955158}, 1, now32).SetTag("aggregatedBy", "stddev")},
		},

		// sum
		{
			`sum(metric1,metric2)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0, -1, 2, -3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{0, 1, -2, 3, -4, -5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{0, 0, 0, 0, 0, 0}, 1, now32).SetTag("aggregatedBy", "sum")},
		},
		{
			"sum(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, math.NaN()}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 3, math.NaN(), 5, 6, math.NaN()}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, 4, 5, 6, math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2,metric3)", []float64{6, 9, 8, 15, 11, math.NaN()}, 1, now32).SetTag("aggregatedBy", "sum")},
		},
		{
			"sum(metric1,metric2,metric3,metric4)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5, math.NaN()}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, 3, math.NaN(), 5, 6, math.NaN()}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, 4, 5, 6, math.NaN(), math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2,metric3)", []float64{6, 9, 8, 15, 11, math.NaN()}, 1, now32).SetTag("aggregatedBy", "sum")},
		},

		// minMax
		{
			"maxSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("maxSeries(metric1,metric2,metric3)",
				[]float64{3, math.NaN(), 4, 5, 6, 6}, 1, now32).SetTag("aggregatedBy", "max")},
		},
		{
			"minSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("minSeries(metric1,metric2,metric3)",
				[]float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32).SetTag("aggregatedBy", "min")},
		},

		// avg
		{
			"averageSeries(metric1,metric2,metric3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				{"metric3", 0, 1}: {types.MakeMetricData("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("averageSeries(metric1,metric2,metric3)",
				[]float64{2, math.NaN(), 3, 4, 5, 5.5}, 1, now32).SetTag("aggregatedBy", "average")},
		},

		// sumSeies with tags, common tags to all metrics
		{
			"sum(seriesByTag('tag2=value*', 'name=metric'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value*', 'name=metric')", 0, 1}: {
					// No tags in common
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric;tag2=value23;tag3=value3;tag4=val4", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2=value*', 'name=metric'))",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTags(map[string]string{"aggregatedBy": "sum", "name": "metric"})},
		},
		{
			"sum(seriesByTag('tag2!=value2*', 'name=metric.name'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2!=value2*', 'name=metric.name')", 0, 1}: {
					// One tag in common
					types.MakeMetricData("metric.name;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric.name;tag2=value23;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric.name", 0, 1}: {types.MakeMetricData("metric.name", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2!=value2*', 'name=metric.name'))",
				// []float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"name": "metric.name"})},
				[]float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"aggregatedBy": "sum", "name": "metric.name", "tag3": "value3"})},
		},
		{
			"sum(seriesByTag('tag2=value21'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value21')", 0, 1}: {
					// All tags in common
					types.MakeMetricData("metric;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric;tag2=value22;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2=value21'))",
				// []float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"name": "sumSeries", "tag2": "value21"})},
				[]float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"aggregatedBy": "sum", "name": "metric", "tag2": "value22", "tag3": "value3"})},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func TestAverageSeriesExtractSeriesByTag(t *testing.T) {
	fconfig.Config.ExtractTagsFromArgs = true

	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		// sumSeies with tags, tags from first metric
		{
			"sum(seriesByTag('tag2=value*', 'name=metric'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value*', 'name=metric')", 0, 1}: {
					types.MakeMetricData("metric;tag1=value1;tag2=value21", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric;tag2=value23;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2=value*', 'name=metric'))",
				[]float64{6, math.NaN(), 9, 8, 15, 11}, 1, now32).SetTags(map[string]string{"name": "metric", "tag2": "value*", "aggregatedBy": "sum"})},
		},
		{
			"sum(seriesByTag('tag2!=value2*', 'name=metric.name'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2!=value2*', 'name=metric.name')", 0, 1}: {
					types.MakeMetricData("metric.name;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric.name;tag2=value23;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric.name", 0, 1}: {types.MakeMetricData("metric.name", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2!=value2*', 'name=metric.name'))",
				[]float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"name": "metric.name", "aggregatedBy": "sum"})},
		},
		{
			"sum(seriesByTag('tag2=value21'))",
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('tag2=value21')", 0, 1}: {
					types.MakeMetricData("metric;tag2=value22;tag3=value3", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric;tag2=value23;tag3=value3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
				},
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{2, math.NaN(), 3, math.NaN(), 5, 11}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(seriesByTag('tag2=value21'))",
				[]float64{5, math.NaN(), 7, 5, 11, 6}, 1, now32).SetTags(map[string]string{"name": "sumSeries", "tag2": "value21", "aggregatedBy": "sum"})},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}

func TestAverageSeriesAlign(t *testing.T) {
	tests := []th.EvalTestItem{
		// Indx      |   0  |   1  |   2  |   3  |   4  |
		// commonStep  2
		// Start  0 (1 - 1 % 2)
		// metric1_2 |      |   1  |   3  |   5  |      |
		// metric1_2 |   1  |      |   4  |      |      |
		//
		// metric2_1 |      |   1  |      |   5  |      |
		// metric2_1 |   1  |      |   5  |      |      |

		// sum       |   2  |      |   9  |      |      |
		{
			// timeseries with different length
			Target: "sum(metric1_2,metric2_1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1_2", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 3, 5}, 1, 1)},
				{"metric2_1", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 5}, 2, 1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("sumSeries(metric1_2,metric2_1)",
				[]float64{2, 9}, 2, 0).SetTag("aggregatedBy", "sum")},
		},
		{
			// First timeseries with broker StopTime
			Target: "sum(metric1,metric2)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 3, 5, 8}, 1, 1).AppendStopTime(-1)},
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 5, 7}, 1, 1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("sumSeries(metric1,metric2)",
				[]float64{2, 8, 12, 8}, 1, 1).SetTag("aggregatedBy", "sum")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprResult(t, &tt)
		})
	}

}

func BenchmarkAverageSeries(b *testing.B) {
	target := "sum(metric*)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{"metric*", 0, 1}: {
			types.MakeMetricData("metric2", compare.GenerateMetrics(2046, 1, 10, 1), 2, 1),
			types.MakeMetricData("metric1", compare.GenerateMetrics(4096, 1, 10, 1), 1, 1),
			types.MakeMetricData("metric3", compare.GenerateMetrics(1360, 1, 10, 1), 3, 1),
		},
	}

	evaluator := metadata.GetEvaluator()
	exp, _, err := parser.ParseExpr(target)
	if err != nil {
		b.Fatalf("failed to parse %s: %+v", target, err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g, err := evaluator.Eval(context.Background(), exp, 0, 1, metrics)
		if err != nil {
			b.Fatalf("failed to eval %s: %+v", target, err)
		}
		_ = g
	}
}
