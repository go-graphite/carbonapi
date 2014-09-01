package main

import (
	"math"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParseExpr(t *testing.T) {

	tests := []struct {
		s string
		e *expr
	}{
		{"metric",
			&expr{target: "metric"},
		},
		{
			"metric.foo",
			&expr{target: "metric.foo"},
		},
		{"metric.*.foo",
			&expr{target: "metric.*.foo"},
		},
		{
			"func(metric)",
			&expr{
				target:    "func",
				etype:     etFunc,
				args:      []*expr{&expr{target: "metric"}},
				argString: "metric",
			},
		},
		{
			"func(metric1,metric2,metric3)",
			&expr{
				target: "func",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
					&expr{target: "metric3"}},
				argString: "metric1,metric2,metric3",
			},
		},
		{
			"func1(metric1,func2(metricA,metricB),metric3)",
			&expr{
				target: "func1",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "func2",
						etype:     etFunc,
						args:      []*expr{&expr{target: "metricA"}, &expr{target: "metricB"}},
						argString: "metricA,metricB",
					},
					&expr{target: "metric3"}},
				argString: "metric1,func2(metricA,metricB),metric3",
			},
		},

		{
			"3",
			&expr{val: 3, etype: etConst},
		},
		{
			"func1(metric1,3)",
			&expr{
				target: "func1",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 3, etype: etConst},
				},
				argString: "metric1,3",
			},
		},
	}

	for _, tt := range tests {
		e, _, err := parseExpr(tt.s)
		if err != nil {
			t.Errorf("parse for %+v failed: err=%v", tt.s, err)
			continue
		}
		if !reflect.DeepEqual(e, tt.e) {
			t.Errorf("parse for %+v failed: got %+s want %+v", tt.s, spew.Sdump(e), spew.Sdump(tt.e))
		}
	}
}

func TestEvalExpression(t *testing.T) {

	tests := []struct {
		e *expr
		m map[string][]namedExpr
		w []float64
	}{
		{
			&expr{target: "metric"},
			map[string][]namedExpr{
				"metric": []namedExpr{{"metric", []float64{1, 2, 3, 4, 5}}},
			},
			[]float64{1, 2, 3, 4, 5},
		},
		{
			&expr{
				target: "sum",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
					&expr{target: "metric3"}}},
			map[string][]namedExpr{
				"metric1": []namedExpr{{"metric1", []float64{1, 2, 3, 4, 5}}},
				"metric2": []namedExpr{{"metric2", []float64{2, 3, math.NaN(), 5, 6}}},
				"metric3": []namedExpr{{"metric3", []float64{3, 4, 5, 6, math.NaN()}}},
			},
			[]float64{6, 9, 8, 15, 11},
		},
		{
			&expr{
				target: "nonNegativeDerivative",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
			},
			map[string][]namedExpr{
				"metric1": []namedExpr{{"metric1", []float64{2, 4, 6, 10, 14, 20}}},
			},
			[]float64{math.NaN(), 2, 2, 4, 4, 6},
		},
		{
			&expr{
				target: "nonNegativeDerivative",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
			},
			map[string][]namedExpr{
				"metric1": []namedExpr{{"metric1", []float64{2, 4, 6, 0, 4, 8}}},
			},
			[]float64{math.NaN(), 2, 2, math.NaN(), 4, 4},
		},
		{
			&expr{
				target: "movingAverage",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 4, etype: etConst},
				},
			},
			map[string][]namedExpr{
				"metric1": []namedExpr{{"metric1", []float64{2, 4, 6, 4, 6, 8}}},
			},
			[]float64{2, 3, 4, 4, 5, 6},
		},
		{
			&expr{
				target: "scale",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 2.5, etype: etConst},
				},
			},
			map[string][]namedExpr{
				"metric1": []namedExpr{{"metric1", []float64{1, 2, math.NaN(), 4, 5}}},
			},
			[]float64{2.5, 5.0, math.NaN(), 10.0, 12.5},
		},
	}

	for _, tt := range tests {
		g := evalExpr(tt.e, tt.m)
		if !nearlyEqual(g[0].data, tt.w) {
			t.Errorf("failed: got %+v, want %+v", g, tt.w)
		}
	}
}

const eps = 0.0000000001

func nearlyEqual(a, b []float64) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		// "same"
		if math.IsNaN(v) && math.IsNaN(b[i]) {
			continue
		}
		if math.IsNaN(v) || math.IsNaN(b[i]) {
			// unexpected NaN
			return false
		}
		// "close enough"
		if math.Abs(v-b[i]) > eps {
			return false
		}
	}

	return true
}
