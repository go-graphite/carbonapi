package holtWintersAberration

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/holtwinters"

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

func TestHoltWintersAberration(t *testing.T) {
	var startTime int64 = 2678400
	var step int64 = 600
	var points int64 = 10
	var seconds int64 = 86400

	tests := []th.EvalTestItemWithRange{
		{
			Target: "holtWintersAberration(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", startTime, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, points*step, step, 0), step, startTime),
					types.MakeMetricData("metric2", generateHwRange(0, points*step, step, 10), step, startTime),
				},
				{"metric*", startTime - holtwinters.DefaultBootstrapInterval, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step, 0), step, startTime-holtwinters.DefaultBootstrapInterval),
					types.MakeMetricData("metric2", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step, 10), step, startTime-holtwinters.DefaultBootstrapInterval),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersAberration(metric1)", []float64{-0.2841206166091448, -0.05810270987744115, 0, 0, 0, 0, 0, 0, 0, 0}, step, startTime).SetTag("holtWintersAberration", "1"),
				types.MakeMetricData("holtWintersAberration(metric2)", []float64{-0.284120616609151, -0.05810270987744737, 0, 0, 0, 0, 0, 0, 0, 0}, step, startTime).SetTag("holtWintersAberration", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersAberration(metric*,4,'4d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", startTime, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, points*step, step, 0), step, startTime),
					types.MakeMetricData("metric2", generateHwRange(0, points*step, step, 10), step, startTime),
				},
				{"metric*", startTime - 4*seconds, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, ((4*seconds/step)+points)*step, step, 0), step, startTime-4*seconds),
					types.MakeMetricData("metric2", generateHwRange(0, ((4*seconds/step)+points)*step, step, 10), step, startTime-4*seconds),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersAberration(metric1)", []float64{-1.4410544085511923, -0.5199507849641569, 0, 0, 0, 0, 0, 0, 0, 0.09386319244056907}, step, startTime).SetTag("holtWintersAberration", "1"),
				types.MakeMetricData("holtWintersAberration(metric2)", []float64{-1.4410544085511923, -0.5199507849641609, 0, 0, 0, 0, 0, 0, 0, 0.09386319244056551}, step, startTime).SetTag("holtWintersAberration", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersAberration(metric*,4,'1d','2d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", startTime, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, points*step, step, 0), step, startTime),
					types.MakeMetricData("metric2", generateHwRange(0, points*step, step, 10), step, startTime),
				},
				{"metric*", startTime - seconds, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, ((seconds/step)+points)*step, step, 0), step, startTime-seconds),
					types.MakeMetricData("metric2", generateHwRange(0, ((seconds/step)+points)*step, step, 10), step, startTime-seconds),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersAberration(metric1)", []float64{-4.106587168490873, -2.8357974803355406, -1.5645896296885762, -0.4213549577359168, 0, 0, 0, 0.5073914761326588, 2.4432248533746543, 4.186719764193769}, step, startTime).SetTag("holtWintersAberration", "1"),
				types.MakeMetricData("holtWintersAberration(metric2)", []float64{-4.1065871684908775, -2.8357974803355486, -1.5645896296885837, -0.42135495773592346, 0, 0, 0, 0.5073914761326499, 2.4432248533746446, 4.186719764193759}, step, startTime).SetTag("holtWintersAberration", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersAberration(metric*)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", startTime, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, points*step, step, 0), step, startTime),
					types.MakeMetricData("metric2", generateHwRange(0, points*step, step, 10), step, startTime),
				},
				{"metric*", startTime - holtwinters.DefaultBootstrapInterval, startTime + step*points}: {
					types.MakeMetricData("metric1", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step, 0), step, startTime-holtwinters.DefaultBootstrapInterval),
					types.MakeMetricData("metric2", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step, 10), step, startTime-holtwinters.DefaultBootstrapInterval),
					types.MakeMetricData("metric3", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step, 20), step, startTime-holtwinters.DefaultBootstrapInterval), // Verify that metrics that don't match those fetched with the unadjusted start time are not included in the results
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersAberration(metric1)", []float64{-0.2841206166091448, -0.05810270987744115, 0, 0, 0, 0, 0, 0, 0, 0}, step, startTime).SetTag("holtWintersAberration", "1"),
				types.MakeMetricData("holtWintersAberration(metric2)", []float64{-0.284120616609151, -0.05810270987744737, 0, 0, 0, 0, 0, 0, 0, 0}, step, startTime).SetTag("holtWintersAberration", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}

func generateHwRange(x, y, jump, t int64) []float64 {
	var valuesList []float64
	for x < y {
		val := float64(t + (x/jump)%10)
		valuesList = append(valuesList, val)
		x += jump
	}
	return valuesList
}
