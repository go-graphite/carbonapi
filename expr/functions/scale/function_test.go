package scale

import (
	"fmt"
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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"scale(metric1,2.5)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, math.NaN(), 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("scale(metric1,2.5)", []float64{2.5, 5.0, math.NaN(), 10.0, 12.5}, 1, now32).SetTag("scale", "2.5")},
		},
		{
			fmt.Sprintf("scale(x.y.z, -2.5, %d)", int(now32+14)),
			map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{
					Metric: "x.y.z",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"x.y.z",
						[]float64{1, -2, -3, 4, math.NaN(), 0, math.NaN(), 5, 6},
						5,
						now32,
					),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData(
					fmt.Sprintf("scale(x.y.z,-2.5,%d)", now32+14),
					[]float64{1, -2, -3, -10, math.NaN(), 0, math.NaN(), -12.5, -15},
					5,
					now32,
				).SetTag("scale", "-2.5"),
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
