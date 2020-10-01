package broadcast

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ansel1/merry"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pathcache"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"go.uber.org/zap"
)

type BroadcastGroup struct {
	limiter                   limiter.ServerLimiter
	groupName                 string
	timeout                   types.Timeouts
	backends                  []types.BackendServer
	servers                   []string
	maxMetricsPerRequest      int
	doMultipleRequestsIfSplit bool
	tldCacheDisabled          bool

	fetcher   types.Fetcher
	pathCache pathcache.PathCache
	logger    *zap.Logger
}

func (bg *BroadcastGroup) Children() []types.BackendServer {
	return bg.backends
}

func (bg *BroadcastGroup) SetDoMultipleRequestIfSplit(v bool) {
	bg.doMultipleRequestsIfSplit = v
	if v {
		bg.fetcher = bg.doMultiFetch
	} else {
		bg.fetcher = bg.doSingleFetch
	}
}

func NewBroadcastGroup(logger *zap.Logger, groupName string, doMultipleRequestsIfSplit bool, servers []types.BackendServer, expireDelaySec int32, concurrencyLimit, maxBatchSize int, timeout types.Timeouts, tldCacheDisabled bool) (*BroadcastGroup, merry.Error) {
	if len(servers) == 0 {
		return nil, types.ErrNoServersSpecified
	}
	serverNames := make([]string, 0, len(servers))
	for _, s := range servers {
		serverNames = append(serverNames, s.Name())
	}
	pathCache := pathcache.NewPathCache(expireDelaySec)
	l := limiter.NewServerLimiter(serverNames, concurrencyLimit)

	return NewBroadcastGroupWithLimiter(logger, groupName, doMultipleRequestsIfSplit, servers, serverNames, maxBatchSize, pathCache, l, timeout, tldCacheDisabled)
}

func NewBroadcastGroupWithLimiter(logger *zap.Logger, groupName string, doMultipleRequestsIfSplit bool, servers []types.BackendServer, serverNames []string, maxBatchSize int, pathCache pathcache.PathCache, limiter limiter.ServerLimiter, timeout types.Timeouts, tldCacheDisabled bool) (*BroadcastGroup, merry.Error) {
	b := &BroadcastGroup{
		timeout:              timeout,
		groupName:            groupName,
		backends:             servers,
		limiter:              limiter,
		servers:              serverNames,
		maxMetricsPerRequest: maxBatchSize,
		tldCacheDisabled:     tldCacheDisabled,

		pathCache: pathCache,
		logger:    logger.With(zap.String("type", "broadcastGroup"), zap.String("groupName", groupName)),
	}

	b.SetDoMultipleRequestIfSplit(doMultipleRequestsIfSplit)

	b.logger.Debug("created broadcast group",
		zap.String("group_name", b.groupName),
		zap.Strings("backends", b.servers),
		zap.Int("max_batch_size", maxBatchSize),
		zap.Bool("do_single_request_at_a_time", b.doMultipleRequestsIfSplit),
	)

	return b, nil
}

func (bg BroadcastGroup) Name() string {
	return bg.groupName
}

func (bg BroadcastGroup) Backends() []string {
	return bg.servers
}

func (bg *BroadcastGroup) filterServersByTLD(requests []string, backends []types.BackendServer) []types.BackendServer {
	// do not check TLDs if internal routing cache is disabled
	if bg.tldCacheDisabled {
		return backends
	}

	tldBackends := make(map[types.BackendServer]bool)
	for _, request := range requests {
		// TODO(Civil): Tags: improve logic
		if strings.HasPrefix(request, "seriesByTag") {
			return backends
		}
		idx := strings.Index(request, ".")
		if idx > 0 {
			request = request[:idx]
		}
		if cachedBackends, ok := bg.pathCache.Get(request); ok && len(backends) > 0 {
			for _, cachedBackend := range cachedBackends {
				tldBackends[cachedBackend] = true
			}
		}
	}

	var filteredBackends []types.BackendServer
	for _, k := range backends {
		if tldBackends[k] {
			filteredBackends = append(filteredBackends, k)
		}
	}

	if len(filteredBackends) == 0 {
		return backends
	}

	return filteredBackends
}

func (bg BroadcastGroup) MaxMetricsPerRequest() int {
	return bg.maxMetricsPerRequest
}

func (bg *BroadcastGroup) doMultiFetch(ctx context.Context, logger *zap.Logger, backend types.BackendServer, reqs interface{}, resCh chan types.ServerFetcherResponse) {
	requests, ok := reqs.([]*protov3.MultiFetchRequest)
	if !ok {
		logger.Fatal("unhandled error in doMultiFetch",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", reqs)),
			zap.String("expected_type", fmt.Sprintf("%T", requests)),
		)
	}

	for _, req := range requests {
		go func(req *protov3.MultiFetchRequest) {
			logger = logger.With(zap.String("backend_name", backend.Name()))
			logger.Debug("waiting for slot",
				zap.Int("max_connections", bg.limiter.Capacity()),
			)

			response := types.NewServerFetchResponse()
			response.Server = backend.Name()

			if err := bg.limiter.Enter(ctx, backend.Name()); err != nil {
				logger.Debug("timeout waiting for a slot")
				resCh <- response.NonFatalError(merry.Prepend(err, "timeout waiting for slot"))
				return
			}

			logger.Debug("got slot")
			defer bg.limiter.Leave(ctx, backend.Name())

			// uuid := util.GetUUID(ctx)
			var err merry.Error
			logger.Debug("sending request")
			response.Response, response.Stats, err = backend.Fetch(ctx, req)
			response.AddError(err)
			logger.Debug("got response")

			resCh <- response
		}(req)
	}

}

func (bg *BroadcastGroup) doSingleFetch(ctx context.Context, logger *zap.Logger, backend types.BackendServer, reqs interface{}, resCh chan types.ServerFetcherResponse) {
	requests, ok := reqs.([]*protov3.MultiFetchRequest)
	if !ok {
		logger.Fatal("unhandled error in doSingleFetch",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", reqs)),
			zap.String("expected_type", fmt.Sprintf("%T", requests)),
		)
	}
	logger = logger.With(zap.String("backend_name", backend.Name()))
	logger.Debug("waiting for slot",
		zap.Int("max_connections", bg.limiter.Capacity()),
	)

	response := types.NewServerFetchResponse()
	response.Server = backend.Name()

	if err := bg.limiter.Enter(ctx, backend.Name()); err != nil {
		logger.Debug("timeout waiting for a slot")
		resCh <- response.NonFatalError(merry.Prepend(err, "timeout waiting for slot"))
		return
	}

	logger.Debug("got slot")
	defer bg.limiter.Leave(ctx, backend.Name())

	// uuid := util.GetUUID(ctx)
	var err merry.Error
	for _, req := range requests {
		logger.Debug("sending request")
		r := types.NewServerFetchResponse()
		r.Response, r.Stats, err = backend.Fetch(ctx, req)
		r.AddError(err)
		logger.Debug("got response")
		_ = response.Merge(r)
	}

	resCh <- response
}

func (bg *BroadcastGroup) splitRequest(ctx context.Context, request *protov3.MultiFetchRequest) ([]*protov3.MultiFetchRequest, merry.Error) {
	if bg.MaxMetricsPerRequest() == 0 {
		return []*protov3.MultiFetchRequest{request}, nil
	}

	var requests []*protov3.MultiFetchRequest
	var err merry.Error
	for _, metric := range request.Metrics {
		newRequest := &protov3.MultiFetchRequest{}
		// TODO(Civil): Tags: improve logic
		if strings.HasPrefix(metric.Name, "seriesByTag") {
			newRequest.Metrics = append(newRequest.Metrics, protov3.FetchRequest{
				Name:            metric.PathExpression,
				StartTime:       metric.StartTime,
				StopTime:        metric.StopTime,
				PathExpression:  metric.PathExpression,
				FilterFunctions: metric.FilterFunctions,
			})
			requests = append(requests, newRequest)
			continue
		}

		// Do not send Find requests if we have neither globs in the request nor metric expansions
		if !strings.ContainsAny(metric.Name, "*{") {
			newRequest.Metrics = append(newRequest.Metrics, protov3.FetchRequest{
				Name:            metric.PathExpression,
				StartTime:       metric.StartTime,
				StopTime:        metric.StopTime,
				PathExpression:  metric.PathExpression,
				FilterFunctions: metric.FilterFunctions,
			})
			requests = append(requests, newRequest)
			continue
		}

		f, _, e := bg.Find(ctx, &protov3.MultiGlobRequest{Metrics: []string{metric.Name}})
		if e != nil || f == nil || len(f.Metrics) == 0 {
			if e == nil {
				e = merry.Errorf("no result fetched")
				if f == nil {
					e = e.WithCause(types.ErrUnmarshalFailed)
				} else {
					e = e.WithCause(types.ErrNoMetricsFetched)
				}
			}
			err = e

			errStr := ""
			if e.Cause() != nil {
				errStr = e.Cause().Error()
			} else {
				// e != nil, but len(f.Metrics) == 0 or f == nil, then Cause could be nil
				errStr = e.Error()
			}

			if ce := bg.logger.Check(zap.DebugLevel, "find request failed when resolving globs (verbose)"); ce != nil {
				ce.Write(
					zap.String("metric_name", metric.Name),
					zap.String("error", errStr),
					zap.Any("stack", e),
				)
			} else {
				bg.logger.Warn("find request failed when resolving globs",
					zap.String("metric_name", metric.Name),
					zap.String("error", errStr),
				)
			}

			if f == nil {
				continue
			}
		}

		for _, m := range f.Metrics {
			for _, match := range m.Matches {
				newRequest.Metrics = append(newRequest.Metrics, protov3.FetchRequest{
					Name:            match.Path,
					StartTime:       metric.StartTime,
					StopTime:        metric.StopTime,
					PathExpression:  metric.PathExpression,
					FilterFunctions: metric.FilterFunctions,
				})

				if len(newRequest.Metrics) == bg.MaxMetricsPerRequest() {
					requests = append(requests, newRequest)
					newRequest = &protov3.MultiFetchRequest{}
				}
			}
		}

		if len(newRequest.Metrics) > 0 {
			requests = append(requests, newRequest)
		}
	}

	return requests, err
}

func (bg *BroadcastGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	requestNames := make([]string, 0, len(request.Metrics))
	for i := range request.Metrics {
		requestNames = append(requestNames, request.Metrics[i].Name)
	}
	logger := bg.logger.With(zap.String("type", "fetch"), zap.Strings("request", requestNames))
	logger.Debug("will try to fetch data")

	backends := bg.filterServersByTLD(requestNames, bg.Children())
	requests, err := bg.splitRequest(ctx, request)
	zipperRequests, totalMetricsCount := getFetchRequestMetricStats(requests, bg, backends)

	result := types.NewServerFetchResponse()
	result.Stats.ZipperRequests = int64(zipperRequests)
	result.Stats.TotalMetricsCount = int64(totalMetricsCount)
	if len(requests) == 0 || err != nil {
		return result.Response, result.Stats, err
	}

	ctxNew, cancel := context.WithTimeout(ctx, bg.timeout.Render)
	defer cancel()

	resultNew, responseCount := types.DoRequest(ctxNew, logger, backends, result, requests, bg.fetcher)

	result, ok := resultNew.Self().(*types.ServerFetchResponse)
	if !ok {
		logger.Fatal("unhandled error in Fetch",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", resultNew.Self())),
			zap.String("expected_type", fmt.Sprintf("%T", result)),
		)
	}

	if len(result.Response.Metrics) == 0 {
		nonNotFoundErrors := types.ReturnNonNotFoundError(result.Err)
		if nonNotFoundErrors != nil {
			err := types.ErrFailedToFetch.WithHTTPCode(500)
			for _, e := range nonNotFoundErrors {
				err = err.WithCause(e)
			}
			logger.Debug("non-404 errors while fetching data from backends",
				zap.Any("errors", result.Err),
			)
			return nil, result.Stats, err
		}
		return nil, result.Stats, types.ErrNotFound.WithHTTPCode(404)
	}

	// Recalculate metrics start/step/stop parameters to avoid upstream misbehavior
	for i, metric := range result.Response.Metrics {
		result.Response.Metrics[i].StopTime = metric.StartTime + int64(len(metric.Values))*metric.StepTime
	}

	logger.Debug("got some fetch responses",
		zap.Int("backends_count", len(backends)),
		zap.Int("response_count", responseCount),
		zap.Bool("have_errors", len(result.Err) != 0),
		zap.Any("errors", result.Err),
		zap.Int("response_count", len(result.Response.Metrics)),
	)

	if result.Err != nil && len(result.Err) > 0 {
		err = types.ErrNonFatalErrors
		for _, e := range result.Err {
			err = err.WithCause(e)
		}
	}

	return result.Response, result.Stats, err
}

func getFetchRequestMetricStats(requests []*protov3.MultiFetchRequest, bg *BroadcastGroup, backends []types.BackendServer) (int, int) {
	var totalMetricsCount int
	var zipperRequests int
	if bg.MaxMetricsPerRequest() > 0 {
		zipperRequests = 1
	}
	if len(requests) > 0 {
		totalMetricsCount = (len(requests)-1)*bg.MaxMetricsPerRequest() + len(requests[len(requests)-1].Metrics)
		zipperRequests += len(requests) * len(backends)
	}
	return zipperRequests, totalMetricsCount
}

// Find request handling
func (bg *BroadcastGroup) doFind(ctx context.Context, logger *zap.Logger, backend types.BackendServer, reqs interface{}, resCh chan types.ServerFetcherResponse) {
	request, ok := reqs.(*protov3.MultiGlobRequest)
	if !ok {
		logger.Fatal("unhandled error",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", reqs)),
			zap.String("expected_type", fmt.Sprintf("%T", request)),
		)
	}
	logger = logger.With(
		zap.String("group_name", bg.groupName),
		zap.String("backend_name", backend.Name()),
	)
	logger.Debug("waiting for a slot")

	r := types.NewServerFindResponse()
	r.Server = backend.Name()

	if err := bg.limiter.Enter(ctx, backend.Name()); err != nil {
		logger.Debug("timeout waiting for a slot")
		r.AddError(merry.Prepend(err, "timeout waiting for slot"))
		resCh <- r
		return
	}

	logger.Debug("got slot")
	defer bg.limiter.Leave(ctx, backend.Name())

	var err merry.Error
	r.Response, r.Stats, err = backend.Find(ctx, request)
	r.AddError(err)
	logger.Debug("fetched response",
		zap.Any("response", r),
	)
	resCh <- r
}

func (bg *BroadcastGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := bg.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))

	backends := bg.Children()

	logger.Debug("will do query with timeout",
		zap.Any("backends", backends),
		zap.Float64("timeout", bg.timeout.Find.Seconds()),
	)

	ctxNew, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()

	result := types.NewServerFindResponse()
	result.Server = bg.Name()
	result.Stats.ZipperRequests = int64(len(backends))
	resultNew, responseCount := types.DoRequest(ctxNew, logger, backends, result, request, bg.doFind)

	result, ok := resultNew.Self().(*types.ServerFindResponse)
	if !ok {
		logger.Fatal("unhandled error in Find",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", resultNew.Self())),
			zap.String("expected_type", fmt.Sprintf("%T", result)),
		)
	}

	if len(result.Response.Metrics) == 0 {
		nonNotFoundErrors := types.ReturnNonNotFoundError(result.Err)
		if nonNotFoundErrors != nil {
			err := types.ErrFailedToFetch.WithHTTPCode(500)
			for _, e := range nonNotFoundErrors {
				err = err.WithCause(e)
			}
			logger.Debug("non-404 errors while fetching data from backends",
				zap.Any("errors", result.Err),
			)
			return &protov3.MultiGlobResponse{}, result.Stats, err
		}

		return &protov3.MultiGlobResponse{}, result.Stats, types.ErrNotFound.WithHTTPCode(404)
	}
	result.Stats.TotalMetricsCount = 0
	for _, x := range result.Response.Metrics {
		result.Stats.TotalMetricsCount += int64(len(x.Matches))
	}
	logger.Debug("got some find responses",
		zap.Int("backends_count", len(backends)),
		zap.Int("response_count", responseCount),
		zap.Bool("have_errors", len(result.Err) != 0),
		zap.Any("errors", result.Err),
		zap.Any("response", result.Response),
	)

	var err merry.Error
	if result.Err != nil {
		err = types.ErrNonFatalErrors
		for _, e := range result.Err {
			err = err.WithCause(e)
		}
	}

	return result.Response, result.Stats, err
}

// Info request handling
func (bg *BroadcastGroup) doInfoRequest(ctx context.Context, logger *zap.Logger, backend types.BackendServer, reqs interface{}, resCh chan types.ServerFetcherResponse) {
	logger = logger.With(
		zap.String("group_name", bg.groupName),
		zap.String("backend_name", backend.Name()),
	)
	request, ok := reqs.(*protov3.MultiMetricsInfoRequest)
	if !ok {
		logger.Fatal("unhandled error",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", reqs)),
			zap.String("expected_type", fmt.Sprintf("%T", request)),
		)
	}
	r := &types.ServerInfoResponse{
		Server: backend.Name(),
	}

	if err := bg.limiter.Enter(ctx, backend.Name()); err != nil {
		logger.Debug("timeout waiting for a slot")
		r.AddError(merry.Prepend(err, "timeout waiting for slot"))
		resCh <- r
		return
	}
	defer bg.limiter.Leave(ctx, backend.Name())

	logger.Debug("got a slot")
	var err merry.Error
	r.Response, r.Stats, err = backend.Info(ctx, request)
	r.AddError(err)
	resCh <- r
}

func (bg *BroadcastGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	logger := bg.logger.With(zap.String("type", "info"), zap.Strings("request", request.Names))

	ctxNew, cancel := context.WithTimeout(ctx, bg.timeout.Render)
	defer cancel()
	backends := bg.Children()
	result := types.NewServerInfoResponse()
	result.Server = bg.Name()
	result.Stats.ZipperRequests = int64(len(backends))

	resultNew, responseCount := types.DoRequest(ctxNew, logger, backends, result, request, bg.doInfoRequest)

	result, ok := resultNew.Self().(*types.ServerInfoResponse)
	if !ok {
		logger.Fatal("unhandled error in Find",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", resultNew.Self())),
			zap.String("expected_type", fmt.Sprintf("%T", result)),
		)
	}

	logger.Debug("got some responses",
		zap.Int("backends_count", len(backends)),
		zap.Int("response_count", responseCount),
		zap.Bool("have_errors", len(result.Err) != 0),
	)

	var err merry.Error
	if result.Err != nil {
		err = types.ErrNonFatalErrors
		for _, e := range result.Err {
			err = err.WithCause(e)
		}
	}

	return result.Response, result.Stats, err
}

func (bg *BroadcastGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (bg *BroadcastGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}

type tagQuery struct {
	Query  string
	Limit  int64
	IsName bool
}

// Info request handling
func (bg *BroadcastGroup) doTagRequest(ctx context.Context, logger *zap.Logger, backend types.BackendServer, reqs interface{}, resCh chan types.ServerFetcherResponse) {
	request, ok := reqs.(tagQuery)
	logger = logger.With(
		zap.String("group_name", bg.groupName),
		zap.String("backend_name", backend.Name()),
	)
	if !ok {
		logger.Fatal("unhandled error",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", reqs)),
			zap.String("expected_type", fmt.Sprintf("%T", request)),
		)
	}
	r := &types.ServerTagResponse{
		Server:   backend.Name(),
		Response: []string{},
	}

	logger.Debug("waiting for a slot")

	if err := bg.limiter.Enter(ctx, backend.Name()); err != nil {
		logger.Debug("timeout waiting for a slot")
		r.AddError(merry.Prepend(err, "timeout waiting for slot"))
		resCh <- r
		return
	}
	defer bg.limiter.Leave(ctx, backend.Name())

	logger.Debug("got a slot")
	var err merry.Error
	if request.IsName {
		r.Response, err = backend.TagNames(ctx, request.Query, request.Limit)
	} else {
		r.Response, err = backend.TagValues(ctx, request.Query, request.Limit)
	}

	if err != nil {
		r.AddError(err)
	}

	if r.Response == nil {
		r.Response = []string{}
	}
	resCh <- r
}

func (bg *BroadcastGroup) tagEverything(ctx context.Context, isTagName bool, query string, limit int64) ([]string, merry.Error) {
	logger := bg.logger.With(zap.String("query", query))
	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
	}

	request := tagQuery{
		Query:  query,
		Limit:  limit,
		IsName: isTagName,
	}

	ctxNew, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()

	backends := bg.Children()
	result := types.NewServerTagResponse()
	result.Server = bg.Name()

	resultNew, responseCount := types.DoRequest(ctxNew, logger, backends, result, request, bg.doTagRequest)

	result, ok := resultNew.Self().(*types.ServerTagResponse)
	if !ok {
		logger.Fatal("unhandled error in Find",
			zap.Stack("stack"),
			zap.String("got_type", fmt.Sprintf("%T", resultNew.Self())),
			zap.String("expected_type", fmt.Sprintf("%T", result)),
		)
	}

	if limit != -1 && int64(len(result.Response)) > limit {
		sort.Strings(result.Response)
		result.Response = result.Response[:limit-1]
	}

	logger.Debug("got some responses",
		zap.Int("backends_count", len(backends)),
		zap.Int("response_count", responseCount),
		zap.Bool("have_errors", len(result.Err) != 0),
	)

	var err merry.Error
	if result.Err != nil {
		err = types.ErrNonFatalErrors
		for _, e := range result.Err {
			err = err.WithCause(e)
		}
	}

	return result.Response, err
}

func (bg *BroadcastGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return bg.tagEverything(ctx, true, query, limit)
}

func (bg *BroadcastGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return bg.tagEverything(ctx, false, query, limit)
}

type tldResponse struct {
	server types.BackendServer
	tlds   []string
	err    merry.Error
}

func doProbe(ctx context.Context, backend types.BackendServer, resCh chan<- tldResponse) {
	res, err := backend.ProbeTLDs(ctx)

	resCh <- tldResponse{
		server: backend,
		tlds:   res,
		err:    err,
	}
}

func (bg *BroadcastGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
	logger := bg.logger.With(zap.String("function", "prober"))

	ctx, cancel := context.WithTimeout(ctx, bg.timeout.Find)
	defer cancel()

	backends := bg.Children()
	resCh := make(chan tldResponse, len(backends))
	for _, backend := range backends {
		go doProbe(ctx, backend, resCh)
	}

	responses := 0
	var errs []merry.Error
	answeredServers := make(map[string]struct{})
	cache := make(map[string][]types.BackendServer)
	tldSet := make(map[string]struct{})

GATHER:
	for {
		if responses == len(backends) {
			break GATHER
		}

		select {
		case r := <-resCh:
			answeredServers[r.server.Name()] = struct{}{}
			responses++
			if r.err != nil {
				errs = append(errs, r.err)
				continue
			}
			for _, tld := range r.tlds {
				tldSet[tld] = struct{}{}
				cache[tld] = append(cache[tld], r.server)
			}

		case <-ctx.Done():
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", types.NoAnswerBackends(backends, answeredServers)),
			)
			errs = append(errs, types.ErrTimeoutExceeded)
			break GATHER
		}
	}

	var tlds []string
	for tld := range tldSet {
		tlds = append(tlds, tld)
	}

	for k, v := range cache {
		bg.pathCache.Set(k, v)
	}

	var err merry.Error
	if errs != nil {
		err = types.ErrNonFatalErrors
		for _, e := range errs {
			err = err.WithCause(e)
		}
	}

	return tlds, err
}
