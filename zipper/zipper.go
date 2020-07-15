package zipper

import (
	"context"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/zipper/broadcast"
	"github.com/go-graphite/carbonapi/zipper/config"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"

	_ "github.com/go-graphite/carbonapi/zipper/protocols/auto"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/graphite"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/prometheus"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v2"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v3"
)

// Zipper provides interface to Zipper-related functions
type Zipper struct {
	probeTicker *time.Ticker
	ProbeQuit   chan struct{}
	ProbeForce  chan int

	timeout           time.Duration
	timeoutConnect    time.Duration
	keepAliveInterval time.Duration

	searchConfigured bool
	searchBackends   types.BackendServer
	searchPrefix     string

	// Will broadcast to all servers there
	storeBackends             types.BackendServer
	concurrencyLimitPerServer int

	sendStats func(*types.Stats)

	logger *zap.Logger
}

func createBackendsV2(logger *zap.Logger, backends types.BackendsV2, expireDelaySec int32, tldCacheDisabled bool) ([]types.BackendServer, merry.Error) {
	storeClients := make([]types.BackendServer, 0)
	var e merry.Error
	timeouts := backends.Timeouts
	for _, backend := range backends.Backends {
		concurrencyLimit := backends.ConcurrencyLimitPerServer
		tries := backends.MaxTries
		maxIdleConnsPerHost := backends.MaxIdleConnsPerHost
		keepAliveInterval := backends.KeepAliveInterval
		maxBatchSize := backends.MaxBatchSize

		if backend.Timeouts == nil {
			backend.Timeouts = &timeouts
		}
		if backend.ConcurrencyLimit == nil {
			backend.ConcurrencyLimit = &concurrencyLimit
		}
		if backend.MaxTries == nil {
			backend.MaxTries = &tries
		}
		if backend.MaxBatchSize == nil {
			backend.MaxBatchSize = maxBatchSize
		}
		if backend.MaxIdleConnsPerHost == nil {
			backend.MaxIdleConnsPerHost = &maxIdleConnsPerHost
		}
		if backend.KeepAliveInterval == nil {
			backend.KeepAliveInterval = &keepAliveInterval
		}

		var client types.BackendServer
		logger.Debug("creating lb group",
			zap.String("name", backend.GroupName),
			zap.Strings("servers", backend.Servers),
			zap.Any("type", backend.LBMethod),
		)

		metadata.Metadata.RLock()
		backendInit, ok := metadata.Metadata.ProtocolInits[backend.Protocol]
		metadata.Metadata.RUnlock()
		if !ok {
			var protocols []string
			metadata.Metadata.RLock()
			for p := range metadata.Metadata.SupportedProtocols {
				protocols = append(protocols, p)
			}
			metadata.Metadata.RUnlock()
			logger.Error("unknown backend protocol",
				zap.Any("backend", backend),
				zap.String("requested_protocol", backend.Protocol),
				zap.Strings("supported_backends", protocols),
			)
			return nil, merry.Errorf("unknown backend protocol '%v'", backend.Protocol)
		}

		var lbMethod types.LBMethod
		err := lbMethod.FromString(backend.LBMethod)
		if err != nil {
			logger.Fatal("failed to parse lbMethod",
				zap.String("lbMethod", backend.LBMethod),
				zap.Error(err),
			)
		}
		if lbMethod == types.RoundRobinLB {
			client, e = backendInit(logger, backend, tldCacheDisabled)
			if e != nil {
				return nil, e
			}
		} else {
			config := backend

			backends := make([]types.BackendServer, 0, len(backend.Servers))
			for _, server := range backend.Servers {
				config.Servers = []string{server}
				config.GroupName = server
				client, e = backendInit(logger, config, tldCacheDisabled)
				if e != nil {
					return nil, e
				}
				backends = append(backends, client)
			}

			client, e = broadcast.NewBroadcastGroup(logger, backend.GroupName, backends, expireDelaySec, *backend.ConcurrencyLimit, *backend.MaxBatchSize, timeouts, tldCacheDisabled)
			if e != nil {
				return nil, e
			}
		}
		storeClients = append(storeClients, client)
	}
	return storeClients, nil
}

// NewZipper allows to create new Zipper
func NewZipper(sender func(*types.Stats), cfg *config.Config, logger *zap.Logger) (*Zipper, merry.Error) {
	if !cfg.IsSanitized() {
		cfg = config.SanitizeConfig(logger, *cfg)
	}

	var searchBackends types.BackendServer
	var prefix string

	if len(cfg.CarbonSearchV2.BackendsV2.Backends) > 0 {
		prefix = cfg.CarbonSearchV2.Prefix
		searchClients, err := createBackendsV2(logger, cfg.CarbonSearchV2.BackendsV2, int32(cfg.InternalRoutingCache.Seconds()), cfg.TLDCacheDisabled)
		if err != nil {
			logger.Fatal("merry.Errors while initialing zipper search backends",
				zap.Any("merry.Errors", err),
			)
		}

		searchBackends, err = broadcast.NewBroadcastGroup(logger, "search", searchClients, int32(cfg.InternalRoutingCache.Seconds()), cfg.ConcurrencyLimitPerServer, *cfg.MaxBatchSize, cfg.Timeouts, cfg.TLDCacheDisabled)
		if err != nil {
			logger.Fatal("merry.Errors while initialing zipper search backends",
				zap.Any("merry.Errors", err),
			)
		}
	}

	storeClients, err := createBackendsV2(logger, cfg.BackendsV2, int32(cfg.InternalRoutingCache.Seconds()), cfg.TLDCacheDisabled)
	if err != nil {
		logger.Fatal("merry.Errors while initialing zipper store backends",
			zap.Any("merry.Errors", err),
		)
	}

	var storeBackends types.BackendServer
	storeBackends, err = broadcast.NewBroadcastGroup(logger, "root", storeClients, int32(cfg.InternalRoutingCache.Seconds()), cfg.ConcurrencyLimitPerServer, *cfg.MaxBatchSize, cfg.Timeouts, cfg.TLDCacheDisabled)
	if err != nil {
		logger.Fatal("merry.Errors while initialing zipper store backends",
			zap.Any("merry.Errors", err),
		)
	}

	z := &Zipper{
		ProbeQuit:  make(chan struct{}),
		ProbeForce: make(chan int),

		sendStats: sender,

		storeBackends:             storeBackends,
		searchBackends:            searchBackends,
		searchPrefix:              prefix,
		searchConfigured:          len(prefix) > 0 && len(searchBackends.Backends()) > 0,
		concurrencyLimitPerServer: cfg.ConcurrencyLimitPerServer,
		keepAliveInterval:         cfg.KeepAliveInterval,
		timeout:                   cfg.Timeouts.Render,
		timeoutConnect:            cfg.Timeouts.Connect,
		logger:                    logger,
	}

	logger.Debug("zipper config",
		zap.Any("config", cfg),
	)

	if !cfg.TLDCacheDisabled {
		z.probeTicker = time.NewTicker(cfg.InternalRoutingCache)

		go z.probeTlds()

		z.ProbeForce <- 1
	}
	return z, nil
}

func (z *Zipper) doProbe(logger *zap.Logger) {
	ctx := context.Background()

	_, err := z.storeBackends.ProbeTLDs(ctx)
	if err != nil {
		logger.Error("failed to probe tlds",
			zap.String("errors", err.Cause().Error()),
		)
		if ce := logger.Check(zap.DebugLevel, "failed to probe tlds (verbose)"); ce != nil {
			ce.Write(
				zap.Any("errorVerbose", err),
			)
		}
	}
}

func (z *Zipper) probeTlds() {
	logger := z.logger.With(zap.String("type", "probe"))
	for {
		select {
		case <-z.probeTicker.C:
			z.doProbe(logger)
		case <-z.ProbeForce:
			z.doProbe(logger)
		case <-z.ProbeQuit:
			z.probeTicker.Stop()
			return
		}
	}
}

// GRPC-compatible methods
func (z Zipper) FetchProtoV3(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "FetchProtoV3"))
	var statsSearch *types.Stats
	var e merry.Error

	if z.searchConfigured {
		realRequest := &protov3.MultiFetchRequest{
			Metrics: make([]protov3.FetchRequest, 0, len(request.Metrics)),
		}

		for _, metric := range request.Metrics {
			if strings.HasPrefix(metric.Name, z.searchPrefix) {
				res, stat, err := z.searchBackends.Find(ctx, &protov3.MultiGlobRequest{
					Metrics: []string{metric.Name},
				})

				if statsSearch == nil {
					statsSearch = stat
				} else {
					statsSearch.Merge(stat)
				}

				if err != nil {
					if e == nil {
						e = err
					} else {
						e = e.WithCause(err)
					}
					continue
				}

				if len(res.Metrics) == 0 {
					continue
				}

				metricRequests := make([]protov3.FetchRequest, 0, len(res.Metrics))
				for _, n := range res.Metrics {
					for _, m := range n.Matches {
						metricRequests = append(metricRequests, protov3.FetchRequest{
							Name:            m.Path,
							StartTime:       metric.StartTime,
							StopTime:        metric.StopTime,
							FilterFunctions: metric.FilterFunctions,
						})
					}
				}

				if len(metricRequests) > 0 {
					realRequest.Metrics = append(realRequest.Metrics, metricRequests...)
				}

			} else {
				realRequest.Metrics = append(realRequest.Metrics, metric)
			}
		}

		if len(realRequest.Metrics) > 0 {
			request = realRequest
		}
	}

	res, stats, err := z.storeBackends.Fetch(ctx, request)
	if statsSearch != nil {
		if stats == nil {
			stats = statsSearch
		} else {
			stats.Merge(statsSearch)
		}
	}

	if e == nil {
		e = err
	} else {
		e = e.WithCause(err)
	}
	if e != nil {
		logger.Debug("had errors while fetching result",
			zap.Any("errors", e),
			zap.Int("httpCode", merry.HTTPCode(e)),
		)
	}
	if res == nil || len(res.Metrics) == 0 {
		logger.Debug("no metrics fetched",
			zap.Any("errors", e),
		)

		err := types.ErrNoMetricsFetched
		if e != nil {
			err = err.WithHTTPCode(merry.HTTPCode(e))
		} else {
			err = err.WithHTTPCode(404)
		}

		return nil, stats, err
	}

	return res, stats, merry.WithHTTPCode(e, 200)
}

func (z Zipper) FindProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "FindProtoV3"))
	searchRequests := &protov3.MultiGlobRequest{}
	if z.searchConfigured {
		realRequest := &protov3.MultiGlobRequest{Metrics: make([]string, 0, len(request.Metrics))}
		for _, m := range request.Metrics {
			if strings.HasPrefix(m, z.searchPrefix) {
				searchRequests.Metrics = append(searchRequests.Metrics, m)
			} else {
				realRequest.Metrics = append(realRequest.Metrics, m)
			}
		}
		if len(searchRequests.Metrics) > 0 {
			request = realRequest
		}
	}

	res, stats, err := z.storeBackends.Find(ctx, request)

	var errs []merry.Error
	if err != nil {
		errs = []merry.Error{err}
	}

	findResponse := &types.ServerFindResponse{
		Response: res,
		Stats:    stats,
		Err:      errs,
	}

	// TODO(civil): Rework merging carbonsearch and other responses
	if len(searchRequests.Metrics) > 0 {
		resSearch, statsSearch, err := z.searchBackends.Find(ctx, request)
		var errs []merry.Error
		if err != nil {
			errs = []merry.Error{err}
		}
		searchResponse := &types.ServerFindResponse{
			Response: resSearch,
			Stats:    statsSearch,
			Err:      errs,
		}
		_ = findResponse.Merge(searchResponse)
	}

	if len(findResponse.Err) > 0 {
		var e merry.Error
		if len(findResponse.Err) == 1 {
			e = findResponse.Err[0]
		} else {
			e = findResponse.Err[1].WithCause(findResponse.Err[0])
		}
		logger.Debug("had errors while fetching result",
			zap.Any("errors", e),
		)
		// TODO(civil): Not Found error cases across all zipper code should be handled in the same way
		// See FetchProtoV3 for more examples
		if findResponse.Response != nil && len(findResponse.Response.Metrics) > 0 {
			return findResponse.Response, findResponse.Stats, merry.WithHTTPCode(e, 200)
		}
		return nil, stats, e
	}

	return findResponse.Response, findResponse.Stats, nil
}

func (z Zipper) InfoProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "InfoProtoV3"))
	realRequest := &protov3.MultiMetricsInfoRequest{Names: make([]string, 0, len(request.Metrics))}
	res, _, err := z.FindProtoV3(ctx, request)
	if err == nil || merry.Is(err, types.ErrNonFatalErrors) {
		for _, m := range res.Metrics {
			for _, match := range m.Matches {
				if match.IsLeaf {
					realRequest.Names = append(realRequest.Names, match.Path)
				}
			}
		}
	} else {
		realRequest.Names = append(realRequest.Names, request.Metrics...)
	}

	r, stats, e := z.storeBackends.Info(ctx, realRequest)
	if e != nil {
		if merry.Is(e, types.ErrNotFound) {
			return nil, nil, e
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", e),
			)
			return nil, stats, e
		}
	}

	return r, stats, nil
}

func (z Zipper) ListProtoV3(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "ListProtoV3"))
	r, stats, e := z.storeBackends.List(ctx)
	if e != nil {
		if merry.Is(e, types.ErrNotFound) {
			return nil, nil, e
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", e),
			)
			return r, stats, e
		}
	}

	return r, stats, e
}
func (z Zipper) StatsProtoV3(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "StatsProtoV3"))
	r, stats, e := z.storeBackends.Stats(ctx)
	if e != nil {
		if merry.Is(e, types.ErrNotFound) {
			return nil, nil, e
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", e),
			)
			return r, stats, e
		}
	}

	return r, stats, nil
}

// Tags

func (z Zipper) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	logger := z.logger.With(zap.String("function", "TagNames"))
	data, err := z.storeBackends.TagNames(ctx, query, limit)
	if err != nil {
		logger.Debug("had errors while fetching result",
			zap.Any("errors", err),
		)
		return data, err
	}

	return data, nil
}

func (z Zipper) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	logger := z.logger.With(zap.String("function", "TagValues"))
	data, err := z.storeBackends.TagValues(ctx, query, limit)
	if err != nil {
		logger.Debug("had errors while fetching result",
			zap.Any("errors", err),
		)
		return data, err
	}

	return data, nil
}
