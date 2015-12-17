// +build !cairo

package main

import "net/http"

const defaultImageRender = "png"

func marshalPNGCairo(r *http.Request, results []*metricData) []byte {
	return marshalPNG(r, results)
}
