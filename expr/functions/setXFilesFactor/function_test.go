package setXFilesFactor

import (
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

func TestSetXFilesFactor(t *testing.T) {
	now64 := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"setXFilesFactor(metric1,0.6)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now64)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1",
				[]float64{1, 2, 3, 4, 5}, 1, now64).SetXFilesFactor(0.6).SetTag("xFilesFactor", "0.6")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
