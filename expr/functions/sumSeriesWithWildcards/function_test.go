package sumSeriesWithWildcards

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
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	helper.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunctionMultiReturn(t *testing.T) {
	now32 := int64(time.Now().Unix())

	tests := []th.MultiReturnEvalTestItem{
		{
			"sumSeriesWithWildcards(metric1.foo.*.*,1,2)",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.foo.*.*", 0, 1}: {
					types.MakeMetricData("metric1.foo.bar1.baz", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric1.foo.bar1.qux", []float64{6, 7, 8, 9, 10}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.baz", []float64{11, 12, 13, 14, 15}, 1, now32),
					types.MakeMetricData("metric1.foo.bar2.qux", []float64{7, 8, 9, 10, 11}, 1, now32),
				},
			},
			"sumSeriesWithWildcards",
			map[string][]*types.MetricData{
				"sumSeriesWithWildcards(metric1.baz)": {types.MakeMetricData("sumSeriesWithWildcards(metric1.baz)", []float64{12, 14, 16, 18, 20}, 1, now32)},
				"sumSeriesWithWildcards(metric1.qux)": {types.MakeMetricData("sumSeriesWithWildcards(metric1.qux)", []float64{13, 15, 17, 19, 21}, 1, now32)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestMultiReturnEvalExpr(t, &tt)
		})
	}

}
