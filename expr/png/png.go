// +build !cairo

package png

import (
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"net/http"
)

const HaveGraphSupport = false

func EvalExprGraph(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return nil, nil
}

func MarshalPNG(params PictureParams, results []*types.MetricData) []byte {
	return nil
}

func MarshalSVG(params PictureParams, results []*types.MetricData) []byte {
	return nil
}

func MarshalPNGRequest(r *http.Request, results []*types.MetricData) []byte {
	return nil
}

func MarshalSVGRequest(r *http.Request, results []*types.MetricData) []byte {
	return nil
}
