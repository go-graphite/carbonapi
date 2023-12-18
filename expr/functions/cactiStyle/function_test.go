package cactiStyle

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

func TestCactiStyle(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{math.NaN(), 20531.733333333334, 20196.4, 17925.333333333332, 20950.4, 35168.13333333333, 19965.866666666665, 24556.4, 22266.4, 58039.86666666667}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:58.04k    Max:58.04k    Min:17.93k",
					[]float64{math.NaN(), 20531.733333333334, 20196.4, 17925.333333333332, 20950.4, 35168.13333333333, 19965.866666666665, 24556.4, 22266.4, 58039.86666666667}, 1, now32).
					SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{1.432729, 1.434207, 1.404762, 1.414609, 1.399159, 1.411343, 1.406217, 1.407123, 1.392078, math.NaN()}, 1, now32).SetNameTag("metric1"),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:1.39    Max:1.43    Min:1.39",
					[]float64{1.432729, 1.434207, 1.404762, 1.414609, 1.399159, 1.411343, 1.406217, 1.407123, 1.392078, math.NaN()}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\",\"carrot\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{1.432729, 1.434207, 1.404762, 1.414609, 1.399159, 1.411343, 1.406217, 1.407123, 1.392078, math.NaN()}, 1, now32).SetNameTag("metric1"),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:1.39 carrot    Max:1.43 carrot    Min:1.39 carrot",
					[]float64{1.432729, 1.434207, 1.404762, 1.414609, 1.399159, 1.411343, 1.406217, 1.407123, 1.392078, math.NaN()}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{math.NaN(), 88364212.53333333, 79008410.93333334, 80312920.0, 69860465.2, 83876830.0, 80399148.8, 90481297.46666667, 79628113.73333333, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:79.63M    Max:90.48M    Min:69.86M",
					[]float64{math.NaN(), 88364212.53333333, 79008410.93333334, 80312920.0, 69860465.2, 83876830.0, 80399148.8, 90481297.46666667, 79628113.73333333, math.NaN()}, 1, now32).
					SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{1000}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:1.00k    Max:1.00k    Min:1.00k", []float64{1000}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{1000}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:1000    Max:1000    Min:1000", []float64{1000}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,units=\"apples\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{10}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:10 apples    Max:10 apples    Min:10 apples", []float64{10}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:240.00    Max:240.00    Min:240.00",
					[]float64{240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, 240.0, math.NaN()}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{-1.0, -2.0, -1.0, -3.0, -1.0, -1.0, -0.0, -0.0, -0.0}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:0.00    Max:0.00   Min:-3.00",
					[]float64{-1.0, -2.0, -1.0, -3.0, -1.0, -1.0, -0.0, -0.0, -0.0}, 1, now32).SetNameTag("metric1"),
			},
		},
		{
			"cactiStyle(metric1,\"si\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1 Current:NaN    Max:NaN    Min:NaN",
					[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32).SetNameTag("metric1"),
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
