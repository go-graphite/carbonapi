// +build !cairo

package expr

import (
	"net/http"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const haveGraphSupport = false

type graphOptions struct {
}

func evalExprGraph(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*MetricData) ([]*MetricData, error) {
	return nil, nil
}

func MarshalPNG(params PictureParams, results []*MetricData) []byte {
	return nil
}

func MarshalSVG(params PictureParams, results []*MetricData) []byte {
	return nil
}

func MarshalPNGRequest(r *http.Request, results []*MetricData) []byte {
	return nil
}

func MarshalSVGRequest(r *http.Request, results []*MetricData) []byte {
	return nil
}
