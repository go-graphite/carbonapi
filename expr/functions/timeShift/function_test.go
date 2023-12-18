package timeShift

import (
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

func TestTimeShift(t *testing.T) {
	var startTime int64 = 86400

	tests := []th.EvalTestItemWithRange{
		// TODO(civil): Do not pass `true` resetEnd parameter in 0.15
		{
			Target: `timeShift(metric1, "0s", true)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 6}: {types.MakeMetricData("metric1", []float64{0, 1, 2, 3, 4, 5}, 1, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'0',true)",
				[]float64{0, 1, 2, 3, 4, 5}, 1, startTime).SetTag("timeshift", "0")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1s", false)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 1, Until: startTime + 5}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, startTime-1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-1',false)",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, startTime).SetTag("timeshift", "-1")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1s", true)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 1, Until: startTime + 5}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3}, 1, startTime-1)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-1',true)",
				[]float64{-1, 0, 1, 2, 3}, 1, startTime).SetTag("timeshift", "-1")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1h", false)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 60*60, Until: startTime - 60*60 + 6}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, startTime-60*60)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-3600',false)",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, startTime).SetTag("timeshift", "-3600")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1h", true)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 60*60, Until: startTime - 60*60 + 6}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, startTime-60*60)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-3600',true)",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, startTime).SetTag("timeshift", "-3600")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1d", false)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 86400, Until: startTime - 86400 + 6}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, startTime-86400)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-86400',false)",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, startTime).SetTag("timeshift", "-86400")},
			From:  startTime,
			Until: startTime + 6,
		},
		{
			Target: `timeShift(metric1, "1d", true)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime - 86400, Until: startTime - 86400 + 6}: {types.MakeMetricData("metric1", []float64{-1, 0, 1, 2, 3, 4}, 1, startTime-86400)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,'-86400',true)",
				[]float64{-1, 0, 1, 2, 3, 4}, 1, startTime).SetTag("timeshift", "-86400")},
			From:  startTime,
			Until: startTime + 6,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
