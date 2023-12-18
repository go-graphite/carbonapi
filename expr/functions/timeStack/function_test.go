package timeStack

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

func TestTimeStack(t *testing.T) {
	var startTime int64 = 86400

	tests := []th.EvalTestItemWithRange{
		// TODO(civil): Do not pass `true` resetEnd parameter in 0.15
		{
			Target: `timeStack(metric1, "10m", 0)`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: startTime, Until: startTime + 6}: {types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), math.NaN(), 3, math.NaN(), 5, 6, math.NaN(), 7, math.NaN(), math.NaN()}, 60, startTime)},
			},
			Want: []*types.MetricData{types.MakeMetricData("timeShift(metric1,10m,0)",
				[]float64{math.NaN(), math.NaN(), math.NaN(), 3, math.NaN(), 5, 6, math.NaN(), 7, math.NaN(), math.NaN()}, 60, startTime).SetTags(map[string]string{"timeShift": "0", "timeShiftUnit": "10m"}).SetNameTag("metric1")},
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
