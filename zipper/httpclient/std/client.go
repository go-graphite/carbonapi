package std

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/internal/dns"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pkg/tlsconfig"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/types"
)

type HTTPClient struct {
	logger    *zap.Logger
	groupName string
	servers   []string
	maxTries  int
	limiter   limiter.ServerLimiter

	counter uint64

	httpClient *http.Client
}

func (c *HTTPClient) Do(ctx context.Context, logger *zap.Logger, server, uri string, r types.Request, encoding string) (*types.ServerResponse,
	error) {
	logger = logger.With(
		zap.String("function", "HttpQuery.doRequest"),
	)

	u, err := url.Parse(server + uri)
	if err != nil {
		return nil, merry.Wrap(err, merry.WithValue("server", server))
	}

	var reader io.Reader
	var body []byte
	if r != nil {
		body, err = r.Marshal()
		if err != nil {
			return nil, merry.Wrap(err, merry.WithValue("server", server))
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

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), reader)
	if err != nil {
		return nil, merry.Wrap(err, merry.WithValue("server", server))
	}

	req.Header.Set("Accept", encoding)
	req = util.MarshalPassHeaders(ctx, util.MarshalCtx(ctx, util.MarshalCtx(ctx, req, util.HeaderUUIDZipper), util.HeaderUUIDAPI))

	logger.Debug("trying to get slot",
		zap.String("name", server),
	)
	err = c.limiter.Enter(ctx, server)
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		return nil, merry.Wrap(err, merry.WithValue("server", server))
	}

	defer c.limiter.Leave(ctx, server)

	logger.Debug("got slot for server",
		zap.String("name", server),
	)

	if r != nil {
		logger = logger.With(zap.Any("payloadData", r.LogInfo()))
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		logger.Debug("error fetching result",
			zap.Error(err),
		)

		return nil, helper.RequestError(err, server)

	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// we don't need to process any further if the response is empty.
	if resp.StatusCode == http.StatusNotFound {
		return &types.ServerResponse{Server: server}, nil
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("error reading body",
			zap.Error(err),
		)
		return nil, merry.Wrap(err, merry.WithValue("server", server))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, merry.Wrap(types.ErrFailedToFetch, merry.WithValue("server", server), merry.WithMessage(string(body)), merry.WithHTTPCode(resp.StatusCode))
	}

	return &types.ServerResponse{Server: server, Response: body}, nil
}

func New(logger *zap.Logger, config *types.BackendV2, l limiter.ServerLimiter) *HTTPClient {
	logger.Debug("creating new HTTP client",
		zap.Any("config", config),
	)

	transport := &http.Transport{
		MaxConnsPerHost:     *config.ConcurrencyLimit,
		MaxIdleConnsPerHost: *config.MaxIdleConnsPerHost,
		IdleConnTimeout:     *config.IdleConnectionTimeout,
		ForceAttemptHTTP2:   config.ForceAttemptHTTP2,
		DialContext:         dns.GetDialContextWithTimeout(config.Timeouts.Connect, *config.KeepAliveInterval),
	}

	if config.TLSClientConfig != nil {
		tlsConfig, warns, err := tlsconfig.ParseClientTLSConfig(config.TLSClientConfig)
		if err != nil {
			logger.Fatal("failed to initialize client for group",
				zap.String("group_name", config.GroupName),
				zap.Error(err),
			)
		}
		if len(warns) > 0 {
			logger.Warn("insecure options detected, while parsing HTTP Client TLS Config for backed",
				zap.String("group_name", config.GroupName),
				zap.Strings("warnings", warns),
			)
		}
		transport.TLSClientConfig = tlsConfig
	}

	c := &HTTPClient{
		logger: logger,
		httpClient: &http.Client{
			Transport: transport,
		},
		limiter: l,
	}

	return c
}
