package util

import (
	"net/http"
	"context"
)

const (
	ctx_header_prefix = "X-CTX-CarbonZipper-"
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
		ctx := req.Context()
		ctx = context.WithValue(ctx, "carbonzipper_uuid", uuid)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func MarshalCtx(ctx context.Context, response *http.Request) *http.Request {
	response.Header.Add(ctx_header_prefix+"UUID", ifaceToString(ctx.Value("carbonzipper_uuid")))

	return response
}
