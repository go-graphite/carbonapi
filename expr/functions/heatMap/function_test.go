package heatMap

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

func TestHeatMap(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"heatMap(a.*)",
			map[parser.MetricRequest][]*types.MetricData{
				{"a.*", 0, 1}: {
					types.MakeMetricData("a.a1", []float64{1, 2, 3, 4, 5, 6}, 1, now32),
					types.MakeMetricData("a.a2", []float64{2, math.NaN(), 20, 8, 10, 7}, 1, now32),
					types.MakeMetricData("a.a3", []float64{10, math.NaN(), 3, 17, 10, 90}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("heatMap(a.a2,a.a1)", []float64{1.0, math.NaN(), 17.0, 4.0, 5.0, 1.0}, 1, now32),
				types.MakeMetricData("heatMap(a.a3,a.a2)", []float64{8.0, math.NaN(), -17.0, 9.0, 0.0, 83.0}, 1, now32),
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
