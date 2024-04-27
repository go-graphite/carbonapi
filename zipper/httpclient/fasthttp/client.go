package fasthttp

import (
	"context"
	"net/http"

	"github.com/ansel1/merry/v2"
	"go.uber.org/zap"

	"github.com/valyala/fasthttp"

	"github.com/go-graphite/carbonapi/internal/dns"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pkg/tlsconfig"
	util "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/types"
)

type HTTPClient struct {
	logger *zap.Logger
	config *types.BackendV2

	Client *fasthttp.Client
}

func (c *HTTPClient) Do(ctx context.Context, logger *zap.Logger, server, uri string,
	r types.Request, encoding string) (*types.ServerResponse,
	error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI(server + uri)
	req.Header.SetMethodBytes([]byte(fasthttp.MethodPost))
	req.Header.Set("Accept", encoding)
	req = util.FastHTTPMarshalPassHeaders(ctx, util.FastHTTPMarshalCtx(ctx, util.FastHTTPMarshalCtx(ctx, req, util.HeaderUUIDZipper), util.HeaderUUIDAPI))
	if r != nil {
		body, err2 := r.Marshal()
		if err2 != nil {
			return nil, merry.Wrap(err2, merry.WithValue("server", server))
		}

		// assuming Request has a body that is of type []byte
		req.SetBody(body)
	}

	// call fasthttp request
	err := c.Client.Do(req, res)
	if err != nil {
		logger.Error("Error while making a http request", zap.Error(err))
		return nil, err
	}

	switch res.StatusCode() {
	case http.StatusNotFound:
		return &types.ServerResponse{Server: server}, nil
	case http.StatusOK:
		body, err := res.BodyUncompressed()
		if err != nil {
			return nil, merry.Wrap(err,
				merry.WithHTTPCode(res.StatusCode()),
				merry.WithValue("server", server),
			)
		}
		// TODO: actually unmarshal data to avoid copy
		b := make([]byte, len(body))
		copy(b, body)

		return &types.ServerResponse{Server: server, Response: b}, nil
	default:
		return nil, merry.Wrap(types.ErrFailedToFetch,
			merry.WithValue("server", server),
			merry.WithHTTPCode(res.StatusCode()))
	}
}

func New(logger *zap.Logger, config *types.BackendV2, _ limiter.ServerLimiter) *HTTPClient {
	logger = logger.With(zap.String("type", "fasthttp"))
	logger.Debug("creating new fasthttp client")
	logger.Warn("fasthttp client is experimental and might not work well")
	c := &fasthttp.Client{
		MaxConnsPerHost: *config.ConcurrencyLimit,
		Dial:            dns.GetFastHTTPDialFunc(),
		DialTimeout:     dns.GetFastHTTPDialFuncWithTimeout(config.Timeouts.Connect, *config.KeepAliveInterval),
	}
	if config.TLSClientConfig != nil {
		tlsConfig, warns, err := tlsconfig.ParseClientTLSConfig(config.TLSClientConfig)
		if err != nil {
			logger.Fatal("failed to initialize client",
				zap.Error(err),
			)
		}
		if len(warns) > 0 {
			logger.Warn("insecure options detected, while parsing HTTP Client TLS Config for backed",
				zap.Strings("warnings", warns),
			)
		}
		c.TLSConfig = tlsConfig
	}

	return &HTTPClient{
		logger: logger,
		config: config,
		Client: c,
	}
}
