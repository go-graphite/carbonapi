// +build !cairo

package expr

import "net/http"

func MarshalPNG(r *http.Request, results []*MetricData) []byte {
	return nil
}

func MarshalSVG(r *http.Request, results []*MetricData) []byte {
	return nil
}
