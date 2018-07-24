package helper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/go-graphite/carbonapi/limiter"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/errors"
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
	logger    *zap.Logger
	limiter   *limiter.ServerLimiter
	client    *http.Client
	encoding  string

	counter uint64
}

func NewHttpQuery(logger *zap.Logger, groupName string, servers []string, maxTries int, limiter *limiter.ServerLimiter, client *http.Client, encoding string) *HttpQuery {
	return &HttpQuery{
		groupName: groupName,
		servers:   servers,
		maxTries:  maxTries,
		logger:    logger.With(zap.String("action", "query")),
		limiter:   limiter,
		client:    client,
		encoding:  encoding,
	}
}

func (c *HttpQuery) pickServer() string {
	if len(c.servers) == 1 {
		// No need to do heavy operations here
		return c.servers[0]
	}
	logger := c.logger.With(zap.String("function", "picker"))
	counter := atomic.AddUint64(&(c.counter), 1)
	idx := counter % uint64(len(c.servers))
	srv := c.servers[int(idx)]
	logger.Debug("picked",
		zap.Uint64("counter", counter),
		zap.Uint64("idx", idx),
		zap.String("Server", srv),
	)

	return srv
}

func (c *HttpQuery) doRequest(ctx context.Context, uri string, r types.Request) (*ServerResponse, error) {
	server := c.pickServer()
	c.logger.Debug("picked server",
		zap.String("server", server),
	)

	u, err := url.Parse(server + uri)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	var body []byte
	if r != nil {
		body, err = r.Marshal()
		if err != nil {
			return nil, err
		}
		if body != nil {
			reader = bytes.NewReader(body)
		}
	}
	logger := c.logger.With(
		zap.String("server", server),
		zap.String("name", c.groupName),
		zap.String("uri", u.String()),
	)

	req, err := http.NewRequest("GET", u.String(), reader)
	req.Header.Set("Accept", c.encoding)
	if err != nil {
		return nil, err
	}
	req = util.MarshalCtx(ctx, util.MarshalCtx(ctx, req, util.HeaderUUIDZipper), util.HeaderUUIDAPI)

	logger.Debug("trying to get slot")

	err = c.limiter.Enter(ctx, c.groupName)
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		return nil, err
	}
	logger.Debug("got slot")
	logger = logger.With(zap.Any("payloadData", r.LogInfo()))

	resp, err := c.client.Do(req.WithContext(ctx))
	c.limiter.Leave(ctx, server)
	if err != nil {
		logger.Error("error fetching result",
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("error reading body",
			zap.Error(err),
		)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("status not ok",
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf(types.ErrFailedToFetchFmt, c.groupName, resp.StatusCode, string(body))
	}

	return &ServerResponse{Server: server, Response: body}, nil
}

func (c *HttpQuery) DoQuery(ctx context.Context, uri string, r types.Request) (*ServerResponse, *errors.Errors) {
	maxTries := c.maxTries
	if len(c.servers) > maxTries {
		maxTries = len(c.servers)
	}

	var e errors.Errors
	for try := 0; try < maxTries; try++ {
		res, err := c.doRequest(ctx, uri, r)
		if err != nil {
			c.logger.Error("have errors",
				zap.Error(err),
			)
			e.Add(err)
			if ctx.Err() != nil {
				e.HaveFatalErrors = true
				return nil, &e
			}
			continue
		}

		return res, nil
	}

	e.AddFatal(types.ErrMaxTriesExceeded)
	return nil, &e
}
