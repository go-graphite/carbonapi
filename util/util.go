package util

import (
	"context"
	"net/http"
)

const (
	ctx_header_prefix = "X-CTX-CarbonAPI-"
)

func ifaceToString(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

func GetCtxString(ctx context.Context, key string) string {
	return ifaceToString(ctx.Value(key))
}

func ParseCtx(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		uuid := req.Header.Get(ctx_header_prefix + "UUID")
		handler := req.Header.Get(ctx_header_prefix + "Handler")
		request_uri := req.Header.Get(ctx_header_prefix + "RequestURL")
		referer := req.Header.Get(ctx_header_prefix + "Referer")
		username := req.Header.Get(ctx_header_prefix + "Username")
		ctx := req.Context()
		ctx = context.WithValue(ctx, "carbonapi_uuid", uuid)
		ctx = context.WithValue(ctx, "carbonapi_handler", handler)
		ctx = context.WithValue(ctx, "carbonapi_request_uri", request_uri)
		ctx = context.WithValue(ctx, "carbonapi_referer", referer)
		ctx = context.WithValue(ctx, "carbonapi_username", username)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func MarshalCtx(ctx context.Context, response *http.Request) *http.Request {
	response.Header.Add(ctx_header_prefix+"UUID", ifaceToString(ctx.Value("carbonapi_uuid")))
	response.Header.Add(ctx_header_prefix+"Handler", ifaceToString(ctx.Value("carbonapi_handler")))
	response.Header.Add(ctx_header_prefix+"RequestURL", ifaceToString(ctx.Value("carbonapi_request_url")))
	response.Header.Add(ctx_header_prefix+"Referer", ifaceToString(ctx.Value("carbonapi_referer")))
	response.Header.Add(ctx_header_prefix+"Username", ifaceToString(ctx.Value("carbonapi_username")))

	return response
}
