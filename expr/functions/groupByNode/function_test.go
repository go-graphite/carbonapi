package groupByNode

import (
	"context"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/functions/aggregate"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	s := aggregate.New("")
	for _, m := range s {
		metadata.RegisterFunction(m.Name, m.F)
	}
	md := New("")
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}

	evaluator := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
}

func TestGroupByNode(t *testing.T) {
	now32 := int64(time.Now().Unix())

	mr := parser.MetricRequest{Metric: "metric1.foo.*.*", From: 0, Until: 1}

	tests := []th.MultiReturnEvalTestItem{
		{
			Target: "groupByNode(metric1.foo.*.*,3,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNode",
			Results: map[string][]*types.MetricData{
				"baz": {types.MakeMetricData("baz", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"qux": {types.MakeMetricData("qux", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
		{
			Target: "groupByNode(metric1.foo.*.*,3,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.bar1.01", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.10", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.01", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.10", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNode_names_with_int",
			Results: map[string][]*types.MetricData{
				"01": {types.MakeMetricData("01", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"10": {types.MakeMetricData("10", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
		{
			Target: "groupByNode(metric1.foo.*.*,3,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.bar1.127_0_0_1:2003", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.127_0_0_1:2004", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.127_0_0_1:2003", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.127_0_0_1:2004", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNode_names_with_colons",
			Results: map[string][]*types.MetricData{
				"127_0_0_1:2003": {types.MakeMetricData("127_0_0_1:2003", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"127_0_0_1:2004": {types.MakeMetricData("127_0_0_1:2004", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
		{
			Target: "groupByNode(metric1.foo.*.*,-2,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNode_with_negative_index",
			Results: map[string][]*types.MetricData{
				"bar1": {types.MakeMetricData("bar1", []float64{7, 9, 11, 13, 15}, 1, now32)},
				"bar2": {types.MakeMetricData("bar2", []float64{18, 20, 22, 24, 26}, 1, now32)},
			},
		},
		{
			Target: "groupByNodes(metric1.foo.*.*,\"sum\",0,1,3)",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNodes",
			Results: map[string][]*types.MetricData{
				"metric1.foo.baz": {types.MakeMetricData("metric1.foo.baz", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"metric1.foo.qux": {types.MakeMetricData("metric1.foo.qux", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
		{
			Target: "groupByNode(metric1.foo.*.*,2,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.Ab1==.lag", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bC2=.lag", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.Ab1==.lag", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bC2=.lag=", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			Name: "groupByNode_names_with_special_symbol_equal",
			Results: map[string][]*types.MetricData{
				"Ab1==": {types.MakeMetricData("Ab1==", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"bC2=":  {types.MakeMetricData("bC2=", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
		{
			Target: "groupByNode(metric1.foo.*.*,3,\"sum\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				mr: {
					types.MakeMetricData("metric1.foo.Ab1==.lag;tag1=value1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.Ab2.lag=;tag1=value1", []float64{1, 0, 3, 4, 5}, 1, now32),
				},
			},
			Name: "groupByNode_tagged_names_with_special_symbol_equal",
			Results: map[string][]*types.MetricData{
				"lag":  {types.MakeMetricData("lag", []float64{1, 2, 3, 4, 5}, 1, now32)},
				"lag=": {types.MakeMetricData("lag=", []float64{1, 0, 3, 4, 5}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}

}

func BenchmarkGroupByNode(b *testing.B) {
	target := "groupByNodes(metric1.foo.bar.*,\"sum\",0,2)"
	metrics := map[parser.MetricRequest][]*types.MetricData{
		{Metric: "metric1.foo.bar.*", From: 0, Until: 1}: {
			types.MakeMetricData("metric1.foo.bar.baz1", []float64{1, 2, 3, 4, 5}, 1, 1),
			types.MakeMetricData("metric1.foo.bar.baz2", []float64{1, 2, 3, 4, 5}, 1, 1),
			types.MakeMetricData("metric1.foo.bar.baz3", []float64{1, 2, 3, 4, 5}, 1, 1),
			types.MakeMetricData("metric1.foo.bar.baz4", []float64{1, 2, 3, 4, 5}, 1, 1),
		},
	}

	evaluator := metadata.GetEvaluator()
	exp, _, err := parser.ParseExpr(target)
	if err != nil {
		b.Fatalf("failed to parse %s: %+v", target, err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g, err := evaluator.Eval(context.Background(), exp, 0, 1, metrics)
		if err != nil {
			b.Fatalf("failed to eval %s: %+v", target, err)
		}
		_ = g
	}
}
