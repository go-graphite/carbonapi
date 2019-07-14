package v2

import (
	"context"
	"encoding/json"
	"github.com/ansel1/merry"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"go.uber.org/zap"
)

const (
	format = "protobuf"
)

func init() {
	aliases := []string{"carbonapi_v2_pb", "proto_v2_pb", "v2_pb", "pb", "pb3", "protobuf", "protobuf3"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGroups, implements BackendServer interface
type ClientProtoV2Group struct {
	groupName string
	servers   []string

	client *http.Client

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	httpQuery *helper.HttpQuery
}

func (c *ClientProtoV2Group) Children() []types.BackendServer {
	return []types.BackendServer{c}
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, l limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "protoV2Group"), zap.String("name", config.GroupName))

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

	httpLimiter := limiter.NewServerLimiter(config.Servers, *config.ConcurrencyLimit)
	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, httpLimiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &ClientProtoV2Group{
		groupName:            config.GroupName,
		servers:              config.Servers,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: config.MaxBatchSize,

		client:  httpClient,
		limiter: l,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2) (types.BackendServer, merry.Error) {
	if config.ConcurrencyLimit == nil {
		return nil, merry.New("concurrency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, merry.New("no servers specified")
	}
	limiter := limiter.NewServerLimiter(config.Servers, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, limiter)
}

func (c ClientProtoV2Group) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c ClientProtoV2Group) Name() string {
	return c.groupName
}

func (c ClientProtoV2Group) Backends() []string {
	return c.servers
}

type queryBatch struct {
	pathExpression string
	from           int64
	until          int64
}

func (c *ClientProtoV2Group) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/render/")

	batches := make(map[queryBatch][]string)
	for _, m := range request.Metrics {
		b := queryBatch{
			pathExpression: m.PathExpression,
			from:           m.StartTime,
			until:          m.StopTime,
		}

		batches[b] = append(batches[b], m.Name)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error
	for batch, targets := range batches {
		v := url.Values{
			"target": targets,
			"format": []string{format},
			"from":   []string{strconv.Itoa(int(batch.from))},
			"until":  []string{strconv.Itoa(int(batch.until))},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		var metrics protov2.MultiFetchResponse
		marshalErr := metrics.Unmarshal(res.Response)
		if marshalErr != nil {
			if e == nil {
				e = err
			} else {
				e = e.WithCause(marshalErr)
			}
			continue
		}

		for _, m := range metrics.Metrics {
			for i, v := range m.IsAbsent {
				if v {
					m.Values[i] = math.NaN()
				}
			}
			r.Metrics = append(r.Metrics, protov3.FetchResponse{
				Name:              m.Name,
				PathExpression:    batch.pathExpression,
				ConsolidationFunc: "Average",
				StopTime:          int64(m.StopTime),
				StartTime:         int64(m.StartTime),
				StepTime:          int64(m.StepTime),
				Values:            m.Values,
				XFilesFactor:      0.0,
				RequestStartTime:  batch.from,
				RequestStopTime:   batch.until,
			})
		}
	}

	if e != nil {
		logger.Warn("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *ClientProtoV2Group) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	var r protov3.MultiGlobResponse
	r.Metrics = make([]protov3.GlobResponse, 0)
	var e merry.Error
	for _, query := range request.Metrics {
		logger.Debug("will do query",
			zap.String("query", query),
		)
		v := url.Values{
			"query":  []string{query},
			"format": []string{format},
		}
		rewrite.RawQuery = v.Encode()
		logger.Debug("doing http query")
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		logger.Debug("done http query")
		if err != nil {
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}
		var globs protov2.GlobResponse
		logger.Debug("started to unmarshal response")
		marshalErr := globs.Unmarshal(res.Response)
		logger.Debug("done unmarshal response")
		if marshalErr != nil {
			if e == nil {
				e = err
			} else {
				e = e.WithCause(marshalErr)
			}
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
		if len(matches) != 0 {
			r.Metrics = append(r.Metrics, protov3.GlobResponse{
				Name:    globs.Name,
				Matches: matches,
			})
		}
	}

	if e != nil {
		logger.Warn("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return nil, stats, e
	}
	return &r, stats, nil
}

func (c *ClientProtoV2Group) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
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
			"format": []string{format},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
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
			if e == nil {
				e = err
			} else {
				e = e.WithCause(marshalErr)
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

	if e != nil {
		logger.Warn("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}

	logger.Debug("got client response",
		zap.Any("response", r),
	)

	return &r, stats, nil
}

func (c *ClientProtoV2Group) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, merry.Error) {
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
		return r, merry.Wrap(err)
	}

	logger.Debug("got client response",
		zap.Strings("response", r),
	)

	return r, nil
}

func (c *ClientProtoV2Group) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, true, query, limit)
}

func (c *ClientProtoV2Group) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, false, query, limit)
}

func (c *ClientProtoV2Group) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (c *ClientProtoV2Group) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}

func (c *ClientProtoV2Group) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
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
