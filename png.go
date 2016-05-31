// +build !cairo

package main

import "net/http"

func marshalPNG(r *http.Request, results []*metricData) []byte {
	return nil
}
