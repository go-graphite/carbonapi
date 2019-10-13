package http

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/dgryski/httputil"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/go-graphite/carbonapi/util/ctx"
)

func InitHandlers(headersToPass, headersToLog []string) *http.ServeMux {
	r := http.NewServeMux()
	r.HandleFunc(config.Config.Prefix+"/render/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc(config.Config.Prefix+"/render", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc(config.Config.Prefix+"/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc(config.Config.Prefix+"/metrics/find", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc(config.Config.Prefix+"/info/", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))
	r.HandleFunc(config.Config.Prefix+"/info", httputil.TrackConnections(httputil.TimeHandler(enrichContextWithHeaders(headersToPass, headersToLog, ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI)), bucketRequestTimes)))

	r.HandleFunc(config.Config.Prefix+"/lb_check", lbcheckHandler)

	r.HandleFunc(config.Config.Prefix+"/version", versionHandler)
	r.HandleFunc(config.Config.Prefix+"/version/", versionHandler)

	r.HandleFunc(config.Config.Prefix+"/functions", enrichContextWithHeaders(headersToPass, headersToLog, functionsHandler))
	r.HandleFunc(config.Config.Prefix+"/functions/", enrichContextWithHeaders(headersToPass, headersToLog, functionsHandler))

	r.HandleFunc(config.Config.Prefix+"/tags", enrichContextWithHeaders(headersToPass, headersToLog, tagHandler))
	r.HandleFunc(config.Config.Prefix+"/tags/", enrichContextWithHeaders(headersToPass, headersToLog, tagHandler))

	r.HandleFunc(config.Config.Prefix+"/_internal/capabilities", enrichContextWithHeaders(headersToPass, headersToLog, capabilityHandler))
	r.HandleFunc(config.Config.Prefix+"/_internal/capabilities/", enrichContextWithHeaders(headersToPass, headersToLog, capabilityHandler))

	r.HandleFunc(config.Config.Prefix+"/", enrichContextWithHeaders(headersToPass, headersToLog, usageHandler))

	if config.Config.Expvar.Enabled {
		if config.Config.Expvar.Listen == "" || config.Config.Expvar.Listen == config.Config.Listen {
			r.HandleFunc(config.Config.Prefix+"/debug/vars", expvar.Handler().ServeHTTP)
			if config.Config.Expvar.PProfEnabled {
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/heap", pprof.Index)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/profile", pprof.Profile)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/symbol", pprof.Symbol)
				r.HandleFunc(config.Config.Prefix+"/debug/pprof/trace", pprof.Trace)
			}
		}
	}
	return r
}
