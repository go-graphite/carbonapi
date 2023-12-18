package aliasByRedis

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var r *miniredis.Miniredis

func init() {
	var err error
	r, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
	runtime.SetFinalizer(r, func(r *miniredis.Miniredis) {
		r.Close()
	})
	r.HSet("alias", "metric1", "new1")
	r.HSet("alias", "metric2", "new2")

	config, err := os.CreateTemp("", "carbonapi-redis-*.yaml")
	if err != nil {
		panic(err)
	}
	defer os.Remove(config.Name())
	config.WriteString(fmt.Sprintf(`
address: %s
enabled: true
`,
		r.Addr(),
	))
	config.Close()

	md := New(config.Name())
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestAliasByRedis(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.EvalTestItem{
		{
			`aliasByRedis(metric*,"alias")`,
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "metric*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("new1", []float64{1, 2, 3, 4, 5}, 1, now32),
				types.MakeMetricData("new2", []float64{1, 2, 3, 4, 5}, 1, now32),
			},
		},
		{
			`aliasByRedis(test.metric2,"alias")`,
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "test.metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData(
						"test.metric2", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("new2", []float64{1, 2, 3, 4, 5}, 1, now32),
			},
		},
		{
			`aliasByRedis(*,"alias")`,
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("test.metric2;tag1=value1", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("new2", []float64{1, 2, 3, 4, 5}, 1, now32).SetTag("tag1", "value1"),
			},
		},
		{
			// non-existing alias
			`aliasByRedis(*,"alias")`,
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "*",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("test.metric3", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("test.metric3;tag1=value1", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("test.metric3", []float64{1, 2, 3, 4, 5}, 1, now32),
				types.MakeMetricData("test.metric3;tag1=value1", []float64{1, 2, 3, 4, 5}, 1, now32),
			},
		},
		{
			// save full path
			`aliasByRedis(test.metric2,"alias", true)`,
			map[parser.MetricRequest][]*types.MetricData{
				{
					Metric: "test.metric2",
					From:   0,
					Until:  1,
				}: {
					types.MakeMetricData("test.metric2", []float64{1, 2, 3, 4, 5}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("test.new2",
				[]float64{1, 2, 3, 4, 5}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}
