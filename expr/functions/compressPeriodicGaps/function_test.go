package compressPeriodicGaps

import (
	"math"
	"testing"

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

func TestCompressPeriodicGaps(t *testing.T) {
	var startTime int64 = 100

	tests := []th.EvalTestItemWithRange{
		{
			Target: `compressPeriodicGaps(metric*)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric*", From: startTime, Until: startTime + 160}: {
					types.MakeMetricData("metric1", []float64{math.NaN(), 1, math.NaN(), math.NaN(), 2, math.NaN(), math.NaN(), 3, math.NaN(), math.NaN(), 4, math.NaN(), math.NaN(), 5, math.NaN(), math.NaN()}, 10, startTime),
					types.MakeMetricData("metric2", []float64{1, math.NaN(), math.NaN(), 2, math.NaN(), math.NaN(), 3, math.NaN(), math.NaN(), 4, math.NaN(), math.NaN(), 5, math.NaN(), math.NaN(), 6}, 10, startTime+10),
					types.MakeMetricData("metric3", []float64{math.NaN(), math.NaN(), 1, math.NaN(), math.NaN(), 2, math.NaN(), math.NaN(), 3, math.NaN(), math.NaN(), 4, math.NaN(), math.NaN(), 5, math.NaN()}, 10, startTime),
					types.MakeMetricData("metric4", []float64{math.NaN(), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13, math.NaN()}, 10, startTime),
					types.MakeMetricData("metric5", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 10, startTime),
					types.MakeMetricData("metric6", []float64{math.NaN(), 1, math.NaN(), 3, math.NaN(), 5, math.NaN(), 7, math.NaN(), 9, math.NaN(), 11, math.NaN(), 13, math.NaN(), 15}, 10, startTime),
					types.MakeMetricData("metric7", []float64{math.NaN(), 1, 2, 3, math.NaN(), 5, math.NaN(), 7, math.NaN(), 9, math.NaN(), math.NaN(), math.NaN(), 13, math.NaN(), math.NaN()}, 10, startTime),
					types.MakeMetricData("metric8", []float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 13, 14, 15}, 10, startTime),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("compressPeriodicGaps(metric1)", []float64{1, 2, 3, 4, 5}, 30, startTime+10),
				types.MakeMetricData("compressPeriodicGaps(metric2)", []float64{1, 2, 3, 4, 5, 6}, 30, startTime+10),
				types.MakeMetricData("compressPeriodicGaps(metric3)", []float64{1, 2, 3, 4, 5}, 30, startTime+20),
				types.MakeMetricData("compressPeriodicGaps(metric4)", []float64{math.NaN(), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13, math.NaN()}, 10, startTime),
				types.MakeMetricData("compressPeriodicGaps(metric5)", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 10, startTime),
				types.MakeMetricData("compressPeriodicGaps(metric6)", []float64{1, 3, 5, 7, 9, 11, 13, 15}, 20, startTime+10),
				types.MakeMetricData("compressPeriodicGaps(metric7)", []float64{math.NaN(), 1, 2, 3, math.NaN(), 5, math.NaN(), 7, math.NaN(), 9, math.NaN(), math.NaN(), math.NaN(), 13, math.NaN(), math.NaN()}, 10, startTime),
				types.MakeMetricData("compressPeriodicGaps(metric8)", []float64{1, 2, 3, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), 13, 14, 15}, 10, startTime),
			},
			From:  startTime,
			Until: 260,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}

}
