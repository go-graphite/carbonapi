//go:build !cairo
// +build !cairo

package png

import (
	"context"
	"net/http"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const HaveGraphSupport = false

func EvalExprGraph(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return nil, nil
}

// skipcq: CRT-P0003
func MarshalPNG(params PictureParams, results []*types.MetricData) []byte {
	return nil
}

// skipcq: CRT-P0003
func MarshalSVG(params PictureParams, results []*types.MetricData) []byte {
	return nil
}

// skipcq: CRT-P0003
func MarshalPNGRequest(r *http.Request, results []*types.MetricData, templateName string) []byte {
	return nil
}

// skipcq: CRT-P0003
func MarshalSVGRequest(r *http.Request, results []*types.MetricData, templateName string) []byte {
	return nil
}

// skipcq: CRT-P0003
func Description() map[string]types.FunctionDescription {
	return nil
}
