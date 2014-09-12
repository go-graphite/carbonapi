package main

import (
	"math"
	"reflect"
	"testing"
	"time"

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
			"3.1",
			&expr{val: 3.1, etype: etConst},
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
		{
			"func1(metric1,'stringconst')",
			&expr{
				target: "func1",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "stringconst", etype: etString},
				},
				argString: "metric1,'stringconst'",
			},
		},
		{
			`func1(metric1,"stringconst")`,
			&expr{
				target: "func1",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "stringconst", etype: etString},
				},
				argString: `metric1,"stringconst"`,
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

func makeResponse(name string, values []float64, step, start int32) *pb.FetchResponse {

	absent := make([]bool, len(values))

	for i, v := range values {
		if math.IsNaN(v) {
			values[i] = 0
			absent[i] = true
		}
	}

	stop := start + int32(len(values))*step

	return &pb.FetchResponse{
		Name:      proto.String(name),
		Values:    values,
		StartTime: proto.Int32(start),
		StepTime:  proto.Int32(step),
		StopTime:  proto.Int32(stop),
		IsAbsent:  absent,
	}
}

func TestEvalExpression(t *testing.T) {

	now32 := int32(time.Now().Unix())

	tests := []struct {
		e    *expr
		m    map[metricRequest][]*pb.FetchResponse
		w    []float64
		name string
	}{
		{
			&expr{target: "metric"},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric", 0, 0}: []*pb.FetchResponse{makeResponse("metric", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4, 5},
			"metric",
		},
		{
			&expr{
				target: "sum",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
					&expr{target: "metric3"}},
				argString: "metric1,metric2,metric3",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
				metricRequest{"metric2", 0, 0}: []*pb.FetchResponse{makeResponse("metric2", []float64{2, 3, math.NaN(), 5, 6}, 1, now32)},
				metricRequest{"metric3", 0, 0}: []*pb.FetchResponse{makeResponse("metric3", []float64{3, 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]float64{6, 9, 8, 15, 11},
			"sumSeries(metric1,metric2,metric3)",
		},
		{
			&expr{
				target: "nonNegativeDerivative",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 10, 14, 20}, 1, now32)},
			},
			[]float64{math.NaN(), 2, 2, 4, 4, 6},
			"nonNegativeDerivative(metric1)",
		},
		{
			&expr{
				target: "nonNegativeDerivative",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 1, 4, math.NaN(), 8}, 1, now32)},
			},
			[]float64{math.NaN(), 2, 2, math.NaN(), 3, math.NaN(), 4},
			"nonNegativeDerivative(metric1)",
		},
		/*
			{
				&expr{
					target: "movingAverage",
					etype:  etFunc,
					args: []*expr{
						&expr{target: "metric1"},
						&expr{val: 4, etype: etConst},
					},
					argString: "metric1,4",
				},
				map[metricRequest][]*pb.FetchResponse{
					metricRequest{"metric1", -4, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 1, 1, 1, 2, 4, 6, 4, 6, 8}, 1, now32)},
				},
				[]float64{1, 2, 3, 4, 4, 5},
				"movingAverage(metric1,4)",
			},
		*/
		{
			&expr{
				target: "scale",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 2.5, etype: etConst},
				},
				argString: "metric1,2.5",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 2, math.NaN(), 4, 5}, 1, now32)},
			},
			[]float64{2.5, 5.0, math.NaN(), 10.0, 12.5},
			"scale(metric1,2.5)",
		},
		{
			&expr{
				target: "scaleToSeconds",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 5, etype: etConst},
				},
				argString: "metric1,5",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{60, 120, math.NaN(), 120, 120}, 60, now32)},
			},
			[]float64{5, 10, math.NaN(), 10, 10},
			"scaleToSeconds(metric1,5)",
		},
		{
			&expr{
				target: "keepLastValue",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 3, etype: etConst},
				},
				argString: "metric1,3",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{math.NaN(), 2, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 4, 5}, 1, now32)},
			},
			[]float64{math.NaN(), 2, 2, 2, 2, math.NaN(), 4, 5},
			"keepLastValue(metric1,3)",
		},

		{
			&expr{
				target: "keepLastValue",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
				argString: "metric1",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{math.NaN(), 2, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 4, 5}, 1, now32)},
			},
			[]float64{math.NaN(), 2, 2, 2, 2, 2, 4, 5},
			"keepLastValue(metric1)",
		},
		{
			&expr{
				target: "alias",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "renamed", etype: etString},
				},
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4, 5},
			"renamed",
		},
		{
			&expr{
				target: "aliasByNode",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1.foo.bar.baz"},
					&expr{val: 1, etype: etConst},
				},
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1.foo.bar.baz", 0, 0}: []*pb.FetchResponse{makeResponse("metric1.foo.bar.baz", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4, 5},
			"foo",
		},
		{
			&expr{
				target: "derivative",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{2, 4, 6, 1, 4, math.NaN(), 8}, 1, now32)},
			},
			[]float64{math.NaN(), 2, 2, -5, 3, math.NaN(), 4},
			"derivative(metric1)",
		},
		{
			&expr{
				target: "avg",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
					&expr{target: "metric3"}},
				argString: "metric1,metric2,metric3",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				metricRequest{"metric2", 0, 0}: []*pb.FetchResponse{makeResponse("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				metricRequest{"metric3", 0, 0}: []*pb.FetchResponse{makeResponse("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]float64{2, math.NaN(), 3, 4, 5, 5.5},
			"averageSeries(metric1,metric2,metric3)",
		},
		{
			&expr{
				target: "maxSeries",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
					&expr{target: "metric3"}},
				argString: "metric1,metric2,metric3",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), 2, 3, 4, 5}, 1, now32)},
				metricRequest{"metric2", 0, 0}: []*pb.FetchResponse{makeResponse("metric2", []float64{2, math.NaN(), 3, math.NaN(), 5, 6}, 1, now32)},
				metricRequest{"metric3", 0, 0}: []*pb.FetchResponse{makeResponse("metric3", []float64{3, math.NaN(), 4, 5, 6, math.NaN()}, 1, now32)},
			},
			[]float64{3, math.NaN(), 4, 5, 6, 6},
			"maxSeries(metric1,metric2,metric3)",
		},
		{
			&expr{
				target: "divideSeries",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
				},
				argString: "metric1,metric2",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				metricRequest{"metric2", 0, 0}: []*pb.FetchResponse{makeResponse("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]float64{0.5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 2},
			"divideSeries(metric1,metric2)",
		},
		{
			&expr{
				target: "diffSeries",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{target: "metric2"},
				},
				argString: "metric1,metric2",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
				metricRequest{"metric2", 0, 0}: []*pb.FetchResponse{makeResponse("metric2", []float64{2, math.NaN(), 3, math.NaN(), 0, 6}, 1, now32)},
			},
			[]float64{-1, math.NaN(), math.NaN(), math.NaN(), 4, 6},
			"diffSeries(metric1,metric2)",
		},
		{
			&expr{
				target: "transformNull",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
				},
				argString: "metric1",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]float64{1, 0, 0, 3, 4, 12},
			"transformNull(metric1)",
		},
		{
			&expr{
				target: "transformNull",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{val: 5, etype: etConst},
				},
				argString: "metric1,5",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, math.NaN(), math.NaN(), 3, 4, 12}, 1, now32)},
			},
			[]float64{1, 5, 5, 3, 4, 12},
			"transformNull(metric1,5)",
		},
	}

	for _, tt := range tests {
		g := evalExpr(tt.e, 0, 0, tt.m)
		if g == nil {
			t.Errorf("failed to eval %v", tt.name)
			continue
		}
		if *g[0].StepTime == 0 {
			t.Errorf("missing step for %+v", g)
		}
		if !nearlyEqual(g[0].Values, g[0].IsAbsent, tt.w) {
			t.Errorf("failed: %s: got %+v, want %+v", *g[0].Name, g[0].Values, tt.w)
		}
		if *g[0].Name != tt.name {
			t.Errorf("bad name for %+v: got %v, want %v", g, *g[0].Name, tt.name)
		}
	}
}

func TestEvalSummarize(t *testing.T) {

	t0, err := time.Parse(time.UnixDate, "Wed Sep 10 10:32:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirtyTwo := int32(t0.Unix())
	t0, err = time.Parse(time.UnixDate, "Wed Sep 10 10:30:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirty := int32(t0.Unix())

	now32 := tenThirty

	tests := []struct {
		e     *expr
		m     map[metricRequest][]*pb.FetchResponse
		w     []float64
		name  string
		step  int32
		start int32
		stop  int32
	}{
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "5s", etype: etString},
				},
				argString: "metric1,'5s'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5}, 1, now32)},
			},
			[]float64{5, 10, 15, 20, 25},
			"summarize(metric1,'5s')",
			5,
			now32,
			now32 + 25*1,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "5s", etype: etString},
					&expr{valStr: "avg", etype: etString},
				},
				argString: "metric1,'5s'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4, 5},
			"summarize(metric1,'5s')",
			5,
			now32,
			now32 + 25*1,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "5s", etype: etString},
					&expr{valStr: "max", etype: etString},
				},
				argString: "metric1,'5s','max'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4.5, 5},
			"summarize(metric1,'5s','max')",
			5,
			now32,
			now32 + 25*1,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "5s", etype: etString},
					&expr{valStr: "min", etype: etString},
				},
				argString: "metric1,'5s','min'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			[]float64{0, 1, 1.5, 2, 5},
			"summarize(metric1,'5s','min')",
			5,
			now32,
			now32 + 25*1,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "5s", etype: etString},
					&expr{valStr: "last", etype: etString},
				},
				argString: "metric1,'5s','last'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 0, 0, 0.5, 1, 2, 1, 1, 1.5, 2, 3, 2, 2, 1.5, 3, 4, 3, 2, 3, 4.5, 5, 5, 5, 5, 5}, 1, now32)},
			},
			[]float64{1, 2, 3, 4.5, 5},
			"summarize(metric1,'5s','last')",
			5,
			now32,
			now32 + 25*1,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "10min", etype: etString},
				},
				argString: "metric1,'10min'",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5}, 60, tenThirtyTwo)},
			},
			[]float64{11, 31, 33},
			"summarize(metric1,'10min')",
			600,
			tenThirty,
			tenThirty + 20*60,
		},
		{
			&expr{
				target: "summarize",
				etype:  etFunc,
				args: []*expr{
					&expr{target: "metric1"},
					&expr{valStr: "10min", etype: etString},
					&expr{valStr: "sum", etype: etString},
					&expr{target: "true", etype: etName},
				},
				argString: "metric1,'10min','sum',true",
			},
			map[metricRequest][]*pb.FetchResponse{
				metricRequest{"metric1", 0, 0}: []*pb.FetchResponse{makeResponse("metric1", []float64{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5}, 60, tenThirtyTwo)},
			},
			[]float64{15, 35, 25},
			"summarize(metric1,'10min','sum',true)",
			600,
			tenThirtyTwo,
			tenThirtyTwo + 25*60,
		},
	}

	for _, tt := range tests {
		g := evalExpr(tt.e, 0, 0, tt.m)
		if g == nil {
			t.Errorf("failed to eval %v", tt.name)
			continue
		}
		if *g[0].StepTime != tt.step {
			t.Errorf("bad step for %s: got %d want %d", *g[0].Name, *g[0].StepTime, tt.step)
		}
		if *g[0].StartTime != tt.start {
			t.Errorf("bad start for %s: got %s want %s", *g[0].Name, time.Unix(int64(*g[0].StartTime), 0).Format(time.StampNano), time.Unix(int64(tt.start), 0).Format(time.StampNano))
		}
		if *g[0].StopTime != tt.stop {
			t.Errorf("bad stop for %s: got %s want %s", *g[0].Name, time.Unix(int64(*g[0].StopTime), 0).Format(time.StampNano), time.Unix(int64(tt.stop), 0).Format(time.StampNano))
		}

		if !nearlyEqual(g[0].Values, g[0].IsAbsent, tt.w) {
			t.Errorf("failed: %s: got %+v, want %+v", *g[0].Name, g[0].Values, tt.w)
		}
		if *g[0].Name != tt.name {
			t.Errorf("bad name for %+v: got %v, want %v", g, *g[0].Name, tt.name)
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
