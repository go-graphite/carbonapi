package smartSummarize

import (
	"math"
	"strconv"
	"testing"

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

func TestSummarizeEmptyData(t *testing.T) {
	tests := []th.EvalTestItem{
		{
			"smartSummarize(metric1,'1hour','sum','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo.bar", 0, 1}: {},
			},
			[]*types.MetricData{},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
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
			"smartSummarize(metric1,'1hour','sum','y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 1800+3.5*3600, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','sum','month')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 1800+3.5*3600, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','month')",
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
			"smartSummarize(metric1,'1minute','sum','minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{1770, 5370, 8970, 12570},
			"smartSummarize(metric1,'1minute','sum','minute')",
			60,
			0,
			240,
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
			"smartSummarize(metric1,'1minute','avg','minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{29.5, 89.5, 149.5, 209.5},
			"smartSummarize(metric1,'1minute','avg','minute')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','last','minute')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','last','minute')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'4hours','sum','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{103672800},
			"smartSummarize(metric1,'4hours','sum','weeks4')",
			14400,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1d','sum','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{62164800},
			"smartSummarize(metric1,'1d','sum','days')",
			86400,
			0,
			86400,
		},
		{
			"smartSummarize(metric1,'1minute','sum','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{1770, 5370, 8970, 12570},
			"smartSummarize(metric1,'1minute','sum','seconds')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1hour','max','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','max','hours')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'6m','sum', 'minutes')", // Test having a smaller interval than the data's step
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", []float64{
					2, 4, 6}, 600, 1410345000)},
			},
			[]float64{2, 4, math.NaN(), 6, math.NaN()},
			"smartSummarize(metric1,'6m','sum','minutes')",
			360,
			1410345000,
			1410345000 + 3*600,
		},
		{
			"smartSummarize(metric2,'2minute','sum')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric2", 0, 1}: {types.MakeMetricData("metric2", []float64{1, 2, 3, 4}, 60, 0)},
			},
			[]float64{3, 7},
			"smartSummarize(metric2,'2minute','sum')",
			120,
			0,
			240,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignTo1Year(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1hour','sum','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','1y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','avg','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{1799.5, 5399.5, 8999.5, 12599.5},
			"smartSummarize(metric1,'1hour','avg','1y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','last','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','last','1y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','max','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','max','1y')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','min','1y')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{0, 3600, 7200, 10800},
			"smartSummarize(metric1,'1hour','min','1y')",
			3600,
			0,
			14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToMonths(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1hour','sum','months')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','months')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','avg','months')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{1799.5, 5399.5, 8999.5, 12599.5},
			"smartSummarize(metric1,'1hour','avg','months')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','last','months')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','last','months')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','max','months')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','max','months')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','min','months')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{0, 3600, 7200, 10800},
			"smartSummarize(metric1,'1hour','min','months')",
			3600,
			0,
			14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToWeeksThursday(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'4hours','sum','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{103672800},
			"smartSummarize(metric1,'4hours','sum','weeks4')",
			14400,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'4hours','avg','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{7199.5},
			"smartSummarize(metric1,'4hours','avg','weeks4')",
			14400,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'4hours','last','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{14399},
			"smartSummarize(metric1,'4hours','last','weeks4')",
			14400,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'4hours','max','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{14399},
			"smartSummarize(metric1,'4hours','max','weeks4')",
			14400,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'4hours','min','weeks4')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{0},
			"smartSummarize(metric1,'4hours','min','weeks4')",
			14400,
			0,
			14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToDays(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1day','sum','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{62164800},
			"smartSummarize(metric1,'1day','sum','days')",
			86400,
			0,
			86400,
		},
		{
			"smartSummarize(metric1,'1day','avg','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{43170.0},
			"smartSummarize(metric1,'1day','avg','days')",
			86400,
			0,
			86400,
		},
		{
			"smartSummarize(metric1,'1day','last','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{86340},
			"smartSummarize(metric1,'1day','last','days')",
			86400,
			0,
			86400,
		},
		{
			"smartSummarize(metric1,'1day','max','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{86340},
			"smartSummarize(metric1,'1day','max','days')",
			86400,
			0,
			86400,
		},
		{
			"smartSummarize(metric1,'1day','min','days')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			[]float64{0},
			"smartSummarize(metric1,'1day','min','days')",
			86400,
			0,
			86400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToHours(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1hour','sum','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{6478200, 19438200, 32398200, 45358200},
			"smartSummarize(metric1,'1hour','sum','hours')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','avg','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{1799.5, 5399.5, 8999.5, 12599.5},
			"smartSummarize(metric1,'1hour','avg','hours')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','last','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','last','hours')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','max','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{3599, 7199, 10799, 14399},
			"smartSummarize(metric1,'1hour','max','hours')",
			3600,
			0,
			14400,
		},
		{
			"smartSummarize(metric1,'1hour','min','hours')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			[]float64{0, 3600, 7200, 10800},
			"smartSummarize(metric1,'1hour','min','hours')",
			3600,
			0,
			14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToMinutes(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1minute','sum','minutes')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{1770, 5370, 8970, 12570},
			"smartSummarize(metric1,'1minute','sum','minutes')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','avg','minutes')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{29.5, 89.5, 149.5, 209.5},
			"smartSummarize(metric1,'1minute','avg','minutes')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','last','minutes')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','last','minutes')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','max','minutes')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','max','minutes')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','min','minutes')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{0, 60, 120, 180},
			"smartSummarize(metric1,'1minute','min','minutes')",
			60,
			0,
			240,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToSeconds(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			"smartSummarize(metric1,'1minute','sum','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{1770, 5370, 8970, 12570},
			"smartSummarize(metric1,'1minute','sum','seconds')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','avg','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{29.5, 89.5, 149.5, 209.5},
			"smartSummarize(metric1,'1minute','avg','seconds')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','last','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','last','seconds')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','max','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{59, 119, 179, 239},
			"smartSummarize(metric1,'1minute','max','seconds')",
			60,
			0,
			240,
		},
		{
			"smartSummarize(metric1,'1minute','min','seconds')",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			[]float64{0, 60, 120, 180},
			"smartSummarize(metric1,'1minute','min','seconds')",
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
				"smartSummarize(metric1.foo,'1minute','last')": {types.MakeMetricData("smartSummarize(metric1.foo,'1minute','last')",
					[]float64{59, 119, 179, 239}, 60, 0).SetTag("smartSummarize", "60").SetTag("smartSummarizeFunction", "last")},
				"smartSummarize(metric1.bar,'1minute','last')": {types.MakeMetricData("smartSummarize(metric1.bar,'1minute','last')",
					[]float64{59, 119, 179, 239}, 60, 0).SetTag("smartSummarize", "60").SetTag("smartSummarizeFunction", "last")},
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

func TestSmartSummarizeErrors(t *testing.T) {
	tests := []th.EvalTestItemWithError{
		{
			Target: "smartSummarize(metric1,'-1minute','sum','minute')", // Test to make sure error occurs when a negative interval is used
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 1}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Error: parser.ErrInvalidInterval,
		},
	}

	for n, tt := range tests {
		testName := tt.Target
		t.Run(testName+"#"+strconv.Itoa(n), func(t *testing.T) {
			th.TestEvalExprWithError(t, &tt)
		})
	}
}

func generateValues(start, stop, step int64) (values []float64) {
	for i := start; i < stop; i += step {
		values = append(values, float64(i))
	}
	return
}
