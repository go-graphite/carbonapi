package httpclient

import (
	"context"
	"errors"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/httpclient/fasthttp"
	"github.com/go-graphite/carbonapi/zipper/httpclient/std"
	"github.com/go-graphite/carbonapi/zipper/types"
)

type Client interface {
	Do(ctx context.Context, logger *zap.Logger, server, uri string, r types.Request, encoding string) (*types.ServerResponse,
		error)
}

func GetClient(logger *zap.Logger, config *types.BackendV2, l limiter.ServerLimiter) (Client, error) {
	if config.FetchClientType == "" {
		return std.New(logger, config, l), nil
	}
	switch config.FetchClientType {
	case "std":
		return std.New(logger, config, l), nil
	case "fasthttp":
		return fasthttp.New(logger, config, l), nil
	default:
		return nil, merry.Wrap(errors.New("unsupported http client"))
	}
}
