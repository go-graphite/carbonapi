package powSeries

import (
	"math"
	"testing"
	"time"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
	th "github.com/grafana/carbonapi/tests"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunction(t *testing.T) {
	now := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"powSeries(collectd.test-db1.load.value, collectd.test-db2.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db1.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db1.load.value", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 312.1}, 1, now)},
				{"collectd.test-db2.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db1.load.value", []float64{1, 3, 5, 7, math.NaN(), 6, 4, 8, 0, 10, 234.2}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db1.load.value, collectd.test-db2.load.value)", []float64{1.0, 8.0, 243.0, 16384.0, math.NaN(), 46656.0, 2401.0, 16777216.0, 1.0, 0.0, math.NaN()}, 1, now)},
		},
		{
			"powSeries(collectd.test-db3.load.value, collectd.test-db4.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db3.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db3.load.value", []float64{1, 2, 666}, 1, now)},
				{"collectd.test-db4.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db4.load.value", []float64{1, 2}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db3.load.value, collectd.test-db4.load.value)", []float64{1.0, 4.0, math.NaN()}, 1, now)},
		},
		{
			"powSeries(collectd.test-db5.load.value, collectd.test-db6.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db5.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db5.load.value", []float64{1, 2}, 1, now)},
				{"collectd.test-db6.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db6.load.value", []float64{1, 2, 666}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db5.load.value, collectd.test-db6.load.value)", []float64{1.0, 4.0, math.NaN()}, 1, now)},
		},
	}

	for _, test := range tests {
		t.Run(test.Target, func(t *testing.T) {
			th.TestEvalExpr(t, &test)
		})
	}
}
