package powSeries

import (
	"context"
	"math"
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

var (
	now   = time.Now().Unix()
	tests = []th.EvalTestItem{
		{
			"powSeries(collectd.test-db1.load.value, collectd.test-db2.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db1.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db1.load.value", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 312.1}, 1, now)},
				{"collectd.test-db2.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db1.load.value", []float64{1, 3, 5, 7, math.NaN(), 6, 4, 8, 0, 10, 234.2}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db1.load.value, collectd.test-db2.load.value)",
				[]float64{1.0, 8.0, 243.0, 16384.0, math.NaN(), 46656.0, 2401.0, 16777216.0, 1.0, 0.0, math.NaN()}, 1, now).SetNameTag("powSeries(collectd.test-db1.load.value, collectd.test-db2.load.value)")},
		},
		{
			"powSeries(collectd.test-db3.load.value, collectd.test-db4.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db3.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db3.load.value", []float64{1, 2, 666}, 1, now)},
				{"collectd.test-db4.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db4.load.value", []float64{1, 2}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db3.load.value, collectd.test-db4.load.value)",
				[]float64{1.0, 4.0, math.NaN()}, 1, now).SetNameTag("powSeries(collectd.test-db3.load.value, collectd.test-db4.load.value)")},
		},
		{
			"powSeries(collectd.test-db5.load.value, collectd.test-db6.load.value)",
			map[parser.MetricRequest][]*types.MetricData{
				{"collectd.test-db5.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db5.load.value", []float64{1, 2}, 1, now)},
				{"collectd.test-db6.load.value", 0, 1}: {types.MakeMetricData("collectd.test-db6.load.value", []float64{1, 2, 666}, 1, now)},
			},
			[]*types.MetricData{types.MakeMetricData("powSeries(collectd.test-db5.load.value, collectd.test-db6.load.value)",
				[]float64{1.0, 4.0, math.NaN()}, 1, now).SetNameTag("powSeries(collectd.test-db5.load.value, collectd.test-db6.load.value)")},
		},
	}
)

func TestFunction(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Target, func(t *testing.T) {
			th.TestEvalExpr(t, &test)
		})
	}
}

func BenchmarkFunction(b *testing.B) {
	for _, test := range tests {
		for i := 0; i < b.N; i++ {
			b.Run(test.Target, func(b *testing.B) {
				evaluator := metadata.GetEvaluator()

				exp, _, err := parser.ParseExpr(test.Target)
				if err != nil {
					b.Fatalf("could not parse target expression %s", test.Target)
				}

				_, err = evaluator.Eval(context.Background(), exp, 0, 1, test.M)
				if err != nil {
					b.Fatalf("could not evaluate expression %s", test.Target)
				}
			})
		}
	}
}
