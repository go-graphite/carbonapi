package applyByNode

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	evaluator := th.DummyEvaluator()
	metadata.SetEvaluator(evaluator)

	md := New("")
	for _, m := range md {
		metadata.RegisterRewriteFunction(m.Name, m.F)
	}
}

func TestApplyByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.RewriteTestItem{
		{
			`applyByNode(test.metric*.name, 1, "%.transform")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"test.metric*.name", 0, 1}: {
					types.MakeMetricData("test.metric1.name", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
					types.MakeMetricData("test.metric2.name", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			th.RewriteTestResult{
				Rewritten: true,
				Targets: []string{
					"test.metric1.transform",
					"test.metric2.transform",
				},
				Err: nil,
			},
		},
		{
			// overflow
			`applyByNode(metric*.name, 2, "%.transform")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*.name", 0, 1}: {
					types.MakeMetricData("metric1.name", []float64{0, 0, 0, 0, 0, 0}, 1, now32),
				},
			},
			th.RewriteTestResult{
				Err: parser.ErrInvalidArg,
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestRewriteExpr(t, &tt)
		})
	}

}
