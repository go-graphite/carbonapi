package transformNull

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

func TestTransformNull(t *testing.T) {
	now := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			`transformNull(metric1)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1)",
				[]float64{1, 0, 0, 3, 4, 12}, 1, now).SetTag("transformNull", "0").SetNameTag("metric1")},
		},
		{
			`transformNull(metric1, default=5)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1,5)",
				[]float64{1, 5, 5, 3, 4, 12}, 1, now).SetTag("transformNull", "5").SetNameTag("metric1")},
		},
		{
			`transformNull(metric1, default=5, referenceSeries=metric2.*)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, math.NaN(), math.NaN(), math.NaN(), 4, 12}, 1, now)},
				{"metric2.*", 0, 1}: {
					types.MakeMetricData("metric2.foo", []float64{math.NaN(), 3, math.NaN(), 3, math.NaN(), 12}, 1, now),
					types.MakeMetricData("metric2.bar", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1,5)",
				[]float64{1, 5, math.NaN(), 5, 4, 12}, 1, now).SetTag("transformNull", "5").SetNameTag("metric1")},
		},
		{
			`transformNull(metric1, default=5, defaultOnAbsent=True)`,
			map[parser.MetricRequest][]*types.MetricData{},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1, default=5, defaultOnAbsent=True)",
				[]float64{5, 5}, 1, 0).SetTag("transformNull", "5").SetNameTag(`metric1`)},
		},
		// tagged metric
		{
			`transformNull(seriesByTag('name=metric1'), 0)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"seriesByTag('name=metric1')", 0, 1}: {types.MakeMetricData("metric1;env=prod", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("transformNull(metric1;env=prod,0)",
				[]float64{1, 0, 0, 3, 4, 12}, 1, now).SetTag("transformNull", "0").SetNameTag("metric1").SetTag("env", "prod")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
