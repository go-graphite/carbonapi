package color

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

func TestColor(t *testing.T) {

	tests := []th.EvalTestItem{
		{
			"color(metric1,\"green\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1",
					[]float64{1, 2, 3}, 1, 0)},
			},
			[]*types.MetricData{types.MakeMetricData("metric1",
				[]float64{1, 2, 3}, 1, 0)},
		},
	}

	for _, tt := range tests {
		eval := th.EvaluatorFromFunc(md[0].F)
		th.TestEvalExpr(t, eval, &tt)
	}
}
