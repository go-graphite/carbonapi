package groupByTags

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/functions/sum"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

func init() {
	s := sum.New("")
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

	tests := []th.MultiReturnEvalTestItem{
		{
			parser.NewExpr("groupByTags",
				"metric1.foo.*",
				parser.ArgValue("sum"),
				parser.ArgValue("dc"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*", 0, 1}: {
					types.MakeMetricData("metric1.foo;cpu=cpu1;dc=dc1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu2;dc=dc1", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu3;dc=dc1", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu4;dc=dc1", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"groupByTags",
			map[string][]*types.MetricData{
				"metric1.foo;dc=dc1": {types.MakeMetricData("metric1.foo;dc=dc1", []float64{25, 29, 33, 37, 41}, 1, now32)},
			},
		},
		{
			parser.NewExpr("groupByTags",
				"metric1.foo.*",
				parser.ArgValue("sum"),
				parser.ArgValue("dc"),
				parser.ArgValue("cpu"),
				parser.ArgValue("rack"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*", 0, 1}: {
					types.MakeMetricData("metric1.foo;cpu=cpu1;dc=dc1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu2;dc=dc1", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu3;dc=dc1", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo;cpu=cpu4;dc=dc1", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"groupByTags",
			map[string][]*types.MetricData{
				"metric1.foo;cpu=cpu1;dc=dc1;rack=": {types.MakeMetricData("metric1.foo;cpu=cpu1;dc=dc1;rack=", []float64{1, 2, 3, 4, 5}, 1, now32)},
				"metric1.foo;cpu=cpu2;dc=dc1;rack=": {types.MakeMetricData("metric1.foo;cpu=cpu2;dc=dc1;rack=", []float64{6, 7, 8, 9, 10}, 1, now32)},
				"metric1.foo;cpu=cpu3;dc=dc1;rack=": {types.MakeMetricData("metric1.foo;cpu=cpu3;dc=dc1;rack=", []float64{11, 12, 13, 14, 15}, 1, now32)},
				"metric1.foo;cpu=cpu4;dc=dc1;rack=": {types.MakeMetricData("metric1.foo;cpu=cpu4;dc=dc1;rack=", []float64{7, 8, 9, 10, 11}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}

}
