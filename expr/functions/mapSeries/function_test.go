package mapSeries

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

func TestFunction(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			"mapSeries(servers.*.cpu.*, 1)",
			map[parser.MetricRequest][]*types.MetricData{
				{"servers.*.cpu.*", 0, 1}: {
					types.MakeMetricData("servers.server1.cpu.valid", []float64{1, 2, 3}, 1, now32),
					types.MakeMetricData("servers.server2.cpu.valid", []float64{6, 7, 8}, 1, now32),
					types.MakeMetricData("servers.server1.cpu.total", []float64{1, 2, 4}, 1, now32),
					types.MakeMetricData("servers.server2.cpu.total", []float64{5, 7, 8}, 1, now32),
					types.MakeMetricData("servers.server3.cpu.valid", []float64{8, 10, 11}, 1, now32),
					types.MakeMetricData("servers.server3.cpu.total", []float64{9, 10, 11}, 1, now32),
					types.MakeMetricData("servers.server4.cpu.valid", []float64{11, 13, 14}, 1, now32),
					types.MakeMetricData("servers.server4.cpu.total", []float64{12, 13, 14}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("servers.server1.cpu.valid", []float64{1, 2, 3}, 1, now32),
				types.MakeMetricData("servers.server1.cpu.total", []float64{1, 2, 4}, 1, now32),
				types.MakeMetricData("servers.server2.cpu.valid", []float64{6, 7, 8}, 1, now32),
				types.MakeMetricData("servers.server2.cpu.total", []float64{5, 7, 8}, 1, now32),
				types.MakeMetricData("servers.server3.cpu.valid", []float64{8, 10, 11}, 1, now32),
				types.MakeMetricData("servers.server3.cpu.total", []float64{9, 10, 11}, 1, now32),
				types.MakeMetricData("servers.server4.cpu.valid", []float64{11, 13, 14}, 1, now32),
				types.MakeMetricData("servers.server4.cpu.total", []float64{12, 13, 14}, 1, now32),
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}

}
