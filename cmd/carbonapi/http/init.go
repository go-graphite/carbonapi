package http

import (
	"net/http"

	"github.com/dgryski/httputil"
	"github.com/go-graphite/carbonapi/util/ctx"
)

func InitHandlers(headersToPass, headersToLog []string) *http.ServeMux {
	r := http.DefaultServeMux
	r.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/render", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/metrics/find", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/info", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/lb_check", lbcheckHandler)

	r.HandleFunc("/version", versionHandler)
	r.HandleFunc("/version/", versionHandler)

	r.HandleFunc("/functions", enrichContextWithHeaders(headersToPass, headersToLog, functionsHandler))
	r.HandleFunc("/functions/", enrichContextWithHeaders(headersToPass, headersToLog, functionsHandler))

	r.HandleFunc("/tags", enrichContextWithHeaders(headersToPass, headersToLog, tagHandler))
	r.HandleFunc("/tags/", enrichContextWithHeaders(headersToPass, headersToLog, tagHandler))

	r.HandleFunc("/", enrichContextWithHeaders(headersToPass, headersToLog, usageHandler))
	return r
}
