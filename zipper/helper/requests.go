package helper

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/limiter"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/types"
	"go.uber.org/zap"
)

type ServerResponse struct {
	Server   string
	Response []byte
}

type HttpQuery struct {
	groupName string
	servers   []string
	maxTries  int
	limiter   limiter.ServerLimiter
	client    *http.Client
	encoding  string

	counter uint64
}

func NewHttpQuery(groupName string, servers []string, maxTries int, limiter limiter.ServerLimiter, client *http.Client, encoding string) *HttpQuery {
	return &HttpQuery{
		groupName: groupName,
		servers:   servers,
		maxTries:  maxTries,
		limiter:   limiter,
		client:    client,
		encoding:  encoding,
	}
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

func (c *HttpQuery) doRequest(ctx context.Context, logger *zap.Logger, uri string, r types.Request) (*ServerResponse, merry.Error) {
	logger = logger.With(
		zap.String("function", "HttpQuery.doRequest"),
	)
	server := c.pickServer(logger)

	u, err := url.Parse(server + uri)
	if err != nil {
		return nil, merry.Here(err).WithValue("server", server)
	}

	var reader io.Reader
	var body []byte
	if r != nil {
		body, err = r.Marshal()
		if err != nil {
			return nil, merry.Here(err).WithValue("server", server)
		}
		if body != nil {
			reader = bytes.NewReader(body)
		}
	}
	logger = logger.With(
		zap.String("server", server),
		zap.String("name", c.groupName),
		zap.String("uri", u.String()),
	)

	req, err := http.NewRequest("GET", u.String(), reader)
	if err != nil {
		return nil, merry.Here(err).WithValue("server", server)
	}

	req.Header.Set("Accept", c.encoding)
	req = util.MarshalPassHeaders(ctx, util.MarshalCtx(ctx, util.MarshalCtx(ctx, req, util.HeaderUUIDZipper), util.HeaderUUIDAPI))

	logger.Debug("trying to get slot",
		zap.String("name", server),
	)
	err = c.limiter.Enter(ctx, server)
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		return nil, merry.Here(err).WithValue("server", server)
	}

	defer c.limiter.Leave(ctx, server)

	logger.Debug("got slot for server",
		zap.String("name", server),
	)

	if r != nil {
		logger = logger.With(zap.Any("payloadData", r.LogInfo()))
	}
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		logger.Debug("error fetching result",
			zap.Error(err),
		)
		return nil, merry.Here(err).WithValue("server", server)
	}
	defer resp.Body.Close()

	// we don't need to process any further if the response is empty.
	if resp.StatusCode == http.StatusNotFound {
		return &ServerResponse{Server: server}, nil
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("error reading body",
			zap.Error(err),
		)
		return nil, merry.Here(err).WithValue("server", server)
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, types.ErrFailedToFetch.Here().WithValue("group", c.groupName).WithValue("status_code", resp.StatusCode).WithValue("body", string(body))
	}

	return &ServerResponse{Server: server, Response: body}, nil
}

func (c *HttpQuery) DoQuery(ctx context.Context, logger *zap.Logger, uri string, r types.Request) (*ServerResponse, merry.Error) {
	maxTries := c.maxTries
	if len(c.servers) > maxTries {
		maxTries = len(c.servers)
	}

	e := types.ErrFailedToFetch.WithValue("uri", uri)
	for try := 0; try < maxTries; try++ {
		res, err := c.doRequest(ctx, logger, uri, r)
		if err != nil {
			logger.Debug("have errors",
				zap.Error(err),
			)

			e = e.WithCause(err)
			continue
		}

		return res, nil
	}

	return nil, types.ErrMaxTriesExceeded.WithCause(e)
}
