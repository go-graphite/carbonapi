package irondb

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/circonus-labs/gosnowth"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"
)

func init() {
	aliases := []string{"irondb", "snowthd"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// IronDBGroup is a protocol group that can query IronDB servers
type IronDBGroup struct {
	types.BackendServer

	groupName string
	servers   []string
	protocol  string

	client *gosnowth.SnowthClient

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	accountID       int64
	graphite_rollup int64
	graphite_prefix string
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "irondb"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

	logger.Warn("support for this backend protocol is experimental, use with caution")

	// initializing config with list of servers from upstream
	cfg, err := gosnowth.NewConfig(config.Servers...)
	if err != nil {
		logger.Fatal("failed to create snowth configuration",
			zap.Error(err))
	}

	// enabling discovery.
	cfg.SetDiscover(true)

	// parse backend options
	accountID := int64(1)
	accountID_opt, ok := config.BackendOptions["irondb_account_id"]
	if ok {
		accountID_tmp, ok := accountID_opt.(int)
		if !ok {
			logger.Fatal("failed to parse irondb_account_id",
				zap.String("type_parsed", fmt.Sprintf("%T", accountID_opt)),
				zap.String("type_expected", "int"),
			)
		}
		accountID = int64(accountID_tmp)
	}

	maxTries := int64(*config.MaxTries)
	retries := int64(0)
	retries_opt, ok := config.BackendOptions["irondb_retries"]
	if ok {
		retries_tmp, ok := retries_opt.(int)
		if !ok {
			logger.Fatal("failed to parse irondb_retries",
				zap.String("type_parsed", fmt.Sprintf("%T", retries_opt)),
				zap.String("type_expected", "int"),
			)
		}
		retries = int64(retries_tmp)
	}
	if maxTries > retries {
		retries = maxTries
	}
	cfg.SetRetries(retries)

	connectRetries := int64(-1)
	connectRetries_opt, ok := config.BackendOptions["irondb_connect_retries"]
	if ok {
		connectRetries_tmp, ok := connectRetries_opt.(int)
		if !ok {
			logger.Fatal("failed to parse irondb_connect_retries",
				zap.String("type_parsed", fmt.Sprintf("%T", connectRetries_opt)),
				zap.String("type_expected", "int"),
			)
		}
		connectRetries = int64(connectRetries_tmp)
	}
	cfg.SetConnectRetries(connectRetries)

	dialTimeout := 500 * time.Millisecond
	dialTimeout_opt, ok := config.BackendOptions["irondb_dial_timeout"]
	if ok {
		dialTimeout_str, ok := dialTimeout_opt.(string)
		if ok {
			interval, err := time.ParseDuration(dialTimeout_str)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_dial_timeout"),
					zap.String("option_value", dialTimeout_str),
					zap.Errors("errors", []error{err}),
				)
			}
			dialTimeout = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_dial_timeout"),
				zap.Any("option_value", dialTimeout_str),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.SetDialTimeout(dialTimeout)

	irondbTimeout := 10 * time.Second
	irondbTimeout_opt, ok := config.BackendOptions["irondb_timeout"]
	if ok {
		irondbTimeout_str, ok := irondbTimeout_opt.(string)
		if ok {
			interval, err := time.ParseDuration(irondbTimeout_str)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_timeout"),
					zap.String("option_value", irondbTimeout_str),
					zap.Errors("errors", []error{err}),
				)
			}
			irondbTimeout = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_timeout"),
				zap.Any("option_value", irondbTimeout_str),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.SetTimeout(irondbTimeout)

	watchInterval := 30 * time.Second
	watchInterval_opt, ok := config.BackendOptions["irondb_watch_interval"]
	if ok {
		watchInterval_str, ok := watchInterval_opt.(string)
		if ok {
			interval, err := time.ParseDuration(watchInterval_str)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_watch_interval"),
					zap.String("option_value", watchInterval_str),
					zap.Errors("errors", []error{err}),
				)
			}
			watchInterval = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_watch_interval"),
				zap.Any("option_value", watchInterval_str),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.SetWatchInterval(watchInterval)

	graphite_rollup := int64(60)
	graphite_rollup_opt, ok := config.BackendOptions["irondb_graphite_rollup"]
	if ok {
		graphite_rollup_tmp, ok := graphite_rollup_opt.(int)
		if !ok {
			logger.Fatal("failed to parse irondb_graphite_rollup",
				zap.String("type_parsed", fmt.Sprintf("%T", graphite_rollup_opt)),
				zap.String("type_expected", "int"),
			)
		}
		graphite_rollup = int64(graphite_rollup_tmp)
	}

	graphite_prefix := ""
	graphite_prefix_opt, ok := config.BackendOptions["irondb_graphite_prefix"]
	if ok {
		graphite_prefix_tmp, ok := graphite_prefix_opt.(string)
		if !ok {
			logger.Fatal("failed to parse irondb_graphite_prefix",
				zap.String("type_parsed", fmt.Sprintf("%T", graphite_prefix_opt)),
				zap.String("type_expected", "string"),
			)
		}
		graphite_prefix = string(graphite_prefix_tmp)
	}

	snowthClient, err := gosnowth.NewClient(cfg)
	if err != nil {
		logger.Fatal("failed to create snowth client",
			zap.Error(err))
	}

	c := &IronDBGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: *config.MaxBatchSize,
		client:               snowthClient,
		accountID:            accountID,
		graphite_rollup:      graphite_rollup,
		graphite_prefix:      graphite_prefix,
		limiter:              limiter,
		logger:               logger,
	}

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

func (c *IronDBGroup) Children() []types.BackendServer {
	return []types.BackendServer{c}
}

func (c IronDBGroup) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c IronDBGroup) Name() string {
	return c.groupName
}

func (c IronDBGroup) Backends() []string {
	return c.servers
}

func (c *IronDBGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error

	start := request.Metrics[0].StartTime
	stop := request.Metrics[0].StopTime
	step := c.graphite_rollup
	count := int64((stop-start)/step) + 1
	max_count := request.Metrics[0].MaxDataPoints
	if count > max_count && max_count > 0 {
		count = max_count
		step = adjustStep(start, stop, max_count, step)
	}
	if count <= 0 {
		logger.Fatal("stop time should be less then start",
			zap.Int64("start", start),
			zap.Int64("stop", stop),
		)
	}
	for pathExpr, targets := range pathExprToTargets {
		for _, target := range targets {
			logger.Debug("got some target to query",
				zap.Any("pathExpr", pathExpr),
				zap.String("target", target),
				zap.Int64("start", start),
				zap.Int64("stop", stop),
				zap.Int64("count", count),
				zap.Int64("period", step),
			)

			var query string
			if strings.HasPrefix(target, "seriesByTag") {
				query = target[len("seriesByTag(") : len(target)-1]
				// d-oh, fetch_multi graphite compatible API not working with tags
				// fallback to IRONdb Fetch API
				query = graphiteExprListToIronDBTagQuery(strings.Split(query, ","))
				findTagOptions := &gosnowth.FindTagsOptions{
					// start and stop according to request
					Start:     time.Unix(start, 0),
					End:       time.Unix(stop, 0),
					Activity:  0,
					Latest:    0,
					CountOnly: 0,
					Limit:     -1,
				}
				logger.Debug("send tag find result to irondb",
					zap.String("query", query),
					zap.Any("findTagOptions", findTagOptions),
				)
				stats.FindRequests++
				tag_metrics, err := c.client.FindTags(c.accountID, query, findTagOptions)
				if err != nil {
					stats.FindErrors++
					if merry.Is(err, types.ErrTimeoutExceeded) {
						stats.Timeouts++
						stats.FindTimeouts++
					}
					if e == nil {
						e = merry.Wrap(err).WithValue("query", query)
					} else {
						e = e.WithCause(err)
					}
					continue
				}
				logger.Debug("got tag find result from irondb",
					zap.String("query", query),
					zap.Any("tag_metrics", tag_metrics),
				)
				responses := []*gosnowth.DF4Response{}
				for _, metric := range tag_metrics.Items {
					stats.RenderRequests++
					res, err2 := c.client.FetchValues(&gosnowth.FetchQuery{
						Start:  time.Unix(start, 0),
						Period: time.Duration(step) * time.Second,
						Count:  count,
						Streams: []gosnowth.FetchStream{{
							UUID:      metric.UUID,
							Name:      metric.MetricName,
							Kind:      metric.Type,
							Label:     metric.MetricName,
							Transform: "average",
						}},
						Reduce: []gosnowth.FetchReduce{{
							Label:  "pass",
							Method: "pass",
						}},
					})
					if err2 != nil {
						stats.RenderErrors++
						if merry.Is(err, types.ErrTimeoutExceeded) {
							stats.Timeouts++
							stats.RenderTimeouts++
						}
						if e == nil {
							e = merry.Wrap(err2).WithValue("target", query)
						} else {
							e = e.WithCause(err2)
						}
						continue
					}
					responses = append(responses, res)
				}
				logger.Debug("got fetch result from irondb",
					zap.String("query", query),
					zap.Any("responses", responses),
				)
				for _, response := range responses {
					// We always should trust backend's response (to mimic behavior of graphite for grahpite native protoocols)
					// See https://github.com/go-graphite/carbonapi/issues/504 and https://github.com/go-graphite/carbonapi/issues/514
					realStart := start
					realStop := stop
					if len(response.Data) > 0 {
						realStart = response.Head.Start
						realStop = response.Head.Start + response.Head.Period*(response.Head.Count-1)
					}
					for i, meta := range response.Meta {
						values := make([]float64, (realStop-realStart)/step+1)
						for _, data := range response.Data[i] {
							dv, ok := data.(float64)
							if ok {
								values = append(values, dv)
							} else {
								values = append(values, math.NaN())
							}
						}
						name := convertNameToGraphite(meta.Label)
						r.Metrics = append(r.Metrics, protov3.FetchResponse{
							Name:              name,
							PathExpression:    pathExpr,
							ConsolidationFunc: "Average",
							StartTime:         realStart,
							StopTime:          realStop,
							StepTime:          step,
							Values:            values,
							XFilesFactor:      0.0,
						})
					}
				}
			} else {
				// we can use graphite compatible API to fetch non-tagged metrics
				query = target
				logger.Debug("send metric find result to irondb",
					zap.String("query", query),
				)
				stats.FindRequests++
				metrics, err := c.client.GraphiteFindMetrics(c.accountID, c.graphite_prefix, query, nil)
				if err != nil {
					stats.FindErrors++
					if merry.Is(err, types.ErrTimeoutExceeded) {
						stats.Timeouts++
						stats.FindTimeouts++
					}
					if e == nil {
						e = merry.Wrap(err).WithValue("target", query)
					} else {
						e = e.WithCause(err)
					}
					continue
				}
				logger.Debug("got find result from irondb",
					zap.String("target", query),
					zap.Any("metrics", metrics),
				)
				// if we found no metrics - good luck with next target
				if len(metrics) == 0 {
					continue
				}
				names := make([]string, 0, len(metrics)+1)
				for _, metric := range metrics {
					names = append(names, metric.Name)
				}
				lookup := &gosnowth.GraphiteLookup{
					Start: start,
					End:   stop,
					Names: names,
				}
				stats.RenderRequests++
				logger.Debug("send render request to irondb",
					zap.String("query", query),
					zap.Any("lookup", lookup),
				)
				response, err2 := c.client.GraphiteGetDatapoints(c.accountID, c.graphite_prefix, lookup, nil)
				if err2 != nil {
					stats.RenderErrors++
					if merry.Is(err, types.ErrTimeoutExceeded) {
						stats.Timeouts++
						stats.RenderTimeouts++
					}
					if e == nil {
						e = merry.Wrap(err2).WithValue("query", query)
					} else {
						e = e.WithCause(err2)
					}
					continue
				}
				logger.Debug("got fetch result from irondb",
					zap.String("query", query),
					zap.Any("response", response),
				)
				// We always should trust backend's response (to mimic behavior of graphite for grahpite native protoocols)
				// See https://github.com/go-graphite/carbonapi/issues/504 and https://github.com/go-graphite/carbonapi/issues/514
				realStart := start
				realStop := stop
				if len(response.Series) > 0 {
					realStart = response.From
					realStop = response.To
				}
				for name, values := range response.Series {
					label := convertNameToGraphite(name)
					vals := make([]float64, 0, (realStop-realStart)/step+1)
					for _, data := range values {
						if data != nil {
							vals = append(vals, *data)
						} else {
							vals = append(vals, math.NaN())
						}
					}
					r.Metrics = append(r.Metrics, protov3.FetchResponse{
						Name:              label,
						PathExpression:    pathExpr,
						ConsolidationFunc: "Average",
						StartTime:         realStart,
						StopTime:          realStop,
						StepTime:          response.Step,
						Values:            vals,
						XFilesFactor:      0.0,
					})
				}
			}
		}
	}
	if e != nil {
		stats.FailedServers = []string{c.groupName}
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *IronDBGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}

	r := protov3.MultiGlobResponse{
		Metrics: make([]protov3.GlobResponse, 0),
	}
	var e merry.Error

	for _, query := range request.Metrics {
		resp := protov3.GlobResponse{
			Name:    query,
			Matches: make([]protov3.GlobMatch, 0),
		}

		logger.Debug("will do find query",
			zap.Int64("accountID", c.accountID),
			zap.String("query", query),
			zap.String("prefix", c.graphite_prefix),
		)
		stats.FindRequests++
		find_result, err := c.client.GraphiteFindMetrics(c.accountID, c.graphite_prefix, query, nil)
		if err != nil {
			stats.FindErrors++
			if merry.Is(err, types.ErrTimeoutExceeded) {
				stats.Timeouts++
				stats.FindTimeouts++
			}
			if e == nil {
				e = merry.Wrap(err).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
			continue
		}
		logger.Debug("got find result from irondb",
			zap.String("query", query),
			zap.Any("result", find_result),
		)

		for _, metric := range find_result {
			name := convertNameToGraphite(metric.Name)
			resp.Matches = append(resp.Matches, protov3.GlobMatch{
				IsLeaf: metric.Leaf,
				Path:   name,
			})
			r.Metrics = append(r.Metrics, resp)
		}
		logger.Debug("parsed find result",
			zap.Any("result", r.Metrics),
			zap.Int64("start", request.StartTime),
			zap.Int64("stop", request.StopTime),
		)
	}

	if e != nil {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *IronDBGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotSupportedByBackend
}

func (c *IronDBGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}

func (c *IronDBGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotSupportedByBackend
}

func (c *IronDBGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, merry.Error) {
	logger := c.logger
	params := make(map[string][]string)
	var result []string
	var target string
	var tag_category string

	// decoding query
	queryDecoded, _ := url.QueryUnescape(query)
	querySplit := strings.Split(queryDecoded, "&")
	for _, qvRaw := range querySplit {
		idx := strings.Index(qvRaw, "=")
		//no parameters passed
		if idx < 1 {
			continue
		}
		k := qvRaw[:idx]
		v := qvRaw[idx+1:]
		if v2, ok := params[qvRaw[:idx]]; !ok {
			params[k] = []string{v}
		} else {
			v2 = append(v2, v)
			params[k] = v2
		}
	}
	logger.Debug("doTagQuery",
		zap.Any("query", queryDecoded),
		zap.Bool("isTagName", isTagName),
		zap.Any("params", params),
		zap.Int64("limit", limit),
	)

	// default target - all Graphite metrics
	target = `__name=~.*`
	// but use joined expr value if present
	if len(params["expr"]) > 0 {
		target = strings.Join(params["expr"], ",")
	}
	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		// get tag category (if present and we're looking for values instead categories)
		if tag, ok := params["tag"]; ok {
			tag_category = tag[0]
		} else {
			return []string{}, types.ErrNoTagSpecified
		}
	}

	logger.Debug("sending GraphiteFindTags request to irondb",
		zap.Int64("accountID", c.accountID),
		zap.String("prefix", c.graphite_prefix),
		zap.String("target", target),
	)
	tag_result, err := c.client.GraphiteFindTags(c.accountID, c.graphite_prefix, target, nil)
	if err != nil {
		return []string{}, merry.New("request returned an error").WithValue("error", err)
	}
	logger.Debug("got GraphiteFindTags result from irondb",
		zap.String("target", target),
		zap.Any("tag_result", tag_result),
	)
	for _, metric := range tag_result {
		// if metric name contain tags
		if strings.Contains(metric.Name, ";") {
			namex := strings.Split(metric.Name, ";")
			for i := 1; i < len(namex); i++ {
				tag := strings.SplitN(namex[i], "=", 2)
				// tag[0] is tag category and tag[1] is tag value
				if isTagName {
					// if filtering by tag prefix
					if v, ok := params["tagPrefix"]; ok {
						// and prefix match
						if strings.HasPrefix(tag[0], v[0]) {
							// append tag name to result
							result = append(result, tag[0])
						}
					} else {
						// if not filtering by tag prefix - append all tags
						result = append(result, tag[0])
					}
				} else {
					// if filtering by value prefix
					if v, ok := params["valuePrefix"]; ok {
						// and prefix match
						if strings.HasPrefix(tag[1], v[0]) {
							// append tag value to result
							result = append(result, tag[1])
						}
					} else {
						// if not filtering by value prefix
						// append only values belong to requested tag category to result
						if tag[0] == tag_category {
							result = append(result, tag[1])
						}
					}
				}
			}
		}
	}

	// removing duplicates from result list
	seen := make(map[string]struct{}, len(result))
	i := 0
	for _, v := range result {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result[i] = v
		i++
	}
	result = result[:i]

	// cut result if needed
	if limit > 0 && len(result) > int(limit) {
		result = result[:int(limit)]
	}

	return result, nil

}

func (c *IronDBGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, true, query, limit)
}

func (c *IronDBGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, false, query, limit)
}

func (c *IronDBGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
	// ProbeTLDs is not really needed for IronDB but returning nil causing error
	// so, let's return empty list
	return []string{}, nil
}
