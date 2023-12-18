package randomWalk

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"github.com/stretchr/testify/assert"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestRandomWalk(t *testing.T) {
	tests := []th.EvalTestItemWithCustomValidation{
		{
			Target: "randomWalk('foo')",
			M:      map[parser.MetricRequest][]*types.MetricData{},
			From:   0,
			Until:  120,
			Validator: func(t *testing.T, md []*types.MetricData) {
				assert.Equal(t, 1, len(md))
				m := md[0]
				assert.Equal(t, "foo", m.Name)
				assert.Equal(t, int64(60), m.StepTime)
				assert.Equal(t, 2, len(m.Values))
			},
		},
		{
			Target: "randomWalk('foo', step=3)",
			M:      map[parser.MetricRequest][]*types.MetricData{},
			From:   0,
			Until:  120,
			Validator: func(t *testing.T, md []*types.MetricData) {
				assert.Equal(t, 1, len(md))
				m := md[0]
				assert.Equal(t, "foo", m.Name)
				assert.Equal(t, int64(3), m.StepTime)
				assert.Equal(t, 40, len(m.Values))
			},
		},
		{
			Target: "randomWalk('foo', 4)",
			M:      map[parser.MetricRequest][]*types.MetricData{},
			From:   0,
			Until:  120,
			Validator: func(t *testing.T, md []*types.MetricData) {
				assert.Equal(t, 1, len(md))
				m := md[0]
				assert.Equal(t, "foo", m.Name)
				assert.Equal(t, int64(4), m.StepTime)
				assert.Equal(t, 30, len(m.Values))
			},
		},
		{
			Target: "randomWalk('foo', 5)",
			M:      map[parser.MetricRequest][]*types.MetricData{},
			From:   0,
			Until:  121, // Should be rounded to 120
			Validator: func(t *testing.T, md []*types.MetricData) {
				assert.Equal(t, 1, len(md))
				m := md[0]
				assert.Equal(t, "foo", m.Name)
				assert.Equal(t, int64(5), m.StepTime)
				assert.Equal(t, 24, len(m.Values))
				assert.Equal(t, int64(120), m.StopTime)
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExprWithCustomValidation(t, &tt)
		})
	}
}
