package http

import (
	"net/http"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
)

func tagHandler(w http.ResponseWriter, r *http.Request) {
	if config.Config.TagDBProxy != nil {
		config.Config.TagDBProxy.ServeHTTP(w, r)
	} else {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.Write([]byte{'[', ']'})
	}
}
