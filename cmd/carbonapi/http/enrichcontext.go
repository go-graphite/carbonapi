package http

import (
	"net/http"

	utilctx "github.com/go-graphite/carbonapi/util/ctx"
)

// TrackConnections exports via expvar a list of all currently executing requests
func enrichContextWithHeaders(headersToPass, headersToLog []string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		headersToPassMap := make(map[string]string)
		for _, name := range headersToPass {
			h := req.Header.Get(name)
			if h != "" {
				headersToPassMap[name] = h
			}
		}

		headersToLogMap := make(map[string]string)
		for _, name := range headersToLog {
			h := req.Header.Get(name)
			if h != "" {
				headersToLogMap[name] = h
			}
		}

		ctx := utilctx.SetPassHeaders(req.Context(), headersToPassMap)
		ctx = utilctx.SetLogHeaders(ctx, headersToLogMap)
		req = req.WithContext(ctx)

		fn(w, req)
	}
}
