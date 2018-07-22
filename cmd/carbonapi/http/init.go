package http

import (
	"net/http"

	"github.com/dgryski/httputil"
	"github.com/go-graphite/carbonapi/pkg/features"
	"github.com/go-graphite/carbonapi/util/ctx"
)

func InitHandlers() *http.ServeMux {
	r := http.DefaultServeMux
	r.HandleFunc("/render/", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))
	r.HandleFunc("/render", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(renderHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))

	r.HandleFunc("/metrics/find/", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))
	r.HandleFunc("/metrics/find", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(findHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))

	r.HandleFunc("/info/", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))
	r.HandleFunc("/info", httputil.TrackConnections(httputil.TimeHandler(ctx.ParseCtx(infoHandler, ctx.HeaderUUIDAPI), bucketRequestTimes)))

	r.HandleFunc("/lb_check", lbcheckHandler)

	r.HandleFunc("/version", versionHandler)
	r.HandleFunc("/version/", versionHandler)

	r.HandleFunc("/functions", functionsHandler)
	r.HandleFunc("/functions/", functionsHandler)

	r.HandleFunc("/tags", tagHandler)
	r.HandleFunc("/tags/", tagHandler)

	features := features.GetFeaturesInstance()
	r.HandleFunc("/_internal/flags", features.FlagListHandler)
	r.HandleFunc("/_internal/flags/id", features.FlagPatchByIDHandler)
	r.HandleFunc("/_internal/flags/name", features.FlagPatchByNameHandler)

	r.HandleFunc("/", usageHandler)
	return r
}
