package main

import (
	"math"
	"reflect"
	"testing"

	"code.google.com/p/goprotobuf/proto"

	"github.com/davecgh/go-spew/spew"
	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
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

func makeResponse(name string, values []float64, step int32) *pb.FetchResponse {

	absent := make([]bool, len(values))

	for i, v := range values {
		if math.IsNaN(v) {
			values[i] = 0
			absent[i] = true
		}
	}

	return &pb.FetchResponse{
		Name:     proto.String(name),
		Values:   values,
		StepTime: proto.Int32(step),
		IsAbsent: absent,
	}
}

func TestEvalExpression(t *testing.T) {

	tests := []struct {
		e *expr
		m map[string][]*pb.FetchResponse
		w []float64
	}{
		{
			&expr{target: "metric"},
			map[string][]*pb.FetchResponse{
				"metric": []*pb.FetchResponse{makeResponse("metric", []float64{1, 2, 3, 4, 5}, 1)},
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
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{1, 2, 3, 4, 5}, 1)},
				"metric2": []*pb.FetchResponse{makeResponse("metric2", []float64{2, 3, math.NaN(), 5, 6}, 1)},
				"metric3": []*pb.FetchResponse{makeResponse("metric3", []float64{3, 4, 5, 6, math.NaN()}, 1)},
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
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 10, 14, 20}, 1)},
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
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 1, 4, math.NaN(), 8}, 1)},
			},
			[]float64{math.NaN(), 2, 2, math.NaN(), 3, math.NaN(), 4},
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
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 4, 6, 8}, 1)},
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
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{1, 2, math.NaN(), 4, 5}, 1)},
			},
			[]float64{2.5, 5.0, math.NaN(), 10.0, 12.5},
		},
		{
			&expr{
				target: "scaleToSeconds",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 5, etype: etConst},
				},
			},
			map[string][]*pb.FetchResponse{
				"metric1": []*pb.FetchResponse{makeResponse("metric1", []float64{60, 120, math.NaN(), 120, 120}, 60)},
			},
			[]float64{5, 10, math.NaN(), 10, 10},
		},
	}

	for _, tt := range tests {
		g := evalExpr(tt.e, tt.m)
		if *g[0].StepTime == 0 {
			t.Errorf("missing step for %+v", g)
		}
		if !nearlyEqual(g[0].Values, g[0].IsAbsent, tt.w) {
			t.Errorf("failed: got %+v, want %+v", g[0].Values, tt.w)
		}
	}
}

const eps = 0.0000000001

func nearlyEqual(a []float64, absent []bool, b []float64) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		// "same"
		if absent[i] && math.IsNaN(b[i]) {
			continue
		}
		if absent[i] || math.IsNaN(b[i]) {
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
