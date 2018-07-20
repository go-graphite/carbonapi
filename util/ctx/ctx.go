package ctx

import (
	"context"
	"net/http"
)

type key int

const (
	HeaderUUIDAPI    = "X-CTX-CarbonAPI-UUID"
	HeaderUUIDZipper = "X-CTX-CarbonZipper-UUID"

	uuidKey key = 0
)

func ifaceToString(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

func getCtxString(ctx context.Context, k key) string {
	return ifaceToString(ctx.Value(k))
}

func GetUUID(ctx context.Context) string {
	return getCtxString(ctx, uuidKey)
}

func SetUUID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, uuidKey, v)
}

func ParseCtx(h http.HandlerFunc, uuidKey string) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		uuid := req.Header.Get(uuidKey)

		ctx := req.Context()
		ctx = SetUUID(ctx, uuid)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func MarshalCtx(ctx context.Context, response *http.Request, uuidKey string) *http.Request {
	response.Header.Add(uuidKey, GetUUID(ctx))

	return response
}
