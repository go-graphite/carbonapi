package graphite

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ansel1/merry"

	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/protocols/graphite/msgpack"
	"github.com/go-graphite/carbonapi/zipper/types"

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

// RoundRobin is used to connect to backends inside clientGroups, implements BackendServer interface
type GraphiteGroup struct {
	groupName string
	servers   []string
	protocol  string

	client *http.Client

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	httpQuery *helper.HttpQuery
}

func (g *GraphiteGroup) Children() []types.BackendServer {
	return []types.BackendServer{g}
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "graphite"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

	httpClient := helper.GetHTTPClient(logger, config)

	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &GraphiteGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: *config.MaxBatchSize,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool) (types.BackendServer, merry.Error) {
	if config.ConcurrencyLimit == nil {
		return nil, types.ErrConcurrencyLimitNotSet
	}
	if len(config.Servers) == 0 {
		return nil, types.ErrNoServersSpecified
	}
	limiter := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, tldCacheDisabled, requireSuccessAll, limiter)
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

func (c *GraphiteGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/render/")

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error
	for pathExpr, targets := range pathExprToTargets {
		v := url.Values{
			"target": targets,
			"format": []string{c.protocol},
			"from":   []string{strconv.Itoa(int(request.Metrics[0].StartTime))},
			"until":  []string{strconv.Itoa(int(request.Metrics[0].StopTime))},
		}
		rewrite.RawQuery = v.Encode()
		stats.RenderRequests++
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			stats.RenderErrors++
			if merry.Is(err, types.ErrTimeoutExceeded) {
				stats.Timeouts++
				stats.RenderTimeouts++
			}
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		metrics := msgpack.MultiGraphiteFetchResponse{}
		_, marshalErr := metrics.UnmarshalMsg(res.Response)
		if marshalErr != nil {
			stats.RenderErrors++
			if e == nil {
				e = merry.Wrap(marshalErr).WithValue("targets", targets)
			} else {
				e = e.WithCause(err)
			}
			continue
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

	if e != nil {
		stats.FailedServers = []string{c.groupName}
		logger.Error("errors occurred while getting results",
			zap.Any("error", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *GraphiteGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	var r protov3.MultiGlobResponse
	r.Metrics = make([]protov3.GlobResponse, 0)
	var e merry.Error
	for _, query := range request.Metrics {
		v := url.Values{
			"query":  []string{query},
			"format": []string{c.protocol},
		}
		rewrite.RawQuery = v.Encode()
		stats.FindRequests++
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			stats.FindErrors++
			if merry.Is(err, types.ErrTimeoutExceeded) {
				stats.Timeouts++
				stats.FindTimeouts++
			}
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}
		var globs msgpack.MultiGraphiteGlobResponse
		_, marshalErr := globs.UnmarshalMsg(res.Response)
		if marshalErr != nil {
			if e == nil {
				e = merry.Wrap(marshalErr).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
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

	if e != nil {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *GraphiteGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "info"))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/info/")

	var r protov3.ZipperInfoResponse
	var e merry.Error
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
		stats.InfoRequests++
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			stats.InfoErrors++
			if merry.Is(err, types.ErrTimeoutExceeded) {
				stats.Timeouts++
				stats.InfoTimeouts++
			}
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		var info protov2.InfoResponse
		marshalErr := info.Unmarshal(res.Response)
		if marshalErr != nil {
			stats.InfoErrors++
			if e == nil {
				e = merry.Wrap(marshalErr).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
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

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	if e != nil {
		stats.FailedServers = []string{c.groupName}
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *GraphiteGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (c *GraphiteGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}

func (c *GraphiteGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, merry.Error) {
	logger := c.logger
	var rewrite *url.URL
	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
		rewrite, _ = url.Parse("http://127.0.0.1/tags/autoComplete/tags")
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		rewrite, _ = url.Parse("http://127.0.0.1/tags/autoComplete/values")
	}

	var r []string

	rewrite.RawQuery = query
	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return r, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		e = merry.Wrap(err)
		return r, e
	}

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	return r, nil
}

func (c *GraphiteGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, true, query, limit)
}

func (c *GraphiteGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, false, query, limit)
}

func (c *GraphiteGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
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
