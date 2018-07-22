package applyByNode

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	md := New("")
	evaluator := &th.FuncEvaluator{}
	rewriter := th.RewriterFromFunc(md[0].F)
	metadata.SetRewriter(rewriter)
	metadata.SetEvaluator(evaluator)
	helper.SetRewriter(rewriter)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterRewriteFunction(m.Name, m.F)
	}
}

func TestRewriteExpr(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []struct {
		name       string
		e          parser.Expr
		m          map[parser.MetricRequest][]*types.MetricData
		rewritten  bool
		newTargets []string
	}{
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"metric*",
				1,
				parser.ArgValue("%.count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"metric1.count"},
		},
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"metric*",
				1,
				parser.ArgValue("%.count"),
				parser.ArgValue("% count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"alias(metric1.count,\"metric1 count\")"},
		},
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"foo.metric*",
				2,
				parser.ArgValue("%.count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"foo.metric*", 0, 1}: {
					types.MakeMetricData("foo.metric1", []float64{1, 2, 3}, 1, now32),
					types.MakeMetricData("foo.metric2", []float64{1, 2, 3}, 1, now32),
				},
				{"foo.metric1", 0, 1}: {
					types.MakeMetricData("foo.metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"foo.metric2", 0, 1}: {
					types.MakeMetricData("foo.metric2", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"foo.metric1.count", "foo.metric2.count"},
		},
	}

	rewriter := metadata.GetRewriter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewritten, newTargets, err := rewriter.RewriteExpr(tt.e, 0, 1, tt.m)

			if err != nil {
				t.Errorf("failed to rewrite %v: %+v", tt.name, err)
				return
			}

			if rewritten != tt.rewritten {
				t.Errorf("failed to rewrite %v: expected rewritten=%v but was %v", tt.name, tt.rewritten, rewritten)
				return
			}

			var targetsMatch = true
			if len(tt.newTargets) != len(newTargets) {
				targetsMatch = false
			} else {
				for i := range tt.newTargets {
					targetsMatch = targetsMatch && tt.newTargets[i] == newTargets[i]
				}
			}

			if !targetsMatch {
				t.Errorf("failed to rewrite %v: expected newTargets=%v but was %v", tt.name, tt.newTargets, newTargets)
				return
			}
		})
	}
}
