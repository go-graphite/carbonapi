//go:build cairo
// +build cairo

package holtWintersConfidenceArea

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

func TestHoltWintersConfidenceArea(t *testing.T) {
	var startTime int64 = 2678400
	var step int64 = 600
	var points int64 = 10

	tests := []th.EvalTestItemWithRange{
		{
			Target: "holtWintersConfidenceArea(metric1)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - holtwinters.DefaultBootstrapInterval, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((holtwinters.DefaultBootstrapInterval/step)+points)*step, step), step, startTime-holtwinters.DefaultBootstrapInterval)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{0.2841206166091448, 1.0581027098774411, 0.3338172102994683, 0.5116859493263242, -0.18199175514936972, 0.2366173792019426, -1.2941554508809152, -0.513426806531049, -0.7970905542723132, 0.09868900726536012}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{8.424944558327624, 9.409422251880809, 10.607070189221787, 10.288439865038768, 9.491556863132963, 9.474595784593738, 8.572310478053845, 8.897670449095346, 8.941566968508148, 9.409728797779282}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersConfidenceArea(metric1,4,'6d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - 6*holtwinters.SecondsPerDay, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((6*holtwinters.SecondsPerDay/step)+points)*step, step), step, startTime-6*holtwinters.SecondsPerDay)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{-0.6535362793382391, -0.26554972418633316, -1.1060549683277792, -0.5788026852576289, -1.4594446935142829, -0.6933311085203409, -1.6566119269969288, -1.251651025511391, -1.7938581852717226, -1.1791817117029604}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{8.166528156512886, 8.759008839563066, 9.250962452510654, 9.994110161265208, 10.511931730022393, 11.34313475259535, 12.639554646464758, 11.972601342482212, 10.920216551100442, 10.618692557967133}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
			},
			From:  startTime,
			Until: startTime + step*points,
		},
		{
			Target: "holtWintersConfidenceArea(metric1,4,'1d','2d')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", startTime - holtwinters.SecondsPerDay, startTime + step*points}: {types.MakeMetricData("metric1", generateHwRange(0, ((holtwinters.SecondsPerDay/step)+points)*step, step), step, startTime-holtwinters.SecondsPerDay)},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{4.106587168490873, 3.8357974803355406, 3.564589629688576, 3.421354957735917, 3.393696278743315, 3.470415673952413, 3.2748850646377368, 3.3539750816574316, 3.5243322056965765, 3.7771201010598134}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
				types.MakeMetricData("holtWintersConfidenceArea(metric1)", []float64{4.24870339314537, 4.501056063000946, 4.956252698437961, 5.466294981886822, 6.0258698337471355, 6.630178145979606, 7.6413984841547204, 6.492608523867341, 5.556775146625346, 4.813280235806231}, step, startTime).SetTag("holtWintersConfidenceArea", "1"),
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
