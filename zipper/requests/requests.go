package requests

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/httpclient"
	"github.com/go-graphite/carbonapi/zipper/types"
)

type HttpQuery struct {
	groupName string
	servers   []string
	maxTries  int
	limiter   limiter.ServerLimiter
	client    httpclient.Client
	encoding  string

	counter uint64
}

func NewHttpQuery(logger *zap.Logger, config *types.BackendV2, limiter limiter.ServerLimiter,
	encoding string) (*HttpQuery, error) {
	httpClient, err := httpclient.GetClient(logger, config, limiter)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return &HttpQuery{
		groupName: config.GroupName,
		servers:   config.Servers,
		maxTries:  *config.MaxTries,
		limiter:   limiter,
		client:    httpClient,
		encoding:  encoding,
	}, nil
}

func (c *HttpQuery) pickServer(logger *zap.Logger) string {
	if len(c.servers) == 1 {
		// No need to do heavy operations here
		return c.servers[0]
	}
	logger = logger.With(zap.String("function", "picker"))
	counter := atomic.AddUint64(&(c.counter), 1)
	idx := counter % uint64(len(c.servers))
	srv := c.servers[int(idx)]
	logger.Debug("picked",
		zap.Uint64("counter", counter),
		zap.Uint64("idx", idx),
		zap.String("server", srv),
	)

	return srv
}

func (c *HttpQuery) DoQuery(ctx context.Context, logger *zap.Logger, uri string, r types.Request) (resp *types.ServerResponse, err error) {
	maxTries := c.maxTries
	if len(c.servers) > maxTries {
		maxTries = len(c.servers)
	}

	e := merry.Wrap(types.ErrFailedToFetch, merry.WithValue("uri", uri))
	code := http.StatusInternalServerError
	for try := 0; try < maxTries; try++ {
		server := c.pickServer(logger)
		res, err := c.client.Do(ctx, logger, server, uri, r, c.encoding)
		if err != nil {
			logger.Debug("have errors",
				zap.Error(err),
				zap.String("server", server),
			)

			e = merry.Wrap(e, merry.WithCause(err), merry.WithHTTPCode(merry.HTTPCode(err)))
			code = merry.HTTPCode(err)
			// TODO (msaf1980): may be metric for server failures ?
			// TODO (msaf1980): may be retry policy for avoid retry bad queries ?
			continue
		}

		return res, nil
	}

	return nil, merry.Wrap(types.ErrMaxTriesExceeded,
		merry.WithCause(e),
		merry.WithHTTPCode(code),
	)
}

func (c *HttpQuery) DoQueryToAll(ctx context.Context, logger *zap.Logger, uri string, r types.Request) (resp []*types.ServerResponse, err error) {
	maxTries := c.maxTries
	if len(c.servers) > maxTries {
		maxTries = len(c.servers)
	}

	res := make([]*types.ServerResponse, len(c.servers))
	e := merry.Wrap(types.ErrFailedToFetch,
		merry.WithValue("uri", uri),
	)
	responseCount := 0
	code := http.StatusInternalServerError
	for i := range c.servers {
		for try := 0; try < maxTries; try++ {
			response, err := c.client.Do(ctx, logger, c.servers[i], uri, r, c.encoding)
			if err != nil {
				logger.Debug("have errors",
					zap.Error(err),
				)

				e = merry.Wrap(e, merry.WithCause(err))
				code = merry.HTTPCode(err)
				continue
			}

			res[i] = response
			responseCount++
			break
		}
	}

	if responseCount == len(c.servers) {
		return res, nil
	}

	return res, merry.Wrap(types.ErrMaxTriesExceeded,
		merry.WithCause(e),
		merry.WithHTTPCode(code),
	)
}
