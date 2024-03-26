package zipper

import (
	"context"
	_ "net/http/pprof"
	"time"

	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"

	utilctx "github.com/go-graphite/carbonapi/util/ctx"
	"github.com/go-graphite/carbonapi/zipper/broadcast"
	"github.com/go-graphite/carbonapi/zipper/config"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"

	_ "github.com/go-graphite/carbonapi/zipper/protocols/auto"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/graphite"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/irondb"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/prometheus"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v2"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v3"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/victoriametrics"
)

// Zipper provides interface to Zipper-related functions
type Zipper struct {
	probeTicker *time.Ticker
	ProbeQuit   chan struct{}
	ProbeForce  chan int

	timeout           time.Duration
	timeoutConnect    time.Duration
	keepAliveInterval time.Duration

	// Will broadcast to all servers there
	backend                   types.BackendServer
	concurrencyLimitPerServer int

	ScaleToCommonStep bool

	sendStats func(*types.Stats)

	logger *zap.Logger
}

func createBackendsV2(logger *zap.Logger, backends types.BackendsV2, expireDelaySec int32, tldCacheDisabled, requireSuccessAll bool) ([]types.BackendServer, merry.Error) {
	backendServers := make([]types.BackendServer, 0)
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

		var backendServer types.BackendServer
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
			backendServer, e = backendInit(logger, backend, tldCacheDisabled, requireSuccessAll)
			if e != nil {
				return nil, e
			}
		} else {
			config := backend

			backendServers := make([]types.BackendServer, 0, len(backend.Servers))
			for _, server := range backend.Servers {
				config.Servers = []string{server}
				config.GroupName = server
				backendServer, e = backendInit(logger, config, tldCacheDisabled, requireSuccessAll)
				if e != nil {
					return nil, e
				}
				backendServers = append(backendServers, backendServer)
			}

			backendServer, err = broadcast.NewBroadcastGroup(logger, backend.GroupName, backend.DoMultipleRequestsIfSplit, backendServers,
				expireDelaySec, *backend.ConcurrencyLimit, *backend.MaxBatchSize, timeouts, tldCacheDisabled, requireSuccessAll,
			)
			if err != nil {
				return nil, merry.Wrap(err)
			}
		}
		backendServers = append(backendServers, backendServer)
	}
	return backendServers, nil
}

// NewZipper allows to create new Zipper
func NewZipper(sender func(*types.Stats), cfg *config.Config, logger *zap.Logger) (*Zipper, merry.Error) {
	if !cfg.IsSanitized() {
		cfg = config.SanitizeConfig(logger, *cfg)
	}

	backends, err := createBackendsV2(logger, cfg.BackendsV2, int32(cfg.InternalRoutingCache.Seconds()), cfg.TLDCacheDisabled, cfg.RequireSuccessAll)
	if err != nil {
		logger.Fatal("errors while initialing zipper store backend",
			zap.Any("error", err),
		)
	}

	logger.Error("DEBUG ERROR LOGGGGG", zap.Any("cfg", cfg))
	broadcastGroup, err := broadcast.NewBroadcastGroup(logger, "root", cfg.DoMultipleRequestsIfSplit, backends,
		int32(cfg.InternalRoutingCache.Seconds()), cfg.ConcurrencyLimitPerServer, *cfg.MaxBatchSize, cfg.Timeouts, cfg.TLDCacheDisabled, cfg.RequireSuccessAll,
	)
	if err != nil {
		logger.Fatal("error while initialing zipper store backend",
			zap.Any("error", err),
		)
	}

	z := &Zipper{
		ProbeQuit:  make(chan struct{}),
		ProbeForce: make(chan int),

		ScaleToCommonStep: cfg.ScaleToCommonStep,
		sendStats:         sender,

		backend:                   broadcastGroup,
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

	_, err := z.backend.ProbeTLDs(ctx)
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
	logger := z.logger.With(zap.String("function", "FetchProtoV3"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))

	res, stats, e := z.backend.Fetch(ctx, request)

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

		err := helper.HttpErrorByCode(e)

		return nil, stats, err
	}

	return res, stats, merry.WithHTTPCode(e, 200)
}

func (z Zipper) FindProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "FindProtoV3"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))

	res, stats, err := z.backend.Find(ctx, request)

	var errs []merry.Error
	if err != nil {
		errs = []merry.Error{err}
	}

	findResponse := &types.ServerFindResponse{
		Response: res,
		Stats:    stats,
		Err:      errs,
	}

	if len(findResponse.Err) > 0 {
		var e merry.Error
		if len(findResponse.Err) == 1 {
			e = helper.HttpErrorByCode(findResponse.Err[0])
		} else {
			e = helper.HttpErrorByCode(findResponse.Err[1].WithCause(findResponse.Err[0]))
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
	logger := z.logger.With(zap.String("function", "InfoProtoV3"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))
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

	r, stats, e := z.backend.Info(ctx, realRequest)
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
	logger := z.logger.With(zap.String("function", "ListProtoV3"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))
	r, stats, e := z.backend.List(ctx)
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
	logger := z.logger.With(zap.String("function", "StatsProtoV3"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))
	r, stats, e := z.backend.Stats(ctx)
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
	logger := z.logger.With(zap.String("function", "TagNames"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))
	data, err := z.backend.TagNames(ctx, query, limit)
	if err != nil {
		logger.Debug("had errors while fetching result",
			zap.Any("errors", err),
		)
		return data, err
	}

	return data, nil
}

func (z Zipper) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	logger := z.logger.With(zap.String("function", "TagValues"), zap.String("carbonapi_uuid", utilctx.GetUUID(ctx)))
	data, err := z.backend.TagValues(ctx, query, limit)
	if err != nil {
		logger.Debug("had errors while fetching result",
			zap.Any("errors", err),
		)
		return data, err
	}

	return data, nil
}
