package expr

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
	"unicode"

	"github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/rewrite"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	"github.com/go-graphite/carbonapi/tests/compare"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

func init() {
	rewrite.New(make(map[string]string))
	functions.New(make(map[string]string))
}

func TestGetBuckets(t *testing.T) {
	tests := []struct {
		start       int64
		stop        int64
		bucketSize  int64
		wantBuckets int64
	}{
		{13, 18, 5, 1},
		{13, 17, 5, 1},
		{13, 19, 5, 2},
	}

	for _, test := range tests {
		buckets := helper.GetBuckets(test.start, test.stop, test.bucketSize)
		if buckets != test.wantBuckets {
			t.Errorf("TestGetBuckets failed!\n%v\ngot buckets %d",
				test,
				buckets,
			)
		}
	}
}

func TestAlignToBucketSize(t *testing.T) {
	tests := []struct {
		inputStart int64
		inputStop  int64
		bucketSize int64
		wantStart  int64
		wantStop   int64
	}{
		{
			13, 18, 5,
			10, 20,
		},
		{
			13, 17, 5,
			10, 20,
		},
		{
			13, 19, 5,
			10, 20,
		},
	}

	for _, test := range tests {
		start, stop := helper.AlignToBucketSize(test.inputStart, test.inputStop, test.bucketSize)
		if start != test.wantStart || stop != test.wantStop {
			t.Errorf("TestAlignToBucketSize failed!\n%v\ngot start %d stop %d",
				test,
				start,
				stop,
			)
		}
	}
}

func TestAlignToInterval(t *testing.T) {
	tests := []struct {
		inputStart int64
		inputStop  int64
		bucketSize int64
		wantStart  int64
	}{
		{
			91111, 92222, 5,
			91111,
		},
		{
			91111, 92222, 60,
			91080,
		},
		{
			91111, 92222, 3600,
			90000,
		},
		{
			91111, 92222, 86400,
			86400,
		},
	}

	for _, test := range tests {
		start := helper.AlignStartToInterval(test.inputStart, test.inputStop, test.bucketSize)
		if start != test.wantStart {
			t.Errorf("TestAlignToInterval failed!\n%v\ngot start %d",
				test,
				start,
			)
		}
	}
}

type evalExprTestCase struct {
	metric        string
	request       string
	metricRequest parser.MetricRequest
	values        []float64
	isAbsent      []bool
	stepTime      int64
	from          int64
	until         int64
}

func TestEvalExpr(t *testing.T) {
	tests := map[string]evalExprTestCase{
		"EvalExp with summarize": {
			metric:  "metric1",
			request: "summarize(metric1,'1min')",
			metricRequest: parser.MetricRequest{
				Metric: "metric1",
				From:   1437127020,
				Until:  1437127140,
			},
			values:   []float64{343, 407, 385},
			isAbsent: []bool{false, false, false},
			stepTime: 60,
			from:     1437127020,
			until:    1437127140,
		},
		"metric name starts with digit": {
			metric:  "1metric",
			request: "1metric",
			metricRequest: parser.MetricRequest{
				Metric: "1metric",
				From:   1437127020,
				Until:  1437127140,
			},
			values:   []float64{343, 407, 385},
			isAbsent: []bool{false, false, false},
			stepTime: 60,
			from:     1437127020,
			until:    1437127140,
		},
		"metric unicode name starts with digit": {
			metric:  "1Метрика",
			request: "1Метрика",
			metricRequest: parser.MetricRequest{
				Metric: "1Метрика",
				From:   1437127020,
				Until:  1437127140,
			},
			values:   []float64{343, 407, 385},
			isAbsent: []bool{false, false, false},
			stepTime: 60,
			from:     1437127020,
			until:    1437127140,
		},
	}

	parser.RangeTables = append(parser.RangeTables, unicode.Cyrillic)
	for name, test := range tests {
		t.Run(fmt.Sprintf("%s: %s", "TestEvalExpr", name), func(t *testing.T) {
			exp, e, err := parser.ParseExpr(test.request)
			if err != nil || e != "" {
				t.Errorf("error='%v', leftovers='%v'", err, e)
			}

			metricMap := make(map[parser.MetricRequest][]*types.MetricData)
			request := parser.MetricRequest{
				Metric: test.metric,
				From:   test.from,
				Until:  test.until,
			}

			data := types.MetricData{
				FetchResponse: pb.FetchResponse{
					Name:              request.Metric,
					StartTime:         request.From,
					StopTime:          request.Until,
					StepTime:          test.stepTime,
					Values:            test.values,
					ConsolidationFunc: "average",
					PathExpression:    request.Metric,
				},
				Tags: map[string]string{"name": request.Metric},
			}

			metricMap[request] = []*types.MetricData{
				&data,
			}

			_, err = EvalExpr(context.Background(), exp, request.From, request.Until, metricMap)
			if err != nil {
				t.Errorf("error='%v'", err)
			}
		})
	}
}

func TestEvalExpression(t *testing.T) {

	now32 := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"metric",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric", 0, 1}: {types.MakeMetricData("metric", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("metric", []float64{1, 2, 3, 4, 5}, 1, now32)},
		},
		{
			"metric*",
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
					types.MakeMetricData("metric2", []float64{2, 3, 4, 5, 6}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32),
				types.MakeMetricData("metric2", []float64{2, 3, 4, 5, 6}, 1, now32),
			},
		},
		{
			"reduceSeries(mapSeries(devops.service.*.filter.received.*.count,2), \"asPercent\", 5,\"valid\",\"total\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"devops.service.*.filter.received.*.count", 0, 1}: {
					types.MakeMetricData("devops.service.server1.filter.received.valid.count", []float64{2, 4, 8}, 1, now32),
					types.MakeMetricData("devops.service.server1.filter.received.total.count", []float64{8, 2, 4}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.valid.count", []float64{3, 9, 12}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.total.count", []float64{12, 9, 3}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("devops.service.server1.filter.received.reduce.asPercent.count", []float64{25, 200, 200}, 1, now32),
				types.MakeMetricData("devops.service.server2.filter.received.reduce.asPercent.count", []float64{25, 100, 400}, 1, now32),
			},
		},
		{
			"reduceSeries(mapSeries(devops.service.*.filter.received.*.count,2), \"asPercent\", 5,\"valid\",\"total\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"devops.service.*.filter.received.*.count", 0, 1}: {
					types.MakeMetricData("devops.service.server1.filter.received.total.count", []float64{8, 2, 4}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.valid.count", []float64{3, 9, 12}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.total.count", []float64{12, 9, 3}, 1, now32),
				},
			},
			[]*types.MetricData{
				types.MakeMetricData("devops.service.server2.filter.received.reduce.asPercent.count", []float64{25, 100, 400}, 1, now32),
			},
		},
		{
			"sumSeries(pow(devops.service.*.filter.received.*.count, 0))",
			map[parser.MetricRequest][]*types.MetricData{
				{"devops.service.*.filter.received.*.count", 0, 1}: {
					types.MakeMetricData("devops.service.server1.filter.received.total.count", []float64{8, 2, 4}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.valid.count", []float64{3, 9, 12}, 1, now32),
					types.MakeMetricData("devops.service.server2.filter.received.total.count", []float64{math.NaN(), math.NaN(), math.NaN()}, 1, now32),
				},
			},
			[]*types.MetricData{types.MakeMetricData("sumSeries(pow(devops.service.*.filter.received.*.count, 0))", []float64{2, 2, 2}, 1, now32)},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			th.TestEvalExpr(t, &tt)
		})
	}
}

func TestRewriteExpr(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []struct {
		name       string
		e          parser.Expr
		m          map[parser.MetricRequest][]*types.MetricData
		rewritten  bool
		newTargets []string
	}{
		{
			"ignore non-applyByNode",
			parser.NewExpr("sumSeries",

				"metric*",
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
			},
			false,
			[]string{},
		},
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"metric*",
				0,
				parser.ArgValue("%.count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"metric1.count"},
		},
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"metric*",
				0,
				parser.ArgValue("%.count"),
				parser.ArgValue("% count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"metric*", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"metric1", 0, 1}: {
					types.MakeMetricData("metric1", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"alias(metric1.count,\"metric1 count\")"},
		},
		{
			"applyByNode",
			parser.NewExpr("applyByNode",

				"foo.metric*",
				1,
				parser.ArgValue("%.count"),
			),
			map[parser.MetricRequest][]*types.MetricData{
				{"foo.metric*", 0, 1}: {
					types.MakeMetricData("foo.metric1", []float64{1, 2, 3}, 1, now32),
					types.MakeMetricData("foo.metric2", []float64{1, 2, 3}, 1, now32),
				},
				{"foo.metric1", 0, 1}: {
					types.MakeMetricData("foo.metric1", []float64{1, 2, 3}, 1, now32),
				},
				{"foo.metric2", 0, 1}: {
					types.MakeMetricData("foo.metric2", []float64{1, 2, 3}, 1, now32),
				},
			},
			true,
			[]string{"foo.metric1.count", "foo.metric2.count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewritten, newTargets, err := RewriteExpr(context.Background(), tt.e, 0, 1, tt.m)

			if err != nil {
				t.Errorf("failed to rewrite %v: %+v", tt.name, err)
				return
			}

			if rewritten != tt.rewritten {
				t.Errorf("failed to rewrite %v: expected rewritten=%v but was %v", tt.name, tt.rewritten, rewritten)
				return
			}

			var targetsMatch = true
			if len(tt.newTargets) != len(newTargets) {
				targetsMatch = false
			} else {
				for i := range tt.newTargets {
					targetsMatch = targetsMatch && tt.newTargets[i] == newTargets[i]
				}
			}

			if !targetsMatch {
				t.Errorf("failed to rewrite %v: expected newTargets=%v but was %v", tt.name, tt.newTargets, newTargets)
				return
			}
		})
	}
}

func TestEvalCustomFromUntil(t *testing.T) {
	tests := []struct {
		target string
		m      map[parser.MetricRequest][]*types.MetricData
		w      []float64
		name   string
		from   int64
		until  int64
	}{
		{
			"timeFunction(\"footime\")",
			map[parser.MetricRequest][]*types.MetricData{},
			[]float64{4200.0, 4260.0, 4320.0},
			"footime",
			4200,
			4350,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalMetrics := th.DeepClone(tt.m)
			exp, _, _ := parser.ParseExpr(tt.target)
			g, err := EvalExpr(context.Background(), exp, tt.from, tt.until, tt.m)
			if err != nil {
				t.Errorf("failed to eval %v: %s", tt.name, err)
				return
			}
			if g[0] == nil {
				t.Errorf("returned no value %v", tt.target)
				return
			}

			th.DeepEqual(t, tt.target, originalMetrics, tt.m, false)

			if g[0].StepTime == 0 {
				t.Errorf("missing step for %+v", g)
			}
			if !compare.NearlyEqual(g[0].Values, tt.w) {
				t.Errorf("failed: %s: got %+v, want %+v", g[0].Name, g[0].Values, tt.w)
			}
			if g[0].Name != tt.name {
				t.Errorf("bad name for %+v: got %v, want %v", g, g[0].Name, tt.name)
			}
		})
	}
}
