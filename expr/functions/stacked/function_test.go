package stacked

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestStacked(t *testing.T) {
	input := types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, 0)

	tests := []th.EvalTestItem{
		{
			"stacked(metric1)",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {input},
			},
			[]*types.MetricData{types.MakeMetricData("metric1",
				[]float64{1, 2, 3}, 1, 0).SetTag("stacked", types.DefaultStackName)},
		},
	}

	for _, tt := range tests {
		eval := th.EvaluatorFromFunc(md[0].F)
		th.TestEvalExpr(t, eval, &tt)
	}

	if _, ok := input.Tags["stacked"]; ok {
		t.Error("stacked() must not mutate the tags of its input series")
	}
}
