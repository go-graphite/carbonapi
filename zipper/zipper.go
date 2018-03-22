package zipper

import (
	"context"
	"fmt"
	"math"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/pathcache"
	"github.com/go-graphite/carbonzipper/zipper/broadcast"
	"github.com/go-graphite/carbonzipper/zipper/metadata"
	"github.com/go-graphite/carbonzipper/zipper/types"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"

	_ "github.com/go-graphite/carbonzipper/zipper/protocols/grpc"
	_ "github.com/go-graphite/carbonzipper/zipper/protocols/v2"
)

// Zipper provides interface to Zipper-related functions
type Zipper struct {
	// Limiter limits our concurrency to a particular server
	limiter     limiter.ServerLimiter
	probeTicker *time.Ticker
	ProbeQuit   chan struct{}
	ProbeForce  chan int

	timeout           time.Duration
	timeoutConnect    time.Duration
	timeoutKeepAlive  time.Duration
	keepAliveInterval time.Duration

	searchConfigured bool
	searchBackends   types.ServerClient
	searchPrefix     string

	searchCache pathcache.PathCache

	// Will broadcast to all servers there
	storeBackends             types.ServerClient
	concurrencyLimitPerServer int

	sendStats func(*types.Stats)

	logger *zap.Logger
}

type nameLeaf struct {
	name string
	leaf bool
}

var defaultTimeouts = types.Timeouts{
	Render:  10000 * time.Second,
	Find:    100 * time.Second,
	Connect: 200 * time.Millisecond,
}

func sanitizeTimouts(timeouts, defaultTimeouts types.Timeouts) types.Timeouts {
	if timeouts.Render == 0 {
		timeouts.Render = defaultTimeouts.Render
	}
	if timeouts.Find == 0 {
		timeouts.Find = defaultTimeouts.Find
	}

	if timeouts.Connect == 0 {
		timeouts.Connect = defaultTimeouts.Connect
	}

	return timeouts
}

func createBackendsV2(logger *zap.Logger, backends types.BackendsV2, pathCache pathcache.PathCache) ([]types.ServerClient, error) {
	storeClients := make([]types.ServerClient, 0)
	var err error
	for _, backend := range backends.Backends {
		concurencyLimit := backends.ConcurrencyLimitPerServer
		timeouts := backends.Timeouts
		tries := backends.MaxTries
		maxIdleConnsPerHost := backends.MaxIdleConnsPerHost
		keepAliveInterval := backends.KeepAliveInterval

		if backend.Timeouts == nil {
			backend.Timeouts = &timeouts
		} else {
			timeouts := sanitizeTimouts(*backend.Timeouts, backends.Timeouts)
			backend.Timeouts = &timeouts
		}
		if backend.ConcurrencyLimit == nil {
			backend.ConcurrencyLimit = &concurencyLimit
		}
		if backend.MaxTries == nil {
			backend.MaxTries = &tries
		}
		if backend.MaxIdleConnsPerHost == nil {
			backend.MaxIdleConnsPerHost = &maxIdleConnsPerHost
		}
		if backend.KeepAliveInterval == nil {
			backend.KeepAliveInterval = &keepAliveInterval
		}

		var client types.ServerClient
		logger.Debug("creating lb group",
			zap.String("name", backend.GroupName),
			zap.Strings("servers", backend.Servers),
			zap.Any("type", backend.LBMethod),
		)
		if backend.LBMethod == types.RoundRobinLB {
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
				return nil, fmt.Errorf("unknown backend protocol '%v'", backend.Protocol)
			}
			client, err = backendInit(backend)
			if err != nil {
				return nil, err
			}
		} else {
			config := backend
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
					zap.Strings("supported_protocols", protocols),
				)
				return nil, fmt.Errorf("unknown backend protocol '%v'", backend.Protocol)
			}
			backends := make([]types.ServerClient, 0, len(backend.Servers))
			for _, server := range backend.Servers {
				config.Servers = []string{server}
				client, err = backendInit(backend)
				if err != nil {
					return nil, err
				}
				backends = append(backends, client)
			}

			client, err = broadcast.NewBroadcastGroup(backend.GroupName, backends, pathCache, concurencyLimit, timeouts)
			if err != nil {
				return nil, err
			}
		}
		storeClients = append(storeClients, client)
	}
	return storeClients, nil
}

// NewZipper allows to create new Zipper
func NewZipper(sender func(*types.Stats), config *types.Config, logger *zap.Logger) (*Zipper, error) {
	var err error
	config.Timeouts = sanitizeTimouts(config.Timeouts, defaultTimeouts)

	var searchClients []types.ServerClient
	// Convert old config format to new one
	if config.CarbonSearch.Backend != "" {
		config.CarbonSearchV2.BackendsV2 = types.BackendsV2{
			Backends: []types.BackendV2{{
				GroupName:           config.CarbonSearch.Backend,
				Protocol:            "carbonapi_v2_pb",
				LBMethod:            types.RoundRobinLB,
				Servers:             []string{config.CarbonSearch.Backend},
				Timeouts:            &config.Timeouts,
				ConcurrencyLimit:    &config.ConcurrencyLimitPerServer,
				KeepAliveInterval:   &config.KeepAliveInterval,
				MaxIdleConnsPerHost: &config.MaxIdleConnsPerHost,
				MaxTries:            &config.MaxTries,
				MaxGlobs:            config.MaxGlobs,
			}},
			MaxIdleConnsPerHost:       config.MaxIdleConnsPerHost,
			ConcurrencyLimitPerServer: config.ConcurrencyLimitPerServer,
			Timeouts:                  config.Timeouts,
			KeepAliveInterval:         config.KeepAliveInterval,
			MaxTries:                  config.MaxTries,
			MaxGlobs:                  config.MaxGlobs,
		}
		config.CarbonSearchV2.Prefix = config.CarbonSearch.Prefix
	}

	searchClients, err = createBackendsV2(logger, config.CarbonSearchV2.BackendsV2, config.PathCache)
	if err != nil {
		return nil, err
	}
	prefix := config.CarbonSearchV2.Prefix

	searchBackends, err := broadcast.NewBroadcastGroup("search", searchClients, config.PathCache, config.ConcurrencyLimitPerServer, config.Timeouts)
	if err != nil {
		return nil, err
	}

	storeClients := make([]types.ServerClient, 0)
	// Convert old config format to new one
	if config.Backends != nil && len(config.Backends) != 0 {
		config.BackendsV2 = types.BackendsV2{
			Backends:                  make([]types.BackendV2, 0, len(config.Backends)),
			MaxIdleConnsPerHost:       config.MaxIdleConnsPerHost,
			ConcurrencyLimitPerServer: config.ConcurrencyLimitPerServer,
			Timeouts:                  config.Timeouts,
			KeepAliveInterval:         config.KeepAliveInterval,
			MaxTries:                  config.MaxTries,
			MaxGlobs:                  config.MaxGlobs,
		}
		for _, backend := range config.Backends {
			backend := types.BackendV2{
				GroupName:           config.CarbonSearch.Backend,
				Protocol:            "carbonapi_v2_pb",
				LBMethod:            types.RoundRobinLB,
				Servers:             []string{backend},
				Timeouts:            &config.Timeouts,
				ConcurrencyLimit:    &config.ConcurrencyLimitPerServer,
				KeepAliveInterval:   &config.KeepAliveInterval,
				MaxIdleConnsPerHost: &config.MaxIdleConnsPerHost,
				MaxTries:            &config.MaxTries,
				MaxGlobs:            config.MaxGlobs,
			}
			config.BackendsV2.Backends = append(config.BackendsV2.Backends, backend)
		}
	}

	storeClients, err = createBackendsV2(logger, config.BackendsV2, config.PathCache)
	if err != nil {
		return nil, err
	}
	storeBackends, err := broadcast.NewBroadcastGroup("root", storeClients, config.PathCache, config.ConcurrencyLimitPerServer, config.Timeouts)
	if err != nil {
		return nil, err
	}

	z := &Zipper{
		probeTicker: time.NewTicker(10 * time.Minute),
		ProbeQuit:   make(chan struct{}),
		ProbeForce:  make(chan int),

		sendStats: sender,

		searchCache: config.SearchCache,

		storeBackends:             storeBackends,
		searchBackends:            searchBackends,
		searchPrefix:              prefix,
		searchConfigured:          len(prefix) > 0 && len(searchBackends.Backends()) > 0,
		concurrencyLimitPerServer: config.ConcurrencyLimitPerServer,
		keepAliveInterval:         config.KeepAliveInterval,
		timeout:                   config.Timeouts.Render,
		timeoutConnect:            config.Timeouts.Connect,
		logger:                    logger,
	}

	logger.Info("zipper config",
		zap.Any("config", config),
	)

	go z.probeTlds()

	z.ProbeForce <- 1
	return z, nil
}

func (z *Zipper) doProbe(logger *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), z.timeout)
	defer cancel()

	_, err := z.storeBackends.ProbeTLDs(ctx)
	if err != nil {
		logger.Error("failed to probe tlds",
			zap.Error(err),
		)
	}
}

func (z *Zipper) probeTlds() {
	logger := zapwriter.Logger("probe")
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
func (z Zipper) FetchProtoV3(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, error) {
	var statsSearch *types.Stats
	if z.searchConfigured {
		realRequest := &protov3.MultiFetchRequest{
			Metrics: make([]protov3.FetchRequest, 0, len(request.Metrics)),
		}

		var errors []string
		for _, metric := range request.Metrics {
			if strings.HasPrefix(metric.Name, z.searchPrefix) {
				r := &protov3.MultiGlobRequest{
					Metrics: []string{metric.Name},
				}
				res, stat, err := z.searchBackends.Find(ctx, r)
				statsSearch.Merge(stat)
				if err != nil {
					errors = append(errors, err.Error())
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
		stats.Merge(statsSearch)
	}
	return res, stats, err
}

func (z Zipper) FindProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, error) {
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
	findResponse := &types.ServerFindResponse{
		Response: res,
		Stats:    stats,
		Err:      err,
	}
	if len(searchRequests.Metrics) > 0 {
		resSearch, statsSearch, err := z.searchBackends.Find(ctx, request)
		searchResponse := &types.ServerFindResponse{
			Response: resSearch,
			Stats:    statsSearch,
			Err:      err,
		}
		findResponse.Merge(searchResponse)
	}
	return findResponse.Response, findResponse.Stats, findResponse.Err
}

func (z Zipper) InfoProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.ZipperInfoResponse, *types.Stats, error) {
	realRequest := &protov3.MultiMetricsInfoRequest{Names: make([]string, 0, len(request.Metrics))}
	res, _, err := z.FindProtoV3(ctx, request)
	if err == nil || err == types.ErrNonFatalErrors {
		for _, m := range res.Metrics {
			for _, match := range m.Matches {
				if match.IsLeaf {
					realRequest.Names = append(realRequest.Names, match.Path)
				}
			}
		}
	} else {
		for _, m := range request.Metrics {
			realRequest.Names = append(realRequest.Names, m)
		}
	}

	return z.storeBackends.Info(ctx, realRequest)
}

func (z Zipper) ListProtoV3(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, error) {
	return z.storeBackends.List(ctx)
}
func (z Zipper) StatsProtoV3(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, error) {
	return z.storeBackends.Stats(ctx)
}

// PB3-compatible methods
func (z Zipper) FetchProtoV2(ctx context.Context, query []string, startTime, stopTime int32) (*protov2.MultiFetchResponse, *types.Stats, error) {
	request := &protov3.MultiFetchRequest{}
	for _, q := range query {
		request.Metrics = append(request.Metrics, protov3.FetchRequest{
			Name:      q,
			StartTime: uint32(startTime),
			StopTime:  uint32(stopTime),
		})
	}

	grpcRes, stats, err := z.FetchProtoV3(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	var res protov2.MultiFetchResponse
	for i := range grpcRes.Metrics {
		vals := make([]float64, 0, len(grpcRes.Metrics[i].Values))
		isAbsent := make([]bool, 0, len(grpcRes.Metrics[i].Values))
		for _, v := range grpcRes.Metrics[i].Values {
			if math.IsNaN(v) {
				vals = append(vals, 0)
				isAbsent = append(isAbsent, true)
			} else {
				vals = append(vals, v)
				isAbsent = append(isAbsent, false)
			}
		}
		res.Metrics = append(res.Metrics,
			protov2.FetchResponse{
				Name:      grpcRes.Metrics[i].Name,
				StartTime: int32(grpcRes.Metrics[i].StartTime),
				StopTime:  int32(grpcRes.Metrics[i].StopTime),
				StepTime:  int32(grpcRes.Metrics[i].StepTime),
				Values:    vals,
				IsAbsent:  isAbsent,
			})
	}

	return &res, stats, nil
}

func (z Zipper) FindProtoV2(ctx context.Context, query []string) ([]*protov2.GlobResponse, *types.Stats, error) {
	request := &protov3.MultiGlobRequest{
		Metrics: query,
	}
	grpcReses, stats, err := z.FindProtoV3(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	reses := make([]*protov2.GlobResponse, 0, len(grpcReses.Metrics))
	for _, grpcRes := range grpcReses.Metrics {

		res := &protov2.GlobResponse{
			Name: grpcRes.Name,
		}

		for _, v := range grpcRes.Matches {
			match := protov2.GlobMatch{
				Path:   v.Path,
				IsLeaf: v.IsLeaf,
			}
			res.Matches = append(res.Matches, match)
		}
		reses = append(reses, res)
	}

	return reses, stats, nil
}

func (z Zipper) InfoProtoV2(ctx context.Context, targets []string) (*protov2.ZipperInfoResponse, *types.Stats, error) {
	request := &protov3.MultiGlobRequest{
		Metrics: targets,
	}
	grpcRes, stats, err := z.InfoProtoV3(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	res := &protov2.ZipperInfoResponse{}

	for k, i := range grpcRes.Info {
		for _, v := range i.Metrics {
			rets := make([]protov2.Retention, 0, len(v.Retentions))
			for _, ret := range v.Retentions {
				rets = append(rets, protov2.Retention{
					SecondsPerPoint: int32(ret.SecondsPerPoint),
					NumberOfPoints:  int32(ret.NumberOfPoints),
				})
			}
			i := &protov2.InfoResponse{
				Name:              v.Name,
				AggregationMethod: v.ConsolidationFunc,
				MaxRetention:      int32(v.MaxRetention),
				XFilesFactor:      v.XFilesFactor,
				Retentions:        rets,
			}
			res.Responses = append(res.Responses, protov2.ServerInfoResponse{
				Server: k,
				Info:   i,
			})
		}
	}

	return res, stats, nil
}
func (z Zipper) ListProtoV2(ctx context.Context) (*protov2.ListMetricsResponse, *types.Stats, error) {
	grpcRes, stats, err := z.ListProtoV3(ctx)
	if err != nil {
		return nil, nil, err
	}

	res := &protov2.ListMetricsResponse{
		Metrics: grpcRes.Metrics,
	}
	return res, stats, nil
}
func (z Zipper) StatsProtoV2(ctx context.Context) (*protov2.MetricDetailsResponse, *types.Stats, error) {
	grpcRes, stats, err := z.StatsProtoV3(ctx)
	if err != nil {
		return nil, nil, err
	}

	metrics := make(map[string]*protov2.MetricDetails, len(grpcRes.Metrics))
	for k, v := range grpcRes.Metrics {
		metrics[k] = &protov2.MetricDetails{
			Size_:   v.Size_,
			ModTime: v.ModTime,
			ATime:   v.ATime,
			RdTime:  v.RdTime,
		}
	}

	res := &protov2.MetricDetailsResponse{
		FreeSpace:  grpcRes.FreeSpace,
		TotalSpace: grpcRes.TotalSpace,
		Metrics:    metrics,
	}

	return res, stats, nil
}
