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
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
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

	accountID      int64
	graphiteRollup int64
	graphitePrefix string
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "irondb"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

	logger.Warn("support for this backend protocol is experimental, use with caution")

	// initializing config with list of servers from upstream
	cfg := gosnowth.NewConfig(config.Servers...)

	// enabling discovery.
	cfg.Discover = true

	// parse backend options
	var tmpInt int
	accountID := int64(1)
	if accountIDOpt, ok := config.BackendOptions["irondb_account_id"]; ok {
		if tmpInt, ok = accountIDOpt.(int); !ok {
			logger.Fatal("failed to parse irondb_account_id",
				zap.String("type_parsed", fmt.Sprintf("%T", accountIDOpt)),
				zap.String("type_expected", "int"),
			)
		}
		accountID = int64(tmpInt)
	}

	maxTries := int64(*config.MaxTries)
	retries := int64(0)
	if retriesOpt, ok := config.BackendOptions["irondb_retries"]; ok {
		if tmpInt, ok = retriesOpt.(int); !ok {
			logger.Fatal("failed to parse irondb_retries",
				zap.String("type_parsed", fmt.Sprintf("%T", retriesOpt)),
				zap.String("type_expected", "int"),
			)
		}
		retries = int64(tmpInt)
	}
	if maxTries > retries {
		retries = maxTries
	}
	cfg.Retries = retries

	connectRetries := int64(-1)
	if connectRetriesOpt, ok := config.BackendOptions["irondb_connect_retries"]; ok {
		if tmpInt, ok = connectRetriesOpt.(int); !ok {
			logger.Fatal("failed to parse irondb_connect_retries",
				zap.String("type_parsed", fmt.Sprintf("%T", connectRetriesOpt)),
				zap.String("type_expected", "int"),
			)
		}
		connectRetries = int64(tmpInt)
	}
	cfg.ConnectRetries = connectRetries

	var tmpStr string
	dialTimeout := 500 * time.Millisecond
	if dialTimeoutOpt, ok := config.BackendOptions["irondb_dial_timeout"]; ok {
		if tmpStr, ok = dialTimeoutOpt.(string); ok {
			interval, err := time.ParseDuration(tmpStr)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_dial_timeout"),
					zap.String("option_value", tmpStr),
					zap.Errors("errors", []error{err}),
				)
			}
			dialTimeout = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_dial_timeout"),
				zap.Any("option_value", tmpStr),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.DialTimeout = dialTimeout

	irondbTimeout := 10 * time.Second
	if irondbTimeoutOpt, ok := config.BackendOptions["irondb_timeout"]; ok {
		if tmpStr, ok = irondbTimeoutOpt.(string); ok {
			interval, err := time.ParseDuration(tmpStr)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_timeout"),
					zap.String("option_value", tmpStr),
					zap.Errors("errors", []error{err}),
				)
			}
			irondbTimeout = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_timeout"),
				zap.Any("option_value", tmpStr),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.Timeout = irondbTimeout

	watchInterval := 30 * time.Second
	if watchIntervalOpt, ok := config.BackendOptions["irondb_watch_interval"]; ok {
		if tmpStr, ok = watchIntervalOpt.(string); ok {
			interval, err := time.ParseDuration(tmpStr)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "irondb_watch_interval"),
					zap.String("option_value", tmpStr),
					zap.Errors("errors", []error{err}),
				)
			}
			watchInterval = interval
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "irondb_watch_interval"),
				zap.Any("option_value", tmpStr),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}
	cfg.WatchInterval = watchInterval

	graphiteRollup := int64(60)
	if graphiteRollupOpt, ok := config.BackendOptions["irondb_graphite_rollup"]; ok {
		if tmpInt, ok = graphiteRollupOpt.(int); !ok {
			logger.Fatal("failed to parse irondb_graphite_rollup",
				zap.String("type_parsed", fmt.Sprintf("%T", graphiteRollupOpt)),
				zap.String("type_expected", "int"),
			)
		}
		graphiteRollup = int64(tmpInt)
	}

	graphitePrefix := ""
	if graphitePrefixOpt, ok := config.BackendOptions["irondb_graphite_prefix"]; ok {
		if tmpStr, ok = graphitePrefixOpt.(string); !ok {
			logger.Fatal("failed to parse irondb_graphite_prefix",
				zap.String("type_parsed", fmt.Sprintf("%T", graphitePrefixOpt)),
				zap.String("type_expected", "string"),
			)
		}
		graphitePrefix = tmpStr
	}

	snowthClient, err := gosnowth.NewClient(context.Background(), cfg)
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
		graphiteRollup:       graphiteRollup,
		graphitePrefix:       graphitePrefix,
		limiter:              limiter,
		logger:               logger,
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
	l := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, tldCacheDisabled, requireSuccessAll, l)
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

func processFindErrors(err error, e merry.Error, stats *types.Stats, query string) merry.Error {
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
	return e
}

func processRenderErrors(err error, e merry.Error, stats *types.Stats, query string) merry.Error {
	stats.RenderErrors++
	if merry.Is(err, types.ErrTimeoutExceeded) {
		stats.Timeouts++
		stats.RenderTimeouts++
	}
	if e == nil {
		e = merry.Wrap(err).WithValue("query", query)
	} else {
		e = e.WithCause(err)
	}
	return e
}

func (c *IronDBGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		pathExprToTargets[m.PathExpression] = append(pathExprToTargets[m.PathExpression], m.Name)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error

	start := request.Metrics[0].StartTime
	stop := request.Metrics[0].StopTime
	step := c.graphiteRollup
	count := int64((stop-start)/step) + 1
	maxCount := request.Metrics[0].MaxDataPoints
	if count > maxCount && maxCount > 0 {
		count = maxCount
		step = adjustStep(start, stop, maxCount, step)
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
				query = target[12 : len(target)-1] // 12 is len("seriesByTag(")
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
				tagMetrics, err := c.client.FindTags(c.accountID, query, findTagOptions)
				if err != nil {
					e = processFindErrors(err, e, stats, query)
					continue
				}
				logger.Debug("got tag find result from irondb",
					zap.String("query", query),
					zap.Any("tagMetrics", tagMetrics),
				)
				responses := []*gosnowth.DF4Response{}
				for _, metric := range tagMetrics.Items {
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
						e = processRenderErrors(err2, e, stats, query)
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
				continue
			}
			// target is not starting with seriesByTag
			// we can use graphite compatible API to fetch non-tagged metrics
			query = target
			logger.Debug("send metric find result to irondb",
				zap.String("query", query),
			)
			stats.FindRequests++
			metrics, err := c.client.GraphiteFindMetrics(c.accountID, c.graphitePrefix, query, nil)
			if err != nil {
				e = processFindErrors(err, e, stats, query)
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
			response, err2 := c.client.GraphiteGetDatapoints(c.accountID, c.graphitePrefix, lookup, nil)
			if err2 != nil {
				e = processRenderErrors(err2, e, stats, query)
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
			zap.String("prefix", c.graphitePrefix),
		)
		stats.FindRequests++
		findResult, err := c.client.GraphiteFindMetrics(c.accountID, c.graphitePrefix, query, nil)
		if err != nil {
			e = processFindErrors(err, e, stats, query)
			continue
		}
		logger.Debug("got find result from irondb",
			zap.String("query", query),
			zap.Any("result", findResult),
		)

		for _, metric := range findResult {
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
	var tagCategory string

	// decoding query
	queryDecoded, _ := url.QueryUnescape(query)
	querySplit := strings.Split(queryDecoded, "&")
	for _, qvRaw := range querySplit {
		idx := strings.Index(qvRaw, "=")
		// no parameters passed
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
	tagPrefix := ""
	valuePrefix := ""
	if v, ok := params["tagPrefix"]; ok {
		tagPrefix = v[0]
	}
	if v, ok := params["valuePrefix"]; ok {
		valuePrefix = v[0]
	}
	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		// get tag category from tag parameter
		// if it's present and we're looking for values instead categories
		if tagParam, ok := params["tag"]; ok {
			tagCategory = tagParam[0]
		} else {
			return []string{}, types.ErrNoTagSpecified
		}
	}

	logger.Debug("sending GraphiteFindTags request to irondb",
		zap.Int64("accountID", c.accountID),
		zap.String("prefix", c.graphitePrefix),
		zap.String("target", target),
	)
	tagResult, err := c.client.GraphiteFindTags(c.accountID, c.graphitePrefix, target, nil)
	if err != nil {
		return []string{}, merry.New("request returned an error").WithValue("error", err)
	}
	logger.Debug("got GraphiteFindTags result from irondb",
		zap.String("target", target),
		zap.Any("tagResult", tagResult),
	)

	// struct for dedup
	seen := make(map[string]struct{})
	for _, metric := range tagResult {
		tagList := strings.Split(metric.Name, ";")
		// skipping first element (metric name)
		for i := 1; i < len(tagList); i++ {
			// tags[0] is tag category and tags[1] is tag value
			tags := strings.SplitN(tagList[i], "=", 2)
			r := ""
			if isTagName {
				// this is true for empty prefix too
				if strings.HasPrefix(tags[0], tagPrefix) {
					r = tags[0]
				}
			} else {
				if tags[0] == tagCategory && strings.HasPrefix(tags[1], valuePrefix) {
					r = tags[1]
				}
			}
			// if we got something - append to result if unique
			if len(r) > 0 {
				if _, ok := seen[r]; ok {
					continue
				}
				seen[r] = struct{}{}
				result = append(result, r)
			}
		}
	}

	// cut result if needed
	if limit > 0 && int64(len(result)) > limit {
		result = result[:limit]
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
