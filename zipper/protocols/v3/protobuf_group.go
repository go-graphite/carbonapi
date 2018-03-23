package v3

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/zipper/errors"
	"github.com/go-graphite/carbonzipper/zipper/helper"
	"github.com/go-graphite/carbonzipper/zipper/httpHeaders"
	"github.com/go-graphite/carbonzipper/zipper/metadata"
	"github.com/go-graphite/carbonzipper/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

func init() {
	aliases := []string{"carbonapi_v3_pb", "proto_v3_pb", "v3_pb"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGroups, implements ServerClient interface
type ClientProtoV3Group struct {
	groupName string
	servers   []string

	client *http.Client

	counter             uint64
	maxIdleConnsPerHost int

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	httpQuery *helper.HttpQuery
}

func New(config types.BackendV2) (types.ServerClient, *errors.Errors) {
	if config.ConcurrencyLimit == nil || *config.ConcurrencyLimit == 0 {
		return nil, errors.Fatal("concurency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, errors.Fatal("no servers specified")
	}
	limiter := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(config, limiter)
}

func NewWithLimiter(config types.BackendV2, limiter limiter.ServerLimiter) (types.ServerClient, *errors.Errors) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *config.MaxIdleConnsPerHost,
			DialContext: (&net.Dialer{
				Timeout:   config.Timeouts.Connect,
				KeepAlive: *config.KeepAliveInterval,
				DualStack: true,
			}).DialContext,
		},
	}

	logger := zapwriter.Logger("protobufGroup").With(zap.String("name", config.GroupName))

	httpQuery := helper.NewHttpQuery(logger, config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv3PB)

	c := &ClientProtoV3Group{
		groupName:            config.GroupName,
		servers:              config.Servers,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: config.MaxGlobs,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func (c ClientProtoV3Group) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c ClientProtoV3Group) Name() string {
	return c.groupName
}

func (c ClientProtoV3Group) Backends() []string {
	return c.servers
}

func (c *ClientProtoV3Group) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, *errors.Errors) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/render/")

	var targets []string
	for _, m := range request.Metrics {
		targets = append(targets, m.Name)
	}

	v := url.Values{
		"format": []string{"carbonapi_v3_pb"},
	}
	rewrite.RawQuery = v.Encode()

	data, err := request.Marshal()
	if err != nil {
		return nil, nil, errors.FromErrNonFatal(err)
	}

	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), data)
	if err != nil {
		return nil, stats, e
	}

	var metrics protov3.MultiFetchResponse
	e.AddFatal(metrics.Unmarshal(res.Response))
	if e == nil {
		e = &errors.Errors{}
	}
	if e.HaveFatalErrors {
		return nil, stats, e
	}

	return &metrics, stats, nil
}

func (c *ClientProtoV3Group) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	v := url.Values{
		"format": []string{"carbonapi_v3_pb"},
	}
	rewrite.RawQuery = v.Encode()

	data, err := request.Marshal()
	if err != nil {
		return nil, nil, errors.FromErrNonFatal(err)
	}

	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), data)
	if e == nil {
		e = &errors.Errors{}
	}

	var globs protov3.MultiGlobResponse
	err = globs.Unmarshal(res.Response)
	if err != nil {
		return nil, nil, errors.FromErrNonFatal(err)
	}

	return &globs, stats, nil
}

func (c *ClientProtoV3Group) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	v := url.Values{
		"format": []string{"carbonapi_v3_pb"},
	}
	rewrite.RawQuery = v.Encode()

	data, err := request.Marshal()
	if err != nil {
		return nil, nil, errors.FromErrNonFatal(err)
	}

	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), data)
	if e == nil {
		e = &errors.Errors{}
	}

	var infos protov3.MultiMetricsInfoResponse
	err = infos.Unmarshal(res.Response)
	if err != nil {
		return nil, nil, errors.FromErrNonFatal(err)
	}

	stats.MemoryUsage = int64(infos.Size())

	r := &protov3.ZipperInfoResponse{
		Info: map[string]protov3.MultiMetricsInfoResponse{
			c.Name(): infos,
		},
	}

	return r, stats, nil
}

func (c *ClientProtoV3Group) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (c *ClientProtoV3Group) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

func (c *ClientProtoV3Group) ProbeTLDs(ctx context.Context) ([]string, *errors.Errors) {
	logger := c.logger.With(zap.String("function", "prober"))
	req := &protov3.MultiGlobRequest{
		Metrics: []string{"*"},
	}

	logger.Debug("doing request",
		zap.Strings("request", req.Metrics),
	)

	res, _, err := c.Find(ctx, req)
	if err != nil {
		return nil, err
	}

	var tlds []string
	for _, m := range res.Metrics {
		for _, v := range m.Matches {
			tlds = append(tlds, v.Path)
		}
	}

	logger.Debug("will return data",
		zap.Strings("tlds", tlds),
	)

	return tlds, nil
}
