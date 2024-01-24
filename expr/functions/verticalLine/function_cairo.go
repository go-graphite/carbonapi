//go:build cairo
// +build cairo

package verticalLine

import (
	"context"
	"fmt"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/tags"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

var TsOutOfRangeError = fmt.Errorf("timestamp out of range")

type verticalLine struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &verticalLine{}
	functions := []string{"verticalLine"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *verticalLine) Do(_ context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, _ map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	start, err := e.GetIntervalArg(0, -1)
	if err != nil {
		return nil, err
	}

	ts := until + int64(start)

	if ts < from {
		return nil, fmt.Errorf("ts %s is before start %s: %w", time.Unix(ts, 0), time.Unix(from, 0), TsOutOfRangeError)
	} else if ts > until {
		return nil, fmt.Errorf("ts %s is after end %s: %w", time.Unix(ts, 0), time.Unix(until, 0), TsOutOfRangeError)
	}

	label, err := e.GetStringArgDefault(1, "")
	if err != nil {
		return nil, err
	}

	color, err := e.GetStringArgDefault(2, "")
	if err != nil {
		return nil, err
	}

	md := &types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      label,
			Values:    []float64{1.0, 1.0},
			StartTime: ts,
			StepTime:  1,
			StopTime:  ts,
		},
		Tags: tags.ExtractTags(label),
		GraphOptions: types.GraphOptions{
			DrawAsInfinite: true,
			Color:          color,
		},
	}

	return []*types.MetricData{md}, nil
}

func (f *verticalLine) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"verticalLine": {
			Description: "Draws a vertical line at the designated timestamp with optional\n  'label' and 'color'. Supported timestamp formats include both\n  relative (e.g. -3h) and absolute (e.g. 16:00_20110501) strings,\n  such as those used with ``from`` and ``until`` parameters. When\n  set, the 'label' will appear in the graph legend.",
			Function:    "verticalLine(ts, label=None, color=None)",
			Group:       "Graph",
			Module:      "graphite.render.functions",
			Name:        "verticalLine",
			Params: []types.FunctionParam{
				{
					Name:     "ts",
					Required: true,
					Type:     types.Date,
				},
				{
					Name:     "label",
					Required: false,
					Type:     types.String,
				},
				{
					Name:     "color",
					Required: false,
					Type:     types.String,
				},
			},
		},
	}
}
