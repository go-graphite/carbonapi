package unique

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

func TestUnique(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`unique(metric[1234])`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric[1234]", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar.baz", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2.foo.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
					types.MakeMetricData("metric3.foo.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
					types.MakeMetricData("metric1.foo.bar.baz", []float64{4, math.NaN(), 5, 6, 7, math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1.foo.bar.baz", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32),
				types.MakeMetricData("metric2.foo.bar.baz", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32),
				types.MakeMetricData("metric3.foo.bar.baz", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32),
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
