package http

import (
	"net/http"

	"github.com/dgryski/httputil"
	"github.com/go-graphite/carbonapi/util/ctx"
)

func InitHandlers(headers []string) *http.ServeMux {
	r := http.DefaultServeMux
	r.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/render", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/metrics/find", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc("/info", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headers, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc("/lb_check", lbcheckHandler)

	r.HandleFunc("/version", versionHandler)
	r.HandleFunc("/version/", versionHandler)

	r.HandleFunc("/functions", enrichContextWithHeaders(headers, functionsHandler))
	r.HandleFunc("/functions/", enrichContextWithHeaders(headers, functionsHandler))

	r.HandleFunc("/tags", enrichContextWithHeaders(headers, tagHandler))
	r.HandleFunc("/tags/", enrichContextWithHeaders(headers, tagHandler))

	r.HandleFunc("/", enrichContextWithHeaders(headers, usageHandler))
	return r
}
