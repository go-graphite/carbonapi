package timeSlice

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/date"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestTimeSlice(t *testing.T) {
	var startTime int64 = 0

	// 1994-08-16 12:00 UTC; today (UTC midnight) = 776995200.
	mockNow := time.Date(1994, time.August, 16, 12, 0, 0, 0, time.UTC)
	defer date.MockTimeNow(func() time.Time { return mockNow })()

	tests := []th.EvalTestItemWithRange{
		{
			// Old behavior, kept for retrocompatibility: parse intervals as absolute timestamps
			Target: `timeSlice(metric1, "3m", "8m")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 11*60}: {types.MakeMetricData("metric1", []float64{math.NaN(), 1, 2, 3, math.NaN(), 5, 6, math.NaN(), 7, 8, 9}, 60, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeSlice(metric1,180,480)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 3, math.NaN(), 5, 6, math.NaN(), 7, math.NaN(), math.NaN()}, 60, startTime).SetTags(map[string]string{"name": "metric1", "timeSliceStart": "180", "timeSliceEnd": "480"})},
			From:  startTime,
			Until: startTime + 11*60,
		},
		{
			Target: `timeSlice(metric1, "00:03 19700101", "00:08 19700101")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 11*60}: {types.MakeMetricData("metric1", []float64{math.NaN(), 1, 2, 3, math.NaN(), 5, 6, math.NaN(), 7, 8, 9}, 60, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeSlice(metric1,180,480)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 3, math.NaN(), 5, 6, math.NaN(), 7, math.NaN(), math.NaN()}, 60, startTime).SetTags(map[string]string{"name": "metric1", "timeSliceStart": "180", "timeSliceEnd": "480"})},
			From:  startTime,
			Until: startTime + 11*60,
		},
		{
			Target: `timeSlice(metric1, "today-1h", "today+1h")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 776988000, Until: 777002400}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 3600, 776988000)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeSlice(metric1,776991600,776998800)",
				[]float64{math.NaN(), 2, 3, 4, math.NaN()}, 3600, 776988000).SetTags(map[string]string{"name": "metric1", "timeSliceStart": "776991600", "timeSliceEnd": "776998800"})},
			From:  776988000,
			Until: 777002400,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExprWithRange(t, eval, &tt)
		})
	}
}
