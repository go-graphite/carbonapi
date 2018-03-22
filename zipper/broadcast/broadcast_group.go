package broadcast

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/pathcache"
	"github.com/go-graphite/carbonzipper/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type BroadcastGroup struct {
	limiter   limiter.ServerLimiter
	groupName string
	timeout   types.Timeouts
	clients   []types.ServerClient
	servers   []string

	pathCache pathcache.PathCache
	logger    *zap.Logger
}

func NewBroadcastGroup(groupName string, servers []types.ServerClient, pathCache pathcache.PathCache, concurencyLimit int, timeout types.Timeouts) (*BroadcastGroup, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers specified")
	}
	serverNames := make([]string, 0, len(servers))
	for _, s := range servers {
		serverNames = append(serverNames, s.Name())
	}
	limiter := limiter.NewServerLimiter(serverNames, concurencyLimit)

	return NewBroadcastGroupWithLimiter(groupName, servers, serverNames, pathCache, limiter, timeout)
}

func NewBroadcastGroupWithLimiter(groupName string, servers []types.ServerClient, serverNames []string, pathCache pathcache.PathCache, limiter limiter.ServerLimiter, timeout types.Timeouts) (*BroadcastGroup, error) {
	return &BroadcastGroup{
		timeout:   timeout,
		groupName: groupName,
		clients:   servers,
		limiter:   limiter,
		servers:   serverNames,

		pathCache: pathCache,
		logger:    zapwriter.Logger("broadcastGroup").With(zap.String("groupName", groupName)),
	}, nil
}

func (bg BroadcastGroup) Name() string {
	return bg.groupName
}

func (c BroadcastGroup) Backends() []string {
	return c.servers
}

func (bg *BroadcastGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, error) {
	logger := bg.logger

	resCh := make(chan *types.ServerFetchResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Render)
	defer cancel()
	for _, client := range bg.clients {
		go func(client types.ServerClient) {
			var r types.ServerFetchResponse
			err := bg.limiter.Enter(ctx, bg.groupName)
			if err != nil {
				return
			}
			r.Response, r.Stats, r.Err = client.Fetch(ctx, request)
			resCh <- &r
			bg.limiter.Leave(ctx, bg.groupName)
		}(client)
	}

	var result types.ServerFetchResponse
	var err error
	responseCounts := 0
GATHER:
	for {
		select {
		case r := <-resCh:
			responseCounts++
			if r.Err != nil {
				err = types.ErrNonFatalErrors
			} else {
				if result.Response == nil {
					result = *r
				} else {
					result.Merge(r)
				}
			}

			if responseCounts == len(bg.clients) {
				break GATHER
			}
		case <-ctx.Done():
			err = types.ErrTimeoutExceeded
			break GATHER
		}
	}

	logger.Debug("got some responses",
		zap.Int("clients_count", len(bg.clients)),
		zap.Int("response_count", responseCounts),
		zap.Bool("have_errors", err != nil),
	)

	return result.Response, result.Stats, err
}

func (bg *BroadcastGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, error) {
	logger := bg.logger
	resCh := make(chan *types.ServerFindResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()
	for _, client := range bg.clients {
		go func(client types.ServerClient) {
			logger.Debug("waiting for a slot",
				zap.String("group_name", bg.groupName),
				zap.String("client_name", client.Name()),
			)
			err := bg.limiter.Enter(ctx, bg.groupName)
			if err != nil {
				logger.Debug("timeout waiting for a slot")
				return
			}

			logger.Debug("got a slot")
			var r types.ServerFindResponse
			r.Response, r.Stats, r.Err = client.Find(ctx, request)
			bg.limiter.Leave(ctx, bg.groupName)
			resCh <- &r
		}(client)
	}

	var result types.ServerFindResponse
	var err error
	responseCounts := 0
GATHER:
	for {
		select {
		case r := <-resCh:
			responseCounts++
			if r.Err == nil {
				if result.Response == nil {
					result = *r
				} else {
					result.Merge(r)
				}
			} else {
				err = types.ErrNonFatalErrors
			}

			if responseCounts == len(bg.clients) {
				break GATHER
			}
		case <-ctx.Done():
			logger.Warn("timeout waiting for more responses")
			err = types.ErrTimeoutExceeded
			break GATHER
		}
	}
	logger.Debug("got some responses",
		zap.Int("clients_count", len(bg.clients)),
		zap.Int("response_count", responseCounts),
		zap.Bool("have_errors", err != nil),
	)

	return result.Response, result.Stats, err
}

func (bg *BroadcastGroup) doInfoRequest(ctx context.Context, logger *zap.Logger, request *protov3.MultiMetricsInfoRequest, client types.ServerClient, resCh chan<- *types.ServerInfoResponse) {
	logger.Debug("waiting for a slot",
		zap.String("group_name", bg.groupName),
		zap.String("client_name", client.Name()),
	)
	err := bg.limiter.Enter(ctx, bg.groupName)
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		return
	}

	r := &types.ServerInfoResponse{}
	logger.Debug("got a slot")
	r.Response, r.Stats, r.Err = client.Info(ctx, request)
	bg.limiter.Leave(ctx, bg.groupName)
	logger.Debug("maybe got response",
		zap.Any("r", r),
	)
	resCh <- r
}

func (bg *BroadcastGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, error) {
	logger := bg.logger
	resCh := make(chan *types.ServerInfoResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()
	for _, client := range bg.clients {
		go bg.doInfoRequest(ctx, logger, request, client, resCh)
	}

	var result types.ServerInfoResponse
	var err error
	responseCounts := 0
GATHER:
	for {
		select {
		case r := <-resCh:
			logger.Debug("got response",
				zap.Any("r", r),
			)
			responseCounts++
			if r.Err == nil {
				if result.Response == nil {
					result = *r
				} else {
					for k, v := range r.Response.Info {
						result.Response.Info[k] = v
					}
				}
			} else {
				err = types.ErrNonFatalErrors
			}

			if responseCounts == len(bg.clients) {
				break GATHER
			}
		case <-ctx.Done():
			logger.Warn("timeout waiting for more responses")
			err = types.ErrTimeoutExceeded
			break GATHER
		}
	}
	logger.Debug("got some responses",
		zap.Int("clients_count", len(bg.clients)),
		zap.Int("response_count", responseCounts),
		zap.Bool("have_errors", err != nil),
	)

	return result.Response, result.Stats, err
}

func (bg *BroadcastGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (bg *BroadcastGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, error) {
	return nil, nil, types.ErrNotImplementedYet
}

type tldResponse struct {
	server string
	tlds   []string
	err    error
}

func doProbe(ctx context.Context, client types.ServerClient, resCh chan<- tldResponse) {
	name := client.Name()

	res, err := client.ProbeTLDs(ctx)
	if err != nil {
		resCh <- tldResponse{
			server: name,
			err:    err,
		}
		return
	}

	resCh <- tldResponse{
		server: name,
		tlds:   res,
	}
}

func (bg *BroadcastGroup) ProbeTLDs(ctx context.Context) ([]string, error) {
	logger := bg.logger
	var tlds []string
	cache := make(map[string][]string)
	resCh := make(chan tldResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()

	for _, client := range bg.clients {
		go doProbe(ctx, client, resCh)
	}

	responses := 0
GATHER:
	for {
		select {
		case r := <-resCh:
			responses++
			if r.err != nil {
				logger.Error("failed to probe tld",
					zap.String("name", r.server),
					zap.Error(r.err),
				)
				continue
			}
			tlds = append(tlds, r.tlds...)
			for _, tld := range r.tlds {
				cache[tld] = append(cache[tld], r.server)
			}

			if responses == len(bg.clients) {
				break GATHER
			}
		case <-ctx.Done():
			break GATHER
		}
	}
	cancel()

	logger.Debug("TLD Probe",
		zap.Any("cache", cache),
	)

	for k, v := range cache {
		bg.pathCache.Set(k, v)
	}

	return tlds, nil
}
