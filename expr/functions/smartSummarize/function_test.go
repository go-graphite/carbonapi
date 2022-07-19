package smartSummarize

import (
	"testing"

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

func TestEvalSummarize(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1hour','sum','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 1800+3.5*3600, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','1y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','sum','1month')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 1800+3.5*3600, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','1month')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1minute','sum','1minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{1770, 5370, 8970, 12570},
			"smartSummarize(metric1,'1minute','sum','1minute')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','avg','1minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{29.5, 89.5, 149.5, 209.5},
			"smartSummarize(metric1,'1minute','avg','1minute')",
			60,
			0,
			240,
		},

		{
			"smartSummarize(metric1,'1minute','last','1minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','last','1minute')",
			60,
			0,
			240,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestFunctionUseNameWithWildcards(t *testing.T) {
	tests := []th.MultiReturnEvalTestItem{
		{
			"smartSummarize(metric1.*,'1minute','last')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1.*", 0, 1}: {
					types.MakeMetricData("metric1.foo", generateValues(0, 240, 1), 1, 0),
					types.MakeMetricData("metric1.bar", generateValues(0, 240, 1), 1, 0),
				},
			},
			"smartSummarize",
			map[string][]*types.MetricData{
				"smartSummarize(metric1.foo,'1minute','last')": {types.MakeMetricData("smartSummarize(metric1.foo,'1minute','last')", []float64{59, 119, 179, 239}, 60, 0)},
				"smartSummarize(metric1.bar,'1minute','last')": {types.MakeMetricData("smartSummarize(metric1.bar,'1minute','last')", []float64{59, 119, 179, 239}, 60, 0)},
			},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			err := th.TestMultiReturnEvalExprModifiedOrigin(t, &tt)
			if err != nil {
				t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
				return
			}
		})
	}
}

func generateValues(start, stop, step int64) (values []float64) {
	for i := start; i < stop; i += step {
		values = append(values, float64(i))
	}
	return
}
