package aggregateSeriesLists

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

var (
	now     = time.Now().Unix()
	shipped = []*types.MetricData{
		types.MakeMetricData("mining.other.shipped", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
		types.MakeMetricData("mining.diamond.shipped", []float64{0, -1, -1, 2, 3, -5, -8, 13, 21, -34, -55, 89, 144, -233, -377}, 1, now),
		types.MakeMetricData("mining.graphite.shipped", []float64{math.NaN(), 2.3, math.NaN(), -4.5, math.NaN(), 6.7, math.NaN(), -8.9, math.NaN(), 10.111, math.NaN(), -12.13, math.NaN(), 14.15, math.NaN(), -16.17, math.NaN(), 18.19, math.NaN(), -20.21}, 1, now),
		types.MakeMetricData("mining.carbon.shipped", []float64{3.141, math.NaN(), 2.718, 6.022, 6.674, math.NaN(), 6.626, 1.602, 2.067, 9.274, math.NaN(), 5.555}, 1, now),
	}
	extracted = []*types.MetricData{
		types.MakeMetricData("mining.other.extracted", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
		types.MakeMetricData("mining.diamond.extracted", []float64{0, 1, -1, 2, -3, 5, -8, 13, -21, 34, -55, 89, -144, 233, -377}, 1, now),
		types.MakeMetricData("mining.graphite.extracted", []float64{math.NaN(), 9, 8, math.NaN(), 6, math.NaN(), 4, 3, 2, math.NaN(), 0, -1, math.NaN(), -3, -4, -5, math.NaN(), math.NaN(), -8, -9, -10}, 1, now),
		types.MakeMetricData("mining.carbon.extracted", []float64{7.22, math.NaN(), 2.718, math.NaN(), 2.54, -1.234, -6.16, -13.37, math.NaN(), -7.77, 0.128, 8.912}, 1, now),
	}
)

func TestFunction(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\")", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\")", []float64{0.0, 0.0, -1.0, 2.0, 0.0, 0.0, -8.0, 13.0, 0.0, 0.0, -55.0, 89.0, 0.0, 0.0, -377.0}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\")", []float64{math.NaN(), 5.65, 8.0, -4.5, 6.0, 6.7, 4.0, -2.95, 2.0, 10.111, 0.0, -6.565, math.NaN(), 5.575, -4.0, -10.585, math.NaN(), 18.19, -8.0, -14.605, -10.0}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\")", []float64{5.1805, math.NaN(), 2.718, 6.022, 4.607, -1.234, 0.2330000000000001, -5.8839999999999995, 2.067, 0.7519999999999998, 0.128, 7.2335}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"sum\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"sum\")", []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"sum\")", []float64{0, 0, -2, 4, 0, 0, -16, 26, 0, 0, -110, 178, 0, 0, -754}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"sum\")", []float64{math.NaN(), 11.3, 8, -4.5, 6, 6.7, 4, -5.9, 2, 10.111, 0, -13.13, math.NaN(), 11.15, -4, -21.17, math.NaN(), 18.19, -8, -29.21, -10}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"sum\")", []float64{10.361, math.NaN(), 5.436, 6.022, 9.214, -1.234, 0.4660000000000002, -11.767999999999999, 2.067, 1.5039999999999996, 0.128, 14.467}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"diff\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"diff\")", []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"diff\")", []float64{0, -2, 0, 0, 6, -10, 0, 0, 42, -68, 0, 0, 288, -466, 0}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"diff\")", []float64{math.NaN(), -6.7, 8, -4.5, 6, 6.7, 4, -11.9, 2, 10.111, 0, -11.13, math.NaN(), 17.15, -4, -11.170000000000002, math.NaN(), 18.19, -8, -11.21, -10}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"diff\")", []float64{-4.079, math.NaN(), 0.0, 6.022, 4.134, -1.234, 12.786000000000001, 14.972, 2.067, 17.043999999999997, 0.128, -3.357000000000001}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"multiply\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"multiply\")", []float64{1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121, 144, 169, 196, 225, 256, 289, 324, 361, 400}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"multiply\")", []float64{0, -1, 1, 4, -9, -25, 64, 169, -441, -1156, 3025, 7921, -20736, -54289, 142129}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"multiply\")", []float64{math.NaN(), 20.7, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), -26.7, math.NaN(), math.NaN(), math.NaN(), 12.13, math.NaN(), -42.45, math.NaN(), 80.85, math.NaN(), math.NaN(), math.NaN(), 181.89, math.NaN()}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"multiply\")", []float64{22.67802, math.NaN(), 7.387524, math.NaN(), 16.95196, math.NaN(), -40.81616, -21.41874, math.NaN(), -72.05897999999999, math.NaN(), 49.50616}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"max\")", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"max\")", []float64{0, 1, -1, 2, 3, 5, -8, 13, 21, 34, -55, 89, 144, 233, -377}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"max\")", []float64{math.NaN(), 9, 8, -4.5, 6, 6.7, 4, 3, 2, 10.111, 0, -1, math.NaN(), 14.15, -4, -5, math.NaN(), 18.19, -8, -9, -10}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"max\")", []float64{7.22, math.NaN(), 2.718, 6.022, 6.674, -1.234, 6.626, 1.602, 2.067, 9.274, 0.128, 8.912}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"min\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"min\")", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"min\")", []float64{0, -1, -1, 2, -3, -5, -8, 13, -21, -34, -55, 89, -144, -233, -377}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"min\")", []float64{math.NaN(), 2.3, 8, -4.5, 6, 6.7, 4, -8.9, 2, 10.111, 0, -12.13, math.NaN(), -3, -4, -16.17, math.NaN(), 18.19, -8, -20.21, -10}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"min\")", []float64{3.141, math.NaN(), 2.718, 6.022, 2.54, -1.234, -6.16, -13.37, 2.067, -7.77, 0.128, 5.555}, 1, now),
			},
		},
		{
			"aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\", 0.6)", // Test with xFilesFactor
			map[parser.MetricRequest][]*types.MetricData{
				{"mining.*.shipped", 0, 1}:   shipped,
				{"mining.*.extracted", 0, 1}: extracted,
			},
			[]*types.MetricData{
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\", 0.6)", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\", 0.6)", []float64{0.0, 0.0, -1.0, 2.0, 0.0, 0.0, -8.0, 13.0, 0.0, 0.0, -55.0, 89.0, 0.0, 0.0, -377.0}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\", 0.6)", []float64{math.NaN(), 5.65, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), -2.95, math.NaN(), math.NaN(), math.NaN(), -6.565, math.NaN(), 5.575, math.NaN(), -10.585, math.NaN(), math.NaN(), math.NaN(), -14.605, math.NaN()}, 1, now),
				types.MakeMetricData("aggregateSeriesLists(mining.*.shipped, mining.*.extracted,\"avg\", 0.6)", []float64{5.1805, math.NaN(), 2.718, math.NaN(), 4.607, math.NaN(), 0.2330000000000001, -5.8839999999999995, math.NaN(), 0.7519999999999998, math.NaN(), 7.2335}, 1, now),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Target, func(t *testing.T) {
			th.TestEvalExpr(t, &test)
		})
	}
}
