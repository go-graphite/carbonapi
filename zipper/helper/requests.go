package helper

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pkg/parser"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/types"
	"go.uber.org/zap"
)

func HttpCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if _, ok := err.(*net.OpError); ok {
		return http.StatusServiceUnavailable
	}
	if urlErr, ok := err.(*url.Error); ok {
		if _, ok = urlErr.Err.(*net.OpError); ok {
			return http.StatusServiceUnavailable
		}
	}
	return http.StatusInternalServerError
}

func MergeHttpErrors(errors []merry.Error) (int, []string) {
	returnCode := http.StatusNotFound
	errMsgs := make([]string, 0)
	for _, err := range errors {
		var code int
		c := merry.RootCause(err)
		if c == nil {
			c = err
		}
		code = merry.HTTPCode(c)
		u := merry.Unwrap(c)
		_ = u

		if merry.HTTPCode(c) == 404 || merry.Is(c, parser.ErrSeriesDoesNotExist) {
			continue
		}
		errMsgs = append(errMsgs, c.Error())
		if code == 400 {
			// The 400 is returned on wrong requests, e.g. non-existent functions
			returnCode = code
		} else if returnCode == 404 {
			// First error
			returnCode = code
		} else if code == 503 && (returnCode == 504 || returnCode == 502) {
			returnCode = 503
		} else if (code < 502 || code > 504) && (returnCode >= 502 || returnCode <= 504) {
			returnCode = code
		}
	}

	return returnCode, errMsgs
}

func MergeHttpErrorMap(errors map[string]merry.Error) (int, []string) {
	returnCode := http.StatusNotFound
	errMsgs := make([]string, 0)
	for _, err := range errors {
		var code int
		c := merry.RootCause(err)
		if c == nil {
			c = err
		}
		code = merry.HTTPCode(c)

		if merry.HTTPCode(c) == 404 || merry.Is(c, parser.ErrSeriesDoesNotExist) {
			continue
		}
		errMsgs = append(errMsgs, c.Error())
		if code == 400 {
			// The 400 is returned on wrong requests, e.g. non-existent functions
			returnCode = code
		} else if returnCode == 404 {
			// First error
			returnCode = code
		} else if code == 503 && (returnCode == 504 || returnCode == 502) {
			returnCode = 503
		} else if (code < 502 || code > 504) && (returnCode >= 502 || returnCode <= 504) {
			returnCode = code
		}
	}

	return returnCode, errMsgs
}

func HttpErrorByCode(err merry.Error) merry.Error {
	var returnErr merry.Error
	if err == nil {
		returnErr = types.ErrNoMetricsFetched.WithHTTPCode(404)
	} else {
		code := merry.HTTPCode(err)
		if code == 403 {
			returnErr = types.ErrForbidden.WithHTTPCode(403)
		} else if code >= 502 || code <= 504 {
			returnErr = types.ErrFailedToFetch.WithHTTPCode(code)
		} else {
			returnErr = types.ErrNoMetricsFetched.WithHTTPCode(code)
		}
	}

	return returnErr
}

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

func (c *HttpQuery) doRequest(ctx context.Context, logger *zap.Logger, server, uri string, r types.Request) (*ServerResponse, merry.Error) {
	logger = logger.With(
		zap.String("function", "HttpQuery.doRequest"),
	)

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

	// TODO: change to NewRequestWithContext
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
		return nil, merry.Here(err).WithValue("server", server).WithHTTPCode(HttpCode(err))
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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
		server := c.pickServer(logger)
		res, err := c.doRequest(ctx, logger, server, uri, r)
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

func (c *HttpQuery) DoQueryToAll(ctx context.Context, logger *zap.Logger, uri string, r types.Request) ([]*ServerResponse, merry.Error) {
	maxTries := c.maxTries
	if len(c.servers) > maxTries {
		maxTries = len(c.servers)
	}

	res := make([]*ServerResponse, len(c.servers))
	e := types.ErrFailedToFetch.WithValue("uri", uri)
	responseCount := 0
	for i := range c.servers {
		for try := 0; try < maxTries; try++ {
			response, err := c.doRequest(ctx, logger, c.servers[i], uri, r)
			if err != nil {
				logger.Debug("have errors",
					zap.Error(err),
				)

				e = e.WithCause(err)
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

	return res, types.ErrMaxTriesExceeded.WithCause(e)
}
