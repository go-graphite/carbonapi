package util

import (
	"context"
	"net/http"
	"go.uber.org/zap"
	"time"
	"github.com/go-graphite/carbonapi/helper/carbonapipb"
)

type key int

const (
	ctxHeaderUUID = "X-CTX-CarbonAPI-UUID"

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

func SetupDeferredAccessLogging(accessLogger *zap.Logger, accessLogDetails *carbonapipb.AccessLogDetails, t time.Time, logAsError bool)  {
	accessLogDetails.Runtime = time.Since(t).String()
	if logAsError {
		accessLogger.Error("Request failed", zap.Any("data", *accessLogDetails))
	} else {
		accessLogDetails.HttpCode = http.StatusOK
		accessLogger.Info("Request served", zap.Any("data", *accessLogDetails))
	}
}

func ParseCtx(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		uuid := req.Header.Get(ctxHeaderUUID)

		ctx := req.Context()
		ctx = SetUUID(ctx, uuid)

		h.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func MarshalCtx(ctx context.Context, response *http.Request) *http.Request {
	response.Header.Add(ctxHeaderUUID, GetUUID(ctx))

	return response
}
