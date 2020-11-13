package victoriametrics

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/valyala/fastjson"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"

	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus"

	"go.uber.org/zap"
)

func init() {
	aliases := []string{"victoriametrics", "vm"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGroups, implements BackendServer interface
type VictoriaMetricsGroup struct {
	types.BackendServer

	groupName string
	servers   []string
	protocol  string

	client *http.Client

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	step              int64
	maxPointsPerQuery int64

	startDelay prometheus.StartDelay

	httpQuery  *helper.HttpQuery
	parserPool fastjson.ParserPool
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "victoriametrics"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *config.MaxIdleConnsPerHost,
			DialContext: (&net.Dialer{
				Timeout:   config.Timeouts.Connect,
				KeepAlive: *config.KeepAliveInterval,
			}).DialContext,
		},
	}

	step := int64(15)
	stepI, ok := config.BackendOptions["step"]
	if ok {
		stepNew, ok := stepI.(string)
		if ok {
			if stepNew[len(stepNew)-1] >= '0' && stepNew[len(stepNew)-1] <= '9' {
				stepNew += "s"
			}
			t, err := time.ParseDuration(stepNew)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "step"),
					zap.String("option_value", stepNew),
					zap.Error(err),
				)
			}
			step = int64(t.Seconds())
		} else {
			logger.Fatal("failed to parse step",
				zap.String("type_parsed", fmt.Sprintf("%T", stepI)),
				zap.String("type_expected", "string"),
			)
		}
	}

	maxPointsPerQuery := int64(11000)
	mppqI, ok := config.BackendOptions["max_points_per_query"]
	if ok {
		mppq, ok := mppqI.(int)
		if !ok {
			logger.Fatal("failed to parse max_points_per_query",
				zap.String("type_parsed", fmt.Sprintf("%T", mppqI)),
				zap.String("type_expected", "int"),
			)
		}

		maxPointsPerQuery = int64(mppq)
	}

	delay := prometheus.StartDelay{
		IsSet:      false,
		IsDuration: false,
		T:          -1,
	}
	startI, ok := config.BackendOptions["start"]
	if ok {
		delay.IsSet = true
		startNew, ok := startI.(string)
		if ok {
			startNewInt, err := strconv.Atoi(startNew)
			if err != nil {
				d, err2 := time.ParseDuration(startNew)
				if err2 != nil {
					logger.Fatal("failed to parse option",
						zap.String("option_name", "start"),
						zap.String("option_value", startNew),
						zap.Errors("errors", []error{err, err2}),
					)
				}
				delay.IsDuration = true
				delay.D = d
			} else {
				delay.T = int64(startNewInt)
			}
		}
	}

	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &VictoriaMetricsGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: *config.MaxBatchSize,
		step:                 step,
		maxPointsPerQuery:    maxPointsPerQuery,
		startDelay:           delay,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}

	promLogger := logger.With(zap.String("subclass", "prometheus"))
	c.BackendServer, _ = prometheus.NewWithEverythingInitialized(promLogger, config, tldCacheDisabled, limiter, step, maxPointsPerQuery, delay, httpQuery, httpClient)
	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2, tldCacheDisabled bool) (types.BackendServer, merry.Error) {
	if config.ConcurrencyLimit == nil {
		return nil, types.ErrConcurrencyLimitNotSet
	}
	if len(config.Servers) == 0 {
		return nil, types.ErrNoServersSpecified
	}
	l := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, tldCacheDisabled, l)
}

func (c *VictoriaMetricsGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	var r protov3.MultiGlobResponse
	var e merry.Error

	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/expand/")

	r.Metrics = make([]protov3.GlobResponse, 0)
	parser := c.parserPool.Get()
	defer c.parserPool.Put(parser)

	for _, query := range request.Metrics {
		v := url.Values{
			"query":  []string{query},
			"format": []string{"json"},
		}

		rewrite.RawQuery = v.Encode()
		stats.FindRequests += 1
		res, queryErr := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if queryErr != nil {
			stats.FindErrors += 1
			if merry.Is(queryErr, types.ErrTimeoutExceeded) {
				stats.Timeouts += 1
				stats.FindTimeouts += 1
			}
			if e == nil {
				e = merry.Wrap(queryErr).WithValue("query", query)
			} else {
				e = e.WithCause(queryErr)
			}
			continue
		}

		parsedJson, err := parser.ParseBytes(res.Response)
		if err != nil {
			if e == nil {
				e = merry.Wrap(err).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		globs, err := parsedJson.Array()
		if err != nil {
			if e == nil {
				e = merry.Wrap(err).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		stats.Servers = append(stats.Servers, res.Server)
		matches := make([]protov3.GlobMatch, 0, len(globs))
		var path string
		for _, m := range globs {
			b, _ := m.StringBytes()
			isLeaf := true
			if bytes.HasSuffix(b, []byte{'.'}) {
				isLeaf = false
				path = string(b[:len(b)-1])
			} else {
				path = string(b)
			}
			matches = append(matches, protov3.GlobMatch{
				Path:   path,
				IsLeaf: isLeaf,
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

func (c *VictoriaMetricsGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
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
