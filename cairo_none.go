// +build !cairo

package main

import "net/http"

func marshalPNGCairo(r *http.Request, results []*metricData) []byte {
	return marshalPNG(r, results)
}
