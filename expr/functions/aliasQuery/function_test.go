package aliasQuery

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

func TestAliasQuery(t *testing.T) {
	now := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			Target: "aliasQuery(channel.power.*, \"channel\\.power\\.([0-9]+)\", \"channel.frequency.\\1\", \"Channel %.f MHz\")",
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "channel.frequency.1", From: 0, Until: 1}: {types.MakeMetricData("channel.frequency.1", []float64{0, 200}, 1, now)},
				{Metric: "channel.frequency.2", From: 0, Until: 1}: {types.MakeMetricData("channel.frequency.2", []float64{400}, 1, now)},
				{Metric: "channel.power.*", From: 0, Until: 1}: {
					types.MakeMetricData("channel.power.1", []float64{1, 2, 3, 4, 5}, 1, now),
					types.MakeMetricData("channel.power.2", []float64{10, 20, 30, 40, 50}, 1, now),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("Channel 200 MHz", []float64{1, 2, 3, 4, 5}, 1, now),
				types.MakeMetricData("Channel 400 MHz", []float64{10, 20, 30, 40, 50}, 1, now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Target, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
