package reduce

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"

	"github.com/go-graphite/carbonapi/expr/functions/alias"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByNode"
	"github.com/go-graphite/carbonapi/expr/functions/aliasSub"
	"github.com/go-graphite/carbonapi/expr/functions/asPercent"
	"github.com/go-graphite/carbonapi/expr/functions/group"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
	for _, m := range alias.New("") {
		metadata.RegisterFunction(m.Name, m.F)
	}
	for _, m := range aliasByNode.New("") {
		metadata.RegisterFunction(m.Name, m.F)
	}
	for _, m := range asPercent.New("") {
		metadata.RegisterFunction(m.Name, m.F)
	}
	for _, m := range aliasSub.New("") {
		metadata.RegisterFunction(m.Name, m.F)
	}
	for _, m := range group.New("") {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestReduce(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			Target: `reduceSeries(group.server*.*, "asPercent", 2, "bytes_used", "total_bytes")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "group.server*.*", From: 0, Until: 1}: {
					types.MakeMetricData("group.server1.bytes_used", []float64{1}, 1, now32),
					types.MakeMetricData("group.server1.total_bytes", []float64{2}, 1, now32),
					types.MakeMetricData("group.server2.bytes_used", []float64{3}, 1, now32),
					types.MakeMetricData("group.server2.total_bytes", []float64{4}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("group.server1.reduce.asPercent", []float64{50}, 1, now32).SetNameTag("group.server1.bytes_used"),
				types.MakeMetricData("group.server2.reduce.asPercent", []float64{75}, 1, now32).SetNameTag("group.server2.bytes_used"),
			},
		},
		{
			// regression test: must group and match on the aliased name, not the original "name" tag
			Target: `reduceSeries(group(aliasSub(aliasByNode(servers.us.dc1.host[0-9]*.cpu.raw_used, 3, 5), 'raw_used', 'cpu.actual'), aliasSub(aliasByNode(servers.us.dc1.host[0-9]*.cpu.raw_total, 3, 5), 'raw_total', 'cpu.max')), "asPercent", 2, "actual", "max")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "servers.us.dc1.host[0-9]*.cpu.raw_used", From: 0, Until: 1}: {
					types.MakeMetricData("servers.us.dc1.host01.cpu.raw_used", []float64{1}, 1, now32),
					types.MakeMetricData("servers.us.dc1.host02.cpu.raw_used", []float64{3}, 1, now32),
				},
				{Metric: "servers.us.dc1.host[0-9]*.cpu.raw_total", From: 0, Until: 1}: {
					types.MakeMetricData("servers.us.dc1.host01.cpu.raw_total", []float64{2}, 1, now32),
					types.MakeMetricData("servers.us.dc1.host02.cpu.raw_total", []float64{4}, 1, now32),
				},
			},
			Want: []*types.MetricData{
				types.MakeMetricData("host01.cpu.reduce.asPercent", []float64{50}, 1, now32).SetNameTag("servers.us.dc1.host01.cpu.raw_used"),
				types.MakeMetricData("host02.cpu.reduce.asPercent", []float64{75}, 1, now32).SetNameTag("servers.us.dc1.host02.cpu.raw_used"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Target, func(t *testing.T) {
			eval := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
			th.TestEvalExpr(t, eval, &tt)
		})
	}
}

func TestReduceErrors(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItemWithError{
		{
			// reduceNode out of range for the series name must not panic
			Target: `reduceSeries(group.*, "asPercent", 4, "bytes_used", "total_bytes")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "group.*", From: 0, Until: 1}: {
					types.MakeMetricData("group.bytes_used", []float64{1}, 1, now32),
					types.MakeMetricData("group.total_bytes", []float64{2}, 1, now32),
				},
			},
			Error: parser.ErrInvalidArg,
		},
		{
			Target: `reduceSeries(group.*, "asPercent", -5, "bytes_used", "total_bytes")`,
			M: map[parser.MetricRequest][]*types.MetricData{
				{Metric: "group.*", From: 0, Until: 1}: {
					types.MakeMetricData("group.bytes_used", []float64{1}, 1, now32),
					types.MakeMetricData("group.total_bytes", []float64{2}, 1, now32),
				},
			},
			Error: parser.ErrInvalidArg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Target, func(t *testing.T) {
			eval := th.EvaluatorFromFuncWithMetadata(metadata.FunctionMD.Functions)
			th.TestEvalExprWithError(t, eval, &tt)
		})
	}
}
