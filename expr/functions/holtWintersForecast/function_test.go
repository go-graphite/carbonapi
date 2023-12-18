package holtWintersForecast

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

func TestHoltWintersForecast(t *testing.T) {
	var startTime int64 = 2678400
	var step int64 = 600
	var points int64 = 10
	var seconds int64 = 86400

	tests := []th.EvalTestItemWithRange{
		{
			Target: "holtWintersForecast(metric1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - 7*seconds, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((7*seconds/step)+points)*step, step), step, startTime-7*seconds)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersForecast(metric1)", []float64{4.354532587468384, 5.233762480879125, 5.470443699760628, 5.400062907182546, 4.654782553991797, 4.85560658189784, 3.639077513586465, 4.192121821282148, 4.072238207117917, 4.754208902522321}, step, startTime).SetTag("holtWintersForecast", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersForecast(metric1,'6d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - 6*seconds, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((6*seconds/step)+points)*step, step), step, startTime-6*seconds)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersForecast(metric1)", []float64{3.756495938587323, 4.246729557688366, 4.0724537420914375, 4.707653738003789, 4.526243518254055, 5.324901822037504, 5.491471359733914, 5.360475158485411, 4.56317918291436, 4.719755423132087}, step, startTime).SetTag("holtWintersForecast", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersForecast(metric1,'1d','2d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - seconds, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((seconds/step)+points)*step, step), step, startTime-seconds)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersForecast(metric1)", []float64{4.177645280818122, 4.168426771668243, 4.260421164063269, 4.443824969811369, 4.709783056245225, 5.0502969099660096, 5.458141774396228, 4.923291802762386, 4.540553676160961, 4.2952001684330225}, step, startTime).SetTag("holtWintersForecast", "1"),
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

func generateHwRange(x, y, jump int64) []float64 {
	var valuesList []float64
	for x < y {
		val := float64((x / jump) % 10)
		valuesList = append(valuesList, val)
		x += jump
	}
	return valuesList
}
