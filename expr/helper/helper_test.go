package helper

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/go-graphite/carbonapi/expr/types"
)

func TestGCD(t *testing.T) {
	tests := []struct {
		arg1     int64
		arg2     int64
		expected int64
	}{
		{
			13,
			17,
			1,
		},
		{
			14,
			21,
			7,
		},
		{
			12,
			16,
			4,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("GDC(%v, %v)=>%v", tt.arg1, tt.arg2, tt.expected), func(t *testing.T) {
			value := GCD(tt.arg1, tt.arg2)
			if value != tt.expected {
				t.Errorf("GCD of %v and %v != %v: %v", tt.arg1, tt.arg2, tt.expected, value)
			}
		})
	}
}

func TestLCM(t *testing.T) {
	tests := []struct {
		args     []int64
		expected int64
	}{
		{
			[]int64{2, 3},
			6,
		},
		{
			[]int64{},
			0,
		},
		{
			[]int64{15},
			15,
		},
		{
			[]int64{10, 15, 20},
			60,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("LMC(%v)=>%v", tt.args, tt.expected), func(t *testing.T) {
			value := LCM(tt.args...)
			if value != tt.expected {
				t.Errorf("LCM of %v != %v: %v", tt.args, tt.expected, value)
			}
		})
	}
}

func TestGetCommonStep(t *testing.T) {
	tests := []struct {
		metrics    []*types.MetricData
		commonStep int64
		changed    bool
	}{
		// Different steps and start/stop time
		{
			[]*types.MetricData{
				types.MakeMetricData("metric1", make([]float64, 15), 5, 5), // 5..80
				types.MakeMetricData("metric2", make([]float64, 30), 2, 4), // 4..64
				types.MakeMetricData("metric2", make([]float64, 25), 3, 6), // 6..81
			},
			30,
			true,
		},
		// Same set of points
		{
			[]*types.MetricData{
				types.MakeMetricData("metric1", make([]float64, 15), 5, 5), // 5..80
				types.MakeMetricData("metric2", make([]float64, 15), 5, 5), // 5..80
				types.MakeMetricData("metric3", make([]float64, 15), 5, 5), // 5..80
			},
			5,
			false,
		},
		// Same step, different lengths
		{
			[]*types.MetricData{
				types.MakeMetricData("metric1", make([]float64, 5), 5, 15), // 15..40
				types.MakeMetricData("metric2", make([]float64, 8), 5, 30), // 30..70
				types.MakeMetricData("metric3", make([]float64, 4), 5, 35), // 35..55
			},
			5,
			false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Set %v", i), func(t *testing.T) {
			com, changed := GetCommonStep(tt.metrics)
			if com != tt.commonStep {
				t.Errorf("Result of GetCommonStep: %v; expected is %v", com, tt.commonStep)
			}
			if changed != tt.changed {
				t.Errorf("GetCommonStep changed: %v; expected is %v", changed, tt.changed)
			}
		})
	}
}

func TestScaleToCommonStep(t *testing.T) {
	NaN := math.NaN()
	tests := []struct {
		name       string
		metrics    []*types.MetricData
		commonStep int64
		expected   []*types.MetricData
	}{
		{
			"Normal metrics",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 4), // 4..13
				types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 2, 4),                 // 4..14
				types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5, 6}, 3, 3),              // 3..21
			},
			0,
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{2, 10, 17, NaN}, 6, 0), // 0..18
				types.MakeMetricData("metric2", []float64{1, 3, 5, NaN}, 6, 0),   // 0..18
				types.MakeMetricData("metric3", []float64{1, 2.5, 4.5, 6}, 6, 0), // 0..24
			},
		},

		// Indx     |  0   |   1  |   2  |   3  |   4  |   5  |   6  |   7  |   8  |   9  |   10  |   11  |   12  |   13  |   14  |   15  |   20  |   21  |   22  |   23  |   24  |   25  |   26  |   27  |   28  |   29  |
		// commonStep  6
		// Start  0 (2 - 2 % 6)
		//
		// ConsolidationFunc = "sum", XFilesFactor = 0.45
		//  metric1 |      |      |      |   N  |   3  |   5  |   7  |   9  |  11  |  13  |   15  |   17  |       |       |       |       |       |       |       |       |       |       |       |       |       |       |
		//  metric1 |  N   |      |      |      |      |      |  72  |      |      |      |       |       |       |       |       |       |       |       |       |       |       |       |       |       |       |       |
		//
		// ConsolidationFunc = "min", XFilesFactor = 0.45
		//  metric2 |      |      |      |      |   1  |      |   2  |      |   3  |      |    4  |       |    5  |       |       |       |       |       |       |       |       |       |       |       |       |       |
		//  metric2 |  N   |      |      |      |      |      |   2  |      |      |      |       |       |    N  |       |       |       |       |       |       |       |       |       |       |       |       |       |
		{
			"xFilesFactor and custom aggregate function",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{NaN, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 3).SetConsolidationFunc("sum").SetXFilesFactor(0.45),
				types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 2, 4).SetConsolidationFunc("min").SetXFilesFactor(0.45),
				types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5, 6}, 3, 3).SetConsolidationFunc("max").SetXFilesFactor(0.51),
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4, 5}, 6, 0),
			},
			0,
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{NaN, 72, NaN, NaN, NaN}, 6, 0), // 0..12
				types.MakeMetricData("metric2", []float64{NaN, 2, NaN, NaN, NaN}, 6, 0),  // 0..18
				types.MakeMetricData("metric3", []float64{NaN, 3, 5, NaN, NaN}, 6, 0),    // 0..24
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4, 5}, 6, 0),          // 0..30, unchanged
			},
		},
		{
			"Custom common step",
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{NaN, 3, 5, 7, 9, 11, 13, 15, 17}, 1, 3), // 3..12
				types.MakeMetricData("metric2", []float64{1, 2, 3, 4, 5}, 2, 4),                   // 4..14
				types.MakeMetricData("metric3", []float64{1, 2, 3, 4, 5, 6}, 3, 3),                // 3..21
				types.MakeMetricData("metric6", []float64{1, 2, 3, 4, 5}, 6, 0),                   // 0..30
			},
			12,
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{10, NaN, NaN}, 12, 0), // 0..12
				types.MakeMetricData("metric2", []float64{2.5, 5, NaN}, 12, 0),  // 0..18
				types.MakeMetricData("metric3", []float64{2, 5, NaN}, 12, 0),    // 0..24
				types.MakeMetricData("metric6", []float64{1.5, 3.5, 5}, 12, 0),  // 0..30, unchanged
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScaleToCommonStep(tt.metrics, tt.commonStep)
			if len(result) != len(tt.expected) {
				t.Errorf("Result has different length %v than expected %v", len(result), len(tt.expected))
			}
			for i, r := range result {
				e := tt.expected[i]
				if len(r.Values) != len(e.Values) {
					t.Fatalf("Values of result[%v] has the different length %+v than expected %+v", i, r.Values, e.Values)
				}
				for v, rv := range r.Values {
					ev := e.Values[v]
					if math.IsNaN(rv) != math.IsNaN(ev) {
						t.Errorf("One of result[%v][%v] is NaN, but not the second: result=%v, expected=%v", i, v, rv, ev)
					} else if !math.IsNaN(rv) && (rv != ev) {
						t.Errorf("result[%v][%v] %v != expected[%v][%v]: %v", i, v, rv, i, v, ev)
					}
				}
				if r.StartTime != e.StartTime {
					t.Errorf("result[%v].StartTime %v != expected[%v].StartTime %v", i, r.StartTime, i, e.StartTime)
				}
				if r.StopTime != e.StopTime {
					t.Errorf("result[%v].StopTime %v != expected[%v].StopTime %v", i, r.StopTime, i, e.StopTime)
				}
				if r.StepTime != e.StepTime {
					t.Errorf("result[%v].StepTime %v != expected[%v].StepTime %v", i, r.StepTime, i, e.StepTime)
				}
			}
		})
	}
}

func TestGetCommonTags(t *testing.T) {
	first := map[string]string{"tag1": "value1", "tag2": "onevalue", "tag3": "value3"}
	second := map[string]string{"tag1": "value1", "tag2": "differentvalue", "tag4": "value4"}

	expected := map[string]string{"tag1": "value1"}
	result := GetCommonTags([]*types.MetricData{{Tags: first}, {Tags: second}})

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
