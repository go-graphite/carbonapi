package aliasByTags

import (
	"math"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/functions/perSecond"
)

func init() {
	md := New("")
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
	psFunc := perSecond.New("")
	for _, m := range psFunc {
		metadata.RegisterFunction(m.Name, m.F)
	}

	evaluator := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
}

func TestAliasByTags(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`aliasByTags(*, "foo")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("metric1.foo.bar.baz;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			`aliasByTags(*, "foo", "name")`,
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("metric1;foo=bar", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			`aliasByTags(*, 2, "blah", "foo", 1)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData(".bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			`aliasByTags(*, 2, "baz", "foo", 1)`,
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("bam.bar.metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			`aliasByTags(perSecond(*), 'name')`,
			map[parser.MetricRequest][]*types.MetricData{
				{"*", 0, 1}: {types.MakeMetricData("base.metric1;foo=bar;baz=bam", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("base.metric1", []float64{math.NaN(), 1, 1, 1, 1}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
