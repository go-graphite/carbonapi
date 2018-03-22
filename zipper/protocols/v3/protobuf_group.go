package v3

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/zipper/helper"
	"github.com/go-graphite/carbonzipper/zipper/metadata"
	"github.com/go-graphite/carbonzipper/zipper/types"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
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

	limiter  limiter.ServerLimiter
	logger   *zap.Logger
	timeout  types.Timeouts
	maxTries int

	httpQuery *helper.HttpQuery
}

func NewWithLimiter(config types.BackendV2, limiter limiter.ServerLimiter) (types.ServerClient, error) {
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

	httpQuery := helper.NewHttpQuery(logger, config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient)

	c := &ClientProtoV3Group{
		groupName: config.GroupName,
		servers:   config.Servers,
		timeout:   *config.Timeouts,
		maxTries:  *config.MaxTries,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func New(config types.BackendV2) (types.ServerClient, error) {
	if config.ConcurrencyLimit == nil {
		return nil, fmt.Errorf("concurency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, fmt.Errorf("no servers specified")
	}
	limiter := limiter.NewServerLimiter(config.Servers, *config.ConcurrencyLimit)

	return NewWithLimiter(config, limiter)
}

func (c ClientProtoV3Group) Name() string {
	return c.groupName
}

func (c ClientProtoV3Group) Backends() []string {
	return c.servers
}

func (c *ClientProtoV3Group) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, error) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/render/")

	var targets []string
	for _, m := range request.Metrics {
		targets = append(targets, m.Name)
	}

	v := url.Values{
		"target": targets,
		"format": []string{"protobuf"},
		"from":   []string{strconv.Itoa(int(request.Metrics[0].StartTime))},
		"until":  []string{strconv.Itoa(int(request.Metrics[0].StopTime))},
	}
	rewrite.RawQuery = v.Encode()
	res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI())
	if err != nil {
		return nil, stats, err
	}

	var metrics protov2.MultiFetchResponse
	err = metrics.Unmarshal(res.Response)
	if err != nil {
		return nil, stats, err
	}

	stats.Servers = append(stats.Servers, res.Server)

	var r protov3.MultiFetchResponse
	for _, m := range metrics.Metrics {
		r.Metrics = append(r.Metrics, protov3.FetchResponse{
			Name:              m.Name,
			ConsolidationFunc: "average",
			StopTime:          uint32(m.StopTime),
			StartTime:         uint32(m.StartTime),
			StepTime:          uint32(m.StepTime),
			Values:            m.Values,
		})
	}

	return &r, stats, nil
}

func (c *ClientProtoV3Group) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, error) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	var r protov3.MultiGlobResponse
	r.Metrics = make([]protov3.GlobResponse, 0)
	var errors []error
	for _, query := range request.Metrics {
		v := url.Values{
			"query":  []string{query},
			"format": []string{"protobuf"},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI())
		if err != nil {
			errors = append(errors, err)
			continue
		}
		var globs protov2.GlobResponse
		err = globs.Unmarshal(res.Response)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		stats.Servers = append(stats.Servers, res.Server)
		matches := make([]protov3.GlobMatch, 0, len(globs.Matches))
		for _, m := range globs.Matches {
			matches = append(matches, protov3.GlobMatch{
				Path:   m.Path,
				IsLeaf: m.IsLeaf,
			})
		}
		r.Metrics = append(r.Metrics, protov3.GlobResponse{
			Name:    globs.Name,
			Matches: matches,
		})
	}

	if len(errors) != 0 {
		strErrors := make([]string, 0, len(errors))
		for _, e := range errors {
			strErrors = append(strErrors, e.Error())
		}
		c.logger.Error("errors occurred while getting results",
			zap.Strings("errors", strErrors),
		)
	}

	if len(r.Metrics) == 0 {
		return nil, stats, types.ErrNoResponseFetched
	}
	return &r, stats, nil
}

func (c *ClientProtoV3Group) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, error) {
	return nil, nil, types.ErrNotImplementedYet
}

func (c *ClientProtoV3Group) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (c *ClientProtoV3Group) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, error) {
	return nil, nil, types.ErrNotImplementedYet
}

func (c *ClientProtoV3Group) ProbeTLDs(ctx context.Context) ([]string, error) {
	logger := c.logger.With(zap.String("function", "prober"))
	req := &protov3.MultiGlobRequest{
		Metrics: []string{"*"},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Find)
	defer cancel()

	logger.Debug("doing request",
		zap.Any("request", req),
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
