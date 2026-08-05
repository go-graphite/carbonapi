//go:build !cairo
// +build !cairo

package png

import (
	"net/http"

	"github.com/go-graphite/carbonapi/expr/types"
)

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
