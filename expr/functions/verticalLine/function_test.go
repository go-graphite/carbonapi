//go:build cairo
// +build cairo

package verticalLine

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/tags"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/stretchr/testify/assert"
)

func init() {
	md := New("")
	evaluator := th.EvaluatorFromFunc(md[0].F)
	metadata.SetEvaluator(evaluator)
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestFunction(t *testing.T) {
	now := time.Now()
	nowUnix := now.Unix()

	duration, err := time.ParseDuration("-30m")
	if err != nil {
		assert.NoError(t, err)
	}
	from := now.Add(duration).Unix()

	wantedDuration, err := time.ParseDuration("-5m")
	if err != nil {
		assert.NoError(t, err)
	}
	wantedTs := now.Add(wantedDuration).Unix()

	tests := []th.EvalTestItemWithRange{
		{
			"verticalLine(\"-5m\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo", from, nowUnix}: {types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1, nowUnix)},
			},
			[]*types.MetricData{makeMetricData("", []float64{1.0, 1.0}, 1, wantedTs, wantedTs)},
			from,
			nowUnix,
		},
		{
			"verticalLine(\"-5m\", \"label\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo", from, nowUnix}: {types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1, nowUnix)},
			},
			[]*types.MetricData{makeMetricData("label", []float64{1.0, 1.0}, 1, wantedTs, wantedTs)},
			from,
			nowUnix,
		},
	}

	for _, test := range tests {
		t.Run(test.Target, func(t *testing.T) {
			th.TestEvalExprWithRange(t, &test)
		})
	}
}

func TestFunctionErrors(t *testing.T) {
	now := time.Now()
	nowUnix := now.Unix()

	duration, err := time.ParseDuration("-30m")
	if err != nil {
		assert.NoError(t, err)
	}
	from := now.Add(duration).Unix()

	tests := []th.EvalTestItemWithError{
		{
			"verticalLine(\"-50m\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo", from, nowUnix}: {types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1, nowUnix)},
			},
			[]*types.MetricData{},
			TsOutOfRangeError,
		},
		{
			"verticalLine(\"+5m\")",
			map[parser.MetricRequest][]*types.MetricData{
				{"foo", from, nowUnix}: {types.MakeMetricData("foo", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1, nowUnix)},
			},
			[]*types.MetricData{},
			TsOutOfRangeError,
		},
	}

	for _, test := range tests {
		t.Run(test.Target, func(t *testing.T) {
			th.TestEvalExprWithError(t, &test)
		})
	}
}

func makeMetricData(name string, values []float64, step, start, stop int64) *types.MetricData {
	return &types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      name,
			Values:    values,
			StartTime: start,
			StepTime:  step,
			StopTime:  stop,
		},
		Tags: tags.ExtractTags(name),
	}
}
