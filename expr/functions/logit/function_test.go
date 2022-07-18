package logit

import (
	"math"
	"testing"
	"time"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
	th "github.com/grafana/carbonapi/tests"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"logit(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{0.5, math.NaN(), 0.8, 0.25, 0.1, 1}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("logit(metric1)",
				[]float64{0, math.NaN(), 1.3862943611198908, -1.0986122886681098, -2.197224577336219, math.NaN()}, 1, now32)},
		},
	}

	for _, tt := range tests {
		tt := tt
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			err := th.TestEvalExprModifiedOrigin(t, &tt, 0, 1, false)
			if err != nil {
				t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
				return
			}
		})
	}
}
