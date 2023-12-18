package moving_refetch

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-graphite/carbonapi/expr"
	"github.com/go-graphite/carbonapi/expr/functions/moving"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var M = map[parser.MetricRequest][]*types.MetricData{
	// for refetch
	{"metric*", 10, 25}: {
		types.MakeMetricData("metric1", []float64{math.NaN(), math.NaN(), 2, math.NaN(), 4, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 10).
			SetNameTag(`movingAverage(metric1,10)`).SetPathExpression("metric*"),
	},
	{"test1", 10, 25}: {
		types.MakeMetricData("test1", []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 10).
			SetNameTag(`movingAverage(test1,10)`).SetPathExpression("test1"),
	},
}

func init() {
	md := moving.New("")
	expr.InitWithZipper(nil, th.NewTestZipper(M))
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestMovingRefetch(t *testing.T) {
	tests := []th.EvalTestItemWithRange{
		{
			Target: "movingAverage(metric*,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 20, 25}: {types.MakeMetricData("metric1", th.GenerateValues(10, 25, 1), 1, 20).SetPathExpression("metric*")},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(metric1,10)`,
				[]float64{3, 3, 4, 4, math.NaN()}, 1, 20).SetTag("movingAverage", "10").
				SetNameTag(`movingAverage(metric1,10)`).SetPathExpression("metric*"),
			},
			From:  20,
			Until: 25,
		},
		{
			Target: "movingAverage(test1,10)",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"test1", 20, 25}: {types.MakeMetricData("test1", th.GenerateValues(10, 25, 1), 1, 20).SetPathExpression("test1")},
			},
			Want: []*types.MetricData{types.MakeMetricData(`movingAverage(test1,10)`,
				[]float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}, 1, 20).SetTag("movingAverage", "10").SetNameTag(`movingAverage(test1,10)`)},
			From:  20,
			Until: 25,
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExprWithRange(t, &tt)
		})
	}
}
