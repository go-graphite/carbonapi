package broadcast

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pathcache"
	"github.com/go-graphite/carbonapi/util"
	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"go.uber.org/zap"
)

type BroadcastGroup struct {
	limiter              *limiter.ServerLimiter
	groupName            string
	timeout              types.Timeouts
	clients              []types.ServerClient
	servers              []string
	maxMetricsPerRequest int

	pathCache pathcache.PathCache
	logger    *zap.Logger
}

func (bg *BroadcastGroup) Children() []types.ServerClient {
	children := make([]types.ServerClient, 0)

	for _, c := range bg.clients {
		children = append(children, c.Children()...)
	}

	return children
}

func NewBroadcastGroup(logger *zap.Logger, groupName string, servers []types.ServerClient, expireDelaySec int32, concurencyLimit int, timeout types.Timeouts) (*BroadcastGroup, *errors.Errors) {
	if len(servers) == 0 {
		return nil, errors.Fatal("no servers specified")
	}
	serverNames := make([]string, 0, len(servers))
	for _, s := range servers {
		serverNames = append(serverNames, s.Name())
	}
	pathCache := pathcache.NewPathCache(expireDelaySec)
	limiter := limiter.NewServerLimiter(serverNames, concurencyLimit)

	return NewBroadcastGroupWithLimiter(logger, groupName, servers, serverNames, pathCache, limiter, timeout)
}

func NewBroadcastGroupWithLimiter(logger *zap.Logger, groupName string, servers []types.ServerClient, serverNames []string, pathCache pathcache.PathCache, limiter *limiter.ServerLimiter, timeout types.Timeouts) (*BroadcastGroup, *errors.Errors) {
	b := &BroadcastGroup{
		timeout:              timeout,
		groupName:            groupName,
		clients:              servers,
		limiter:              limiter,
		servers:              serverNames,
		maxMetricsPerRequest: 100, //TODO remove this hardcoded value

		pathCache: pathCache,
		logger:    logger.With(zap.String("type", "broadcastGroup"), zap.String("groupName", groupName)),
	}

	b.logger.Debug("created broadcast group",
		zap.String("group_name", b.groupName),
		zap.Strings("clients", b.servers),
	)

	return b, nil
}

func (bg BroadcastGroup) Name() string {
	return bg.groupName
}

func (bg BroadcastGroup) Backends() []string {
	return bg.servers
}

func (bg *BroadcastGroup) chooseServers(requests []string) []types.ServerClient {
	var res []types.ServerClient

	for _, request := range requests {
		idx := strings.Index(request, ".")
		if idx > 0 {
			request = request[:idx]
		}
		if clients, ok := bg.pathCache.Get(request); ok && len(clients) > 0 {
			res = append(res, clients...)
		}
	}

	if len(res) != 0 {
		return res
	}
	return bg.clients
}

func (bg *BroadcastGroup) filterServersByTLD(requests []string, clients []types.ServerClient) []types.ServerClient {
	tldClients := make(map[types.ServerClient]bool)
	for _, request := range requests {
		idx := strings.Index(request, ".")
		if idx > 0 {
			request = request[:idx]
		}
		if cacheClients, ok := bg.pathCache.Get(request); ok && len(clients) > 0 {
			for _, cacheClient := range cacheClients {
				tldClients[cacheClient] = true
			}
		}
	}

	var filteredClients []types.ServerClient

	for _, k := range clients {
		if tldClients[k] {
			filteredClients = append(filteredClients, k)
		}
	}

	return filteredClients
}

func (bg BroadcastGroup) MaxMetricsPerRequest() int {
	return bg.maxMetricsPerRequest
}

func fetchRequestToKey(prefix string, request *protov3.MultiFetchRequest) string {
	key := []byte("prefix=" + prefix)
	for _, r := range request.Metrics {
		key = append(key, []byte("&"+r.Name+"&start="+strconv.FormatUint(uint64(r.StartTime), 10)+"&stop="+strconv.FormatUint(uint64(r.StopTime), 10)+"\n")...)
	}

	return string(key)
}

func (bg *BroadcastGroup) doSingleFetch(ctx context.Context, logger *zap.Logger, client types.ServerClient, requests []*protov3.MultiFetchRequest, doneCh chan<- string, resCh chan<- *types.ServerFetchResponse) {
	logger.Debug("waiting for slot",
		zap.Int("maxConns", bg.limiter.Capacity()),
	)
	err := bg.limiter.Enter(ctx, client.Name())
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		resCh <- &types.ServerFetchResponse{
			Err: errors.FromErrNonFatal(err),
		}
		doneCh <- client.Name()
		return
	}
	logger.Debug("got slot")
	defer bg.limiter.Leave(ctx, client.Name())

	for _, req := range requests {
		logger.Debug("sending request",
			zap.String("client_name", client.Name()),
		)
		r := &types.ServerFetchResponse{}
		r.Response, r.Stats, r.Err = client.Fetch(ctx, req)
		resCh <- r
	}
	doneCh <- client.Name()
}

func (bg *BroadcastGroup) SplitRequest(ctx context.Context, request *protov3.MultiFetchRequest) []*protov3.MultiFetchRequest {
	var requests []*protov3.MultiFetchRequest
	maxMetricPerRequest := bg.MaxMetricsPerRequest()

	if maxMetricPerRequest == 0 || bg.groupName != "root" {
		requests = []*protov3.MultiFetchRequest{request}
	} else {
		for _, metric := range request.Metrics {
			f, _, e := bg.Find(ctx, &protov3.MultiGlobRequest{Metrics: []string{metric.Name}})
			if (e != nil && e.HaveFatalErrors && len(e.Errors) > 0) || f == nil || len(f.Metrics) == 0 {
				continue
			}
			newRequest := &protov3.MultiFetchRequest{}

			for _, m := range f.Metrics {
				for _, match := range m.Matches {
					newRequest.Metrics = append(newRequest.Metrics, protov3.FetchRequest{
						Name:            match.Path,
						StartTime:       metric.StartTime,
						StopTime:        metric.StopTime,
						PathExpression:  metric.PathExpression,
						FilterFunctions: metric.FilterFunctions,
					})

					if len(newRequest.Metrics) == maxMetricPerRequest {
						requests = append(requests, newRequest)
						newRequest = &protov3.MultiFetchRequest{}
					}
				}
			}

			if len(newRequest.Metrics) > 0 {
				requests = append(requests, newRequest)
			}
		}
	}

	return requests
}

func (bg *BroadcastGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, *errors.Errors) {
	requestNames := make([]string, 0, len(request.Metrics))
	for i := range request.Metrics {
		requestNames = append(requestNames, request.Metrics[i].Name)
	}
	logger := bg.logger.With(zap.String("type", "fetch"), zap.Strings("request", requestNames))
	logger.Debug("will try to fetch data")

	clients := bg.Children()
	requests := bg.SplitRequest(ctx, request)

	var filteredClients []types.ServerClient
	if bg.groupName == "root" {
		filteredClients = bg.filterServersByTLD(requestNames, clients)
	}
	if len(filteredClients) == 0 {
		filteredClients = clients
	}

	resCh := make(chan *types.ServerFetchResponse, len(filteredClients)*len(requests))
	doneCh := make(chan string, len(filteredClients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Render)
	defer cancel()

	for _, client := range filteredClients {
		go bg.doSingleFetch(ctx, logger, client, requests, doneCh, resCh)
	}

	result := &types.ServerFetchResponse{
		Response: &protov3.MultiFetchResponse{},
		Stats:    &types.Stats{},
		Err:      &errors.Errors{},
	}
	var err errors.Errors
	answeredServers := make(map[string]struct{})
	clientDoneCount := 0
	uuid := util.GetUUID(ctx)

GATHER:
	for {
		if clientDoneCount == len(filteredClients) && len(resCh) == 0 {
			break GATHER
		}
		select {
		case name := <-doneCh:
			clientDoneCount++
			answeredServers[name] = struct{}{}
		case res := <-resCh:
			if res.Err != nil {
				err.Merge(res.Err)
			}
			result.Merge(res, uuid)
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range filteredClients {
				if _, ok := answeredServers[s.Name()]; !ok {
					noAnswer = append(noAnswer, s.Name())
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			err.Add(types.ErrTimeoutExceeded)
			break GATHER
		}
	}

	if len(result.Response.Metrics) == 0 {
		logger.Debug("failed to get any response")

		// TODO(gmagnusson): For this message to be useful, we need to be able
		// to tell if we expected to get a response at all. As is, this message
		// is too spammy.
		return nil, nil, err.Addf("failed to get any response from backend group: %v", bg.groupName)
	}

	logger.Debug("got some responses",
		zap.Int("clients_count", len(filteredClients)),
		zap.Int("response_count", clientDoneCount),
		zap.Bool("have_errors", len(err.Errors) != 0),
		zap.Any("errors", err.Errors),
		zap.Int("response_count", len(result.Response.Metrics)),
	)

	return result.Response, result.Stats, &err
}

// Find request handling

func findRequestToKey(prefix string, request *protov3.MultiGlobRequest) string {
	return "prefix=" + prefix + "&" + strings.Join(request.Metrics, "&")
}

func (bg *BroadcastGroup) doFind(ctx context.Context, logger *zap.Logger, client types.ServerClient, request *protov3.MultiGlobRequest, resCh chan<- *types.ServerFindResponse) {
	logger = logger.With(
		zap.String("group_name", bg.groupName),
		zap.String("client_name", client.Name()),
	)
	logger.Debug("waiting for a slot")

	r := &types.ServerFindResponse{
		Server: client.Name(),
	}

	err := bg.limiter.Enter(ctx, client.Name())
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		r.Err = errors.FromErrNonFatal(types.ErrTimeoutExceeded)
		resCh <- r
		return
	}
	defer bg.limiter.Leave(ctx, client.Name())

	logger.Debug("got a slot")

	r.Response, r.Stats, r.Err = client.Find(ctx, request)
	logger.Debug("fetched response",
		zap.Any("response", r),
	)
	resCh <- r
}

func (bg *BroadcastGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	logger := bg.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))

	resCh := make(chan *types.ServerFindResponse, len(bg.clients))

	logger.Debug("will do query with timeout",
		zap.Float64("timeout", bg.timeout.Find.Seconds()),
	)

	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Render)
	defer cancel()
	ctx = context.Background()

	clients := bg.chooseServers(request.Metrics)
	for _, client := range clients {
		go bg.doFind(ctx, logger, client, request, resCh)
	}

	result := &types.ServerFindResponse{}
	var err errors.Errors
	responseCounts := 0
	answeredServers := make(map[string]struct{})
GATHER:
	for {
		select {
		case r := <-resCh:
			answeredServers[r.Server] = struct{}{}
			responseCounts++
			if r.Err != nil {
				err.Merge(r.Err)
			}
			if result.Response == nil {
				result = r
			} else {
				result.Merge(r)
			}

			if responseCounts == len(clients) {
				break GATHER
			}
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range clients {
				if _, ok := answeredServers[s.Name()]; !ok {
					noAnswer = append(noAnswer, s.Name())
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			err.Add(types.ErrTimeoutExceeded)
			break GATHER
		}
	}
	logger.Debug("got some responses",
		zap.Int("clients_count", len(bg.clients)),
		zap.Int("response_count", responseCounts),
		zap.Bool("have_errors", len(err.Errors) != 0),
		zap.Any("errors", err.Errors),
		zap.Any("response", result.Response),
	)

	if result.Response == nil {
		return &protov3.MultiGlobResponse{}, result.Stats, err.Addf("failed to fetch response from the server %v", bg.groupName)
	}

	return result.Response, result.Stats, &err
}

// Info request handling

func infoRequestToKey(prefix string, request *protov3.MultiMetricsInfoRequest) string {
	return "prefix=" + prefix + "&" + strings.Join(request.Names, "&")
}

func (bg *BroadcastGroup) doInfoRequest(ctx context.Context, logger *zap.Logger, request *protov3.MultiMetricsInfoRequest, client types.ServerClient, resCh chan<- *types.ServerInfoResponse) {
	r := &types.ServerInfoResponse{
		Server: client.Name(),
	}
	logger.Debug("waiting for a slot",
		zap.String("group_name", bg.groupName),
		zap.String("client_name", client.Name()),
	)
	err := bg.limiter.Enter(ctx, client.Name())
	if err != nil {
		logger.Debug("timeout waiting for a slot")
		r.Err = errors.FromErrNonFatal(err)
		resCh <- r
		return
	}
	defer bg.limiter.Leave(ctx, client.Name())

	logger.Debug("got a slot")
	r.Response, r.Stats, r.Err = client.Info(ctx, request)
	resCh <- r
}

func (bg *BroadcastGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	logger := bg.logger.With(zap.String("type", "info"), zap.Strings("request", request.Names))

	resCh := make(chan *types.ServerInfoResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()

	clients := bg.chooseServers(request.Names)
	for _, client := range clients {
		go bg.doInfoRequest(ctx, logger, request, client, resCh)
	}

	result := &types.ServerInfoResponse{}
	var err errors.Errors
	responseCounts := 0
	answeredServers := make(map[string]struct{})
GATHER:
	for {
		select {
		case res := <-resCh:
			answeredServers[res.Server] = struct{}{}
			responseCounts++
			if res.Err != nil {
				err.Merge(res.Err)
			}
			if result.Response == nil {
				result = res
			} else if res.Response != nil {
				for k, v := range res.Response.Info {
					result.Response.Info[k] = v
				}
			}

			if responseCounts == len(clients) {
				break GATHER
			}
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range clients {
				if _, ok := answeredServers[s.Name()]; !ok {
					noAnswer = append(noAnswer, s.Name())
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			err.Add(types.ErrTimeoutExceeded)
			break GATHER
		}
	}
	logger.Debug("got some responses",
		zap.Int("clients_count", len(bg.clients)),
		zap.Int("response_count", responseCounts),
		zap.Bool("have_errors", len(err.Errors) == 0),
	)

	return result.Response, result.Stats, &err
}

func (bg *BroadcastGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (bg *BroadcastGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

type tldResponse struct {
	server types.ServerClient
	tlds   []string
	err    *errors.Errors
}

func doProbe(ctx context.Context, client types.ServerClient, resCh chan<- tldResponse) {
	res, err := client.ProbeTLDs(ctx)

	resCh <- tldResponse{
		server: client,
		tlds:   res,
		err:    err,
	}
}

func (bg *BroadcastGroup) ProbeTLDs(ctx context.Context) ([]string, *errors.Errors) {
	logger := bg.logger.With(zap.String("function", "prober"))

	var tlds []string
	resCh := make(chan tldResponse, len(bg.clients))
	ctx, cancel := context.WithTimeout(context.Background(), bg.timeout.Find)
	defer cancel()

	clients := bg.Children()
	for _, client := range clients {
		go doProbe(ctx, client, resCh)
	}

	responses := 0
	size := uint64(0)
	var err errors.Errors
	answeredServers := make(map[string]struct{})
	cache := make(map[string][]types.ServerClient)
	tldMap := make(map[string]struct{})
GATHER:
	for {
		if responses == len(clients) {
			break GATHER
		}
		select {
		case r := <-resCh:
			answeredServers[r.server.Name()] = struct{}{}
			responses++
			if r.err != nil && len(r.err.Errors) > 0 {
				err.Merge(r.err)
				continue
			}
			for _, tld := range r.tlds {
				tldMap[tld] = struct{}{}
			}
			for _, tld := range r.tlds {
				size += uint64(len(tld))
				cache[tld] = append(cache[tld], r.server)
			}
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range bg.clients {
				if _, ok := answeredServers[s.Name()]; !ok {
					noAnswer = append(noAnswer, s.Name())
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			err.Add(types.ErrTimeoutExceeded)
			break GATHER
		}
	}

	for tld, _ := range tldMap {
		tlds = append(tlds, tld)
	}

	if bg.groupName == "root" {
		for k, v := range cache {
			bg.pathCache.Set(k, v)
		}
	} else {
		logger.Error("Setting path cache in non root bg group. somethings off!")
	}

	return tlds, &err
}
