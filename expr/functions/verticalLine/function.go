//go:build !cairo
// +build !cairo

package verticalLine

import (
	"context"
	"fmt"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

var UnsupportedError = fmt.Errorf("must build w/ cairo support")

type verticalLine struct {
	interfaces.FunctionBase
}

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

func (f *verticalLine) Do(_ context.Context, _ parser.Expr, _, _ int64, _ map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return nil, UnsupportedError
}

func (f *verticalLine) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"verticalLine": {
			Description: "Draws a vertical line at the designated timestamp with optional\n  'label' and 'color'. This function is unsupported in this build (built w/o Cairo).",
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
