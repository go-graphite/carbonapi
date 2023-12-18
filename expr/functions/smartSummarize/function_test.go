package smartSummarize

import (
	"math"
	"strconv"
	"testing"

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
			Target: "smartSummarize(metric1,'1hour','sum','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			From:  0,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','sum','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','sum','y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			Name:  "smartSummarize(metric1,'1hour','sum','y')",
			From:  0,
			Until: 14400,
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','sum','month')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			Name:  "smartSummarize(metric1,'1hour','sum','month')",
			From:  0,
			Until: 14400,
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','sum','1month')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			Name:  "smartSummarize(metric1,'1hour','sum','1month')",
			From:  0,
			Until: 14400,
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1minute','sum','minute')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{1770, 5370, 8970, 12570},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','sum','minute')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','sum','1minute')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{1770, 5370, 8970, 12570},
			Name:  "smartSummarize(metric1,'1minute','sum','1minute')",
			From:  0,
			Until: 240,
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','avg','minute')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{29.5, 89.5, 149.5, 209.5},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','avg','minute')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','last','minute')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{59, 119, 179, 239},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','last','minute')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1d','sum','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{62164800},
			From:  0,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1d','sum','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
		{
			Target: "smartSummarize(metric1,'1minute','sum','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{1770, 5370, 8970, 12570},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','sum','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1hour','max','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  0,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','max','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'6m','sum', 'minutes')", // Test having a smaller interval than the data's step
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 1410345000, 1410345000 + 3*600}: {types.MakeMetricData("metric1", []float64{
					2, 4, 6}, 600, 1410345000)},
			},
			Want:  []float64{2, 4, math.NaN(), 6, math.NaN()},
			From:  1410345000,
			Until: 1410345000 + 3*600,
			Name:  "smartSummarize(metric1,'6m','sum','minutes')",
			Step:  360,
			Start: 1410345000,
			Stop:  1410345000 + 3*600,
		},
		{
			Target: "smartSummarize(metric2,'2minute','sum')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric2", 0, 300}: {types.MakeMetricData("metric2", []float64{1, 2, 3, 4}, 60, 0)},
			},
			Want:  []float64{3, 7},
			From:  0,
			Until: 300,
			Name:  "smartSummarize(metric2,'2minute','sum')",
			Step:  120,
			Start: 0,
			Stop:  240,
		},
		{
			// This test case is to check that the stop time of the results will be updated to the final bucket's lower bound
			// if it is smaller than the final bucket's upper bound
			Target: "smartSummarize(metric1,'5s','sum')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 243}: {types.MakeMetricData("metric1", generateValues(0, 243, 3), 3, 0)},
			},
			Want:  []float64{3, 15, 12, 33, 45, 27, 63, 75, 42, 93, 105, 57, 123, 135, 72, 153, 165, 87, 183, 195, 102, 213, 225, 117, 243, 255, 132, 273, 285, 147, 303, 315, 162, 333, 345, 177, 363, 375, 192, 393, 405, 207, 423, 435, 222, 453, 465, 237, 240},
			Name:  "smartSummarize(metric1,'5s','sum')",
			From:  0,
			Until: 243,
			Step:  5,
			Start: 0,
			Stop:  245,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignTo1Year(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1hour','sum','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','sum','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','avg','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{1799.5, 5399.5, 8999.5, 12599.5},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','avg','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','last','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','last','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','max','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','max','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','min','1y')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{0, 3600, 7200, 10800},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','min','1y')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToMonths(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1hour','sum','months')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','sum','months')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','avg','months')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{1799.5, 5399.5, 8999.5, 12599.5},
			Name:  "smartSummarize(metric1,'1hour','avg','months')",
			From:  1800,
			Until: 14400,
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','last','months')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','last','months')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','max','months')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','max','months')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','min','months')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{0, 3600, 7200, 10800},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','min','months')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToWeeksThursday(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'4hours','sum','weeks4')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{103672800},
			From:  174600,
			Until: 14400,
			Name:  "smartSummarize(metric1,'4hours','sum','weeks4')",
			Step:  14400,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'4hours','avg','weeks4')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{7199.5},
			From:  174600,
			Until: 14400,
			Name:  "smartSummarize(metric1,'4hours','avg','weeks4')",
			Step:  14400,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'4hours','last','weeks4')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{14399},
			From:  174600,
			Until: 14400,
			Name:  "smartSummarize(metric1,'4hours','last','weeks4')",
			Step:  14400,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'4hours','max','weeks4')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{14399},
			From:  174600,
			Until: 14400,
			Name:  "smartSummarize(metric1,'4hours','max','weeks4')",
			Step:  14400,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'4hours','min','weeks4')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{0},
			From:  174600,
			Until: 14400,
			Name:  "smartSummarize(metric1,'4hours','min','weeks4')",
			Step:  14400,
			Start: 0,
			Stop:  14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToDays(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1day','sum','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{62164800},
			From:  86399,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1day','sum','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
		{
			Target: "smartSummarize(metric1,'1day','avg','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{43170.0},
			From:  86399,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1day','avg','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
		{
			Target: "smartSummarize(metric1,'1day','last','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{86340},
			From:  86399,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1day','last','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
		{
			Target: "smartSummarize(metric1,'1day','max','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{86340},
			From:  86399,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1day','max','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
		{
			Target: "smartSummarize(metric1,'1day','min','days')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 86400}: {types.MakeMetricData("metric1", generateValues(0, 86400, 60), 60, 0)},
			},
			Want:  []float64{0},
			From:  86399,
			Until: 86400,
			Name:  "smartSummarize(metric1,'1day','min','days')",
			Step:  86400,
			Start: 0,
			Stop:  86400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToHours(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1hour','sum','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{6478200, 19438200, 32398200, 45358200},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','sum','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','avg','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{1799.5, 5399.5, 8999.5, 12599.5},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','avg','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','last','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','last','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','max','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{3599, 7199, 10799, 14399},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','max','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
		{
			Target: "smartSummarize(metric1,'1hour','min','hours')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 14400}: {types.MakeMetricData("metric1", generateValues(0, 14400, 1), 1, 0)},
			},
			Want:  []float64{0, 3600, 7200, 10800},
			From:  1800,
			Until: 14400,
			Name:  "smartSummarize(metric1,'1hour','min','hours')",
			Step:  3600,
			Start: 0,
			Stop:  14400,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToMinutes(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1minute','sum','minutes')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{1770, 5370, 8970, 12570},
			From:  59,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','sum','minutes')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','avg','minutes')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{29.5, 89.5, 149.5, 209.5},
			From:  59,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','avg','minutes')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','last','minutes')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{59, 119, 179, 239},
			From:  59,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','last','minutes')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','max','minutes')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{59, 119, 179, 239},
			From:  59,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','max','minutes')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','min','minutes')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{0, 60, 120, 180},
			From:  59,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','min','minutes')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
	}

	for _, tt := range tests {
		th.TestSummarizeEvalExpr(t, &tt)
	}
}

func TestSmartSummarizeAlignToSeconds(t *testing.T) {
	tests := []th.SummarizeEvalTestItem{
		{
			Target: "smartSummarize(metric1,'1minute','sum','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{1770, 5370, 8970, 12570},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','sum','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','avg','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{29.5, 89.5, 149.5, 209.5},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','avg','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','last','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{59, 119, 179, 239},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','last','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','max','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{59, 119, 179, 239},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','max','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
		},
		{
			Target: "smartSummarize(metric1,'1minute','min','seconds')",
			M: map[parser.MetricRequest][]*types.MetricData{
				{"metric1", 0, 240}: {types.MakeMetricData("metric1", generateValues(0, 240, 1), 1, 0)},
			},
			Want:  []float64{0, 60, 120, 180},
			From:  0,
			Until: 240,
			Name:  "smartSummarize(metric1,'1minute','min','seconds')",
			Step:  60,
			Start: 0,
			Stop:  240,
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
