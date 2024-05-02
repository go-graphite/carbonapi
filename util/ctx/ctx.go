package ctx

import (
	"context"
	"net/http"

	"github.com/valyala/fasthttp"
)

type key int

const (
	HeaderUUIDAPI    = "X-CTX-CarbonAPI-UUID"
	HeaderUUIDZipper = "X-CTX-CarbonZipper-UUID"

	uuidKey key = iota
	headersToPassKey
	headersToLogKey
	maxDataPoints
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

func getCtxInt64(ctx context.Context, k key) int64 {
	v := ctx.Value(k)
	if v != nil {
		return v.(int64)
	}
	return 0
}

func getCtxMapString(ctx context.Context, k key) map[string]string {
	v := ctx.Value(k)
	if v != nil {
		vv, ok := v.(map[string]string)
		if !ok {
			return map[string]string{}
		}
		return vv
	}
	return map[string]string{}
}

func GetPassHeaders(ctx context.Context) map[string]string {
	return getCtxMapString(ctx, headersToPassKey)
}

func SetPassHeaders(ctx context.Context, h map[string]string) context.Context {
	return context.WithValue(ctx, headersToPassKey, h)
}

func GetLogHeaders(ctx context.Context) map[string]string {
	return getCtxMapString(ctx, headersToLogKey)
}

func SetLogHeaders(ctx context.Context, h map[string]string) context.Context {
	return context.WithValue(ctx, headersToLogKey, h)
}

func GetUUID(ctx context.Context) string {
	return getCtxString(ctx, uuidKey)
}

func SetUUID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, uuidKey, v)
}

func SetMaxDatapoints(ctx context.Context, h int64) context.Context {
	return context.WithValue(ctx, maxDataPoints, h)
}

func GetMaxDatapoints(ctx context.Context) int64 {
	return getCtxInt64(ctx, maxDataPoints)
}

func ParseCtx(h http.HandlerFunc, uuidKey string) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		uuid := req.Header.Get(uuidKey)

		ctx := req.Context()
		ctx = SetUUID(ctx, uuid)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func FastHTTPMarshalPassHeaders(ctx context.Context, response *fasthttp.Request) *fasthttp.Request {
	headers := GetPassHeaders(ctx)
	for name, value := range headers {
		response.Header.Add(name, value)
	}

	return response
}

func FastHTTPMarshalCtx(ctx context.Context, response *fasthttp.Request, uuidKey string) *fasthttp.Request {
	response.Header.Add(uuidKey, GetUUID(ctx))

	return response
}

func MarshalPassHeaders(ctx context.Context, response *http.Request) *http.Request {
	headers := GetPassHeaders(ctx)
	for name, value := range headers {
		response.Header.Add(name, value)
	}

	return response
}

func MarshalCtx(ctx context.Context, response *http.Request, uuidKey string) *http.Request {
	response.Header.Add(uuidKey, GetUUID(ctx))

	return response
}
