package timeSlice

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

func TestTimeSlice(t *testing.T) {
	var startTime int64 = 0

	tests := []th.EvalTestItemWithRange{
		{
			Target: `timeSlice(metric1, "3m", "8m")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 11*60}: {types.MakeMetricData("metric1", []float64{math.NaN(), 1, 2, 3, math.NaN(), 5, 6, math.NaN(), 7, 8, 9}, 60, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeSlice(metric1,180,480)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 3, math.NaN(), 5, 6, math.NaN(), 7, math.NaN(), math.NaN()}, 60, startTime).SetTags(map[string]string{"name": "metric1", "timeSliceStart": "180", "timeSliceEnd": "480"})},
			From:  startTime,
			Until: startTime + 11*60,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
