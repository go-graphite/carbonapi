package polyfit

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
			"polyfit(metric1,3)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("polyfit(metric1,3)",
					[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, now32),
			},
		},
		{
			"polyfit(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{7.79, 7.7, 7.92, 5.25, 6.24, 7.25, 7.15, 8.56, 7.82, 8.52}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("polyfit(metric1)",
					[]float64{6.94763636364, 7.05260606061, 7.15757575758, 7.26254545455, 7.36751515152,
						7.47248484848, 7.57745454545, 7.68242424242, 7.78739393939, 7.89236363636}, 1, now32),
			},
		},
		{
			"polyfit(metric1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{7.79, 7.7, 7.92, 5.25, 6.24, math.NaN(), 7.15, 8.56, 7.82, 8.52}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("polyfit(metric1,2)",
					[]float64{7.9733096590909085, 7.364842329545457, 6.933910511363642, 6.680514204545464, 6.604653409090922,
						6.706328125000017, 6.985538352272748, 7.442284090909116, 8.07656534090912, 8.888382102272761}, 1, now32),
			},
		},
		{
			"polyfit(metric1,3,'5sec')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1",
						[]float64{7.79, 7.7, 7.92, 5.25, 6.24, 7.25, 7.15, 8.56, 7.82, 8.52}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("polyfit(metric1,3,'5sec')",
					[]float64{8.22944055944, 7.26958041958, 6.73364801865, 6.54653846154, 6.63314685315,
						6.91836829837, 7.3270979021, 7.78423076923, 8.21466200466, 8.54328671329,
						8.695, 8.5946969697, 8.16727272727, 7.33762237762, 6.03064102564}, 1, now32),
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
