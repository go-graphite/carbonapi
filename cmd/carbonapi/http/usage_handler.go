package http

import (
	"net/http"
)

var usageMsg = []byte(`
supported requests:
    /functions/
    /info/?target=
    /lb_check/
    /metrics/find/?query=
	/render/?target=
	/tags/autoComplete/tags/
    /tags/autoComplete/values/
    /version/
`)

func usageHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(usageMsg)
}
