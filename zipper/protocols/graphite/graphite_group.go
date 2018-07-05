package v2

import (
	"context"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/protocols/graphite/msgpack"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"go.uber.org/zap"
)

func init() {
	aliases := []string{"msgpack"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGroups, implements ServerClient interface
type GraphiteGroup struct {
	groupName string
	servers   []string
	protocol  string

	client *http.Client

	counter             uint64
	maxIdleConnsPerHost int

	limiter              *limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	httpQuery *helper.HttpQuery
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, limiter *limiter.ServerLimiter) (types.ServerClient, *errors.Errors) {
	logger = logger.With(zap.String("type", "graphite"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

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

	httpQuery := helper.NewHttpQuery(logger, config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &GraphiteGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
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

func New(logger *zap.Logger, config types.BackendV2) (types.ServerClient, *errors.Errors) {
	if config.ConcurrencyLimit == nil {
		return nil, errors.Fatal("concurency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, errors.Fatal("no servers specified")
	}
	limiter := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, limiter)
}

func (c GraphiteGroup) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c GraphiteGroup) Name() string {
	return c.groupName
}

func (c GraphiteGroup) Backends() []string {
	return c.servers
}

func (c *GraphiteGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, *errors.Errors) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/render/")

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	for pathExpr, targets := range pathExprToTargets {
		v := url.Values{
			"target": targets,
			"format": []string{c.protocol},
			"from":   []string{strconv.Itoa(int(request.Metrics[0].StartTime))},
			"until":  []string{strconv.Itoa(int(request.Metrics[0].StopTime))},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
		if err == nil {
			err = &errors.Errors{}
		}
		if err.HaveFatalErrors {
			err.HaveFatalErrors = false
			return nil, stats, err
		}

		var metrics msgpack.MultiGraphiteFetchResponse
		_, e := metrics.UnmarshalMsg(res.Response)
		err.AddFatal(e)
		if err.HaveFatalErrors {
			return nil, stats, err
		}

		for _, m := range metrics {
			vals := make([]float64, len(m.Values))
			for i, vIface := range m.Values {
				if v, ok := vIface.(float64); ok {
					vals[i] = v
				} else {
					vals[i] = math.NaN()
				}
			}
			r.Metrics = append(r.Metrics, protov3.FetchResponse{
				Name:              m.Name,
				PathExpression:    pathExpr,
				ConsolidationFunc: "Average",
				StopTime:          int64(m.End),
				StartTime:         int64(m.Start),
				StepTime:          int64(m.Step),
				Values:            vals,
				XFilesFactor:      0.0,
			})
		}
	}

	return &r, stats, nil
}

func (c *GraphiteGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	var r protov3.MultiGlobResponse
	r.Metrics = make([]protov3.GlobResponse, 0)
	var e errors.Errors
	for _, query := range request.Metrics {
		v := url.Values{
			"query":  []string{query},
			"format": []string{c.protocol},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
		if err != nil {
			e.Merge(err)
			continue
		}
		var globs msgpack.MultiGraphiteGlobResponse
		_, marshalErr := globs.UnmarshalMsg(res.Response)
		if marshalErr != nil {
			e.Add(marshalErr)
			continue
		}

		stats.Servers = append(stats.Servers, res.Server)
		matches := make([]protov3.GlobMatch, 0, len(globs))
		for _, m := range globs {
			matches = append(matches, protov3.GlobMatch{
				Path:   m.Path,
				IsLeaf: m.IsLeaf,
			})
		}
		r.Metrics = append(r.Metrics, protov3.GlobResponse{
			Name:    query,
			Matches: matches,
		})
	}

	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
	}

	if len(r.Metrics) == 0 {
		return nil, stats, errors.FromErr(types.ErrNoResponseFetched)
	}
	return &r, stats, nil
}

func (c *GraphiteGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	logger := c.logger.With(zap.String("type", "info"))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/info/")

	var r protov3.ZipperInfoResponse
	var e errors.Errors
	r.Info = make(map[string]protov3.MultiMetricsInfoResponse)
	data := protov3.MultiMetricsInfoResponse{}
	server := c.groupName
	if len(c.servers) == 1 {
		server = c.servers[0]
	}

	for _, query := range request.Names {
		v := url.Values{
			"target": []string{query},
			"format": []string{c.protocol},
		}
		rewrite.RawQuery = v.Encode()
		res, e2 := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
		if e2 != nil {
			e.Merge(e2)
			continue
		}

		var info protov2.InfoResponse
		err := info.Unmarshal(res.Response)
		if err != nil {
			e.Add(err)
			continue
		}
		stats.Servers = append(stats.Servers, res.Server)

		if info.AggregationMethod == "" {
			info.AggregationMethod = "average"
		}
		infoV3 := protov3.MetricsInfoResponse{
			Name:              info.Name,
			ConsolidationFunc: info.AggregationMethod,
			XFilesFactor:      info.XFilesFactor,
			MaxRetention:      int64(info.MaxRetention),
		}

		for _, r := range info.Retentions {
			newR := protov3.Retention{
				SecondsPerPoint: int64(r.SecondsPerPoint),
				NumberOfPoints:  int64(r.NumberOfPoints),
			}
			infoV3.Retentions = append(infoV3.Retentions, newR)
		}

		data.Metrics = append(data.Metrics, infoV3)
	}
	r.Info[server] = data

	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
	}

	if len(r.Info[server].Metrics) == 0 {
		return nil, stats, errors.FromErr(types.ErrNoResponseFetched)
	}

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	return &r, stats, nil
}

func (c *GraphiteGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (c *GraphiteGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

func (c *GraphiteGroup) ProbeTLDs(ctx context.Context) ([]string, *errors.Errors) {
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
