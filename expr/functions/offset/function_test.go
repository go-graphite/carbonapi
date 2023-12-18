package offset

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"offset(metric1,10)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{93, 94, 95, math.NaN(), 97, 98, 99, 100, 101}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("offset(metric1,10)",
				[]float64{103, 104, 105, math.NaN(), 107, 108, 109, 110, 111}, 1, now32).SetTag("offset", "10")},
		},
		{
			"offset(metric1,'10')", // Verify string input can be parsed into int or float
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{93, 94, 95, math.NaN(), 97, 98, 99, 100, 101}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("offset(metric1,10)",
				[]float64{103, 104, 105, math.NaN(), 107, 108, 109, 110, 111}, 1, now32).SetTag("offset", "10")},
		},
		{
			"add(metric*,-10)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{93, 94, 95, math.NaN(), 97, 98, 99, 100, 101}, 1, now32),
					types.MakeMetricData("metric2", []float64{193, 194, 195, math.NaN(), 197, 198, 199, 200, 201}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("add(metric1,-10)", []float64{83, 84, 85, math.NaN(), 87, 88, 89, 90, 91}, 1, now32).SetTag("add", "-10"),
				types.MakeMetricData("add(metric2,-10)", []float64{183, 184, 185, math.NaN(), 187, 188, 189, 190, 191}, 1, now32).SetTag("add", "-10"),
			},
		},
		{
			"add(metric*,'-10')", // Verify string input can be parsed into int or float
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{93, 94, 95, math.NaN(), 97, 98, 99, 100, 101}, 1, now32),
					types.MakeMetricData("metric2", []float64{193, 194, 195, math.NaN(), 197, 198, 199, 200, 201}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("add(metric1,-10)", []float64{83, 84, 85, math.NaN(), 87, 88, 89, 90, 91}, 1, now32).SetTag("add", "-10"),
				types.MakeMetricData("add(metric2,-10)", []float64{183, 184, 185, math.NaN(), 187, 188, 189, 190, 191}, 1, now32).SetTag("add", "-10"),
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
