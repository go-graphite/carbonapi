package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/protocols/graphite/msgpack"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func init() {
	aliases := []string{"prometheus"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGroups, implements ServerClient interface
type PrometheusGroup struct {
	groupName string
	servers   []string
	protocol  string

	client *http.Client

	limiter              *limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	httpQuery *helper.HttpQuery
}

func (g *PrometheusGroup) Children() []types.ServerClient {
	return []types.ServerClient{g}
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, limiter *limiter.ServerLimiter) (types.ServerClient, *errors.Errors) {
	logger = logger.With(zap.String("type", "prometheus"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

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

	httpQuery := helper.NewHttpQuery(logger, config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &PrometheusGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: config.MaxBatchSize,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2) (types.ServerClient, *errors.Errors) {
	if config.ConcurrencyLimit == nil {
		return nil, errors.Fatal("concurency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, errors.Fatal("no servers specified")
	}
	limiter := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, limiter)
}

func (c PrometheusGroup) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c PrometheusGroup) Name() string {
	return c.groupName
}

func (c PrometheusGroup) Backends() []string {
	return c.servers
}

func (c *PrometheusGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, *errors.Errors) {
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/api/v1/query_range")

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	// TODO: Do something clever with "step"
	step := int64(15)
	for pathExpr, targets := range pathExprToTargets {
		for _, target := range targets {
			c.logger.Debug("got some target to query",
				zap.Any("pathExpr", pathExpr),
				zap.Any("target", target),
			)
			// rewrite metric for tag
			if strings.HasPrefix(target, "seriesByTag") {
				target = c.convertGraphiteQueryToProm(target)
			}
			c.logger.Debug("will do query",
				zap.String("query", target),
				zap.Int64("start", request.Metrics[0].StartTime),
				zap.Int64("stop", request.Metrics[0].StopTime),
				)
			stepStr := strconv.FormatInt(step, 10)
			v := url.Values{
				"query": []string{target},
				"start": []string{strconv.Itoa(int(request.Metrics[0].StartTime))},
				"stop":  []string{strconv.Itoa(int(request.Metrics[0].StopTime))},
				"step":  []string{stepStr},
			}
			rewrite.RawQuery = v.Encode()
			res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
			if err == nil {
				err = &errors.Errors{}
			}
			if err.HaveFatalErrors {
				err.HaveFatalErrors = false
				return nil, stats, err
			}

			var response prometheusResponse
			err2 := json.Unmarshal(res.Response, &response)
			if err2 != nil {
				c.logger.Debug("failed to unmarshal response",
					zap.Error(err2),
				)
				err.Add(err2)
				continue
			}

			if response.Status != "success" {
				err.Addf("query=%s, err=%s", target, response.Status)
				continue
			}

			for _, m := range response.Data.Result {
				vals := make([]float64, len(m.Values))
				// TODO: Verify timestamps
				for i, v := range m.Values {
					vals[i] = v.Value
				}
				r.Metrics = append(r.Metrics, protov3.FetchResponse{
					Name:              c.promMetricToGraphite(m.Metric),
					PathExpression:    pathExpr,
					ConsolidationFunc: "Average",
					StopTime:          request.Metrics[0].StopTime,
					StartTime:         request.Metrics[0].StartTime,
					StepTime:          step,
					Values:            vals,
					XFilesFactor:      0.0,
				})
			}
		}
	}

	return &r, stats, nil
}

func (c *PrometheusGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/metrics/find/")

	var r protov3.MultiGlobResponse
	r.Metrics = make([]protov3.GlobResponse, 0)
	var e errors.Errors
	for _, query := range request.Metrics {
		v := url.Values{
			"query":  []string{query},
			"format": []string{c.protocol},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
		if err != nil {
			e.Merge(err)
			continue
		}
		var globs msgpack.MultiGraphiteGlobResponse
		_, marshalErr := globs.UnmarshalMsg(res.Response)
		if marshalErr != nil {
			e.Add(marshalErr)
			continue
		}

		stats.Servers = append(stats.Servers, res.Server)
		matches := make([]protov3.GlobMatch, 0, len(globs))
		for _, m := range globs {
			matches = append(matches, protov3.GlobMatch{
				Path:   m.Path,
				IsLeaf: m.IsLeaf,
			})
		}
		r.Metrics = append(r.Metrics, protov3.GlobResponse{
			Name:    query,
			Matches: matches,
		})
	}

	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
	}

	if len(r.Metrics) == 0 {
		return nil, stats, errors.FromErr(types.ErrNoResponseFetched)
	}
	return &r, stats, nil
}

func (c *PrometheusGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

func (c *PrometheusGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (c *PrometheusGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

func (c *PrometheusGroup) doSimpleTagQuery(ctx context.Context, isTagName bool, params map[string][]string, limit int64) ([]string, *errors.Errors) {
	logger := c.logger
	var rewrite *url.URL

	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
		rewrite, _ = url.Parse("http://127.0.0.1/api/v1/labels")
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		if tag, ok := params["tag"]; ok {
			rewrite, _ = url.Parse(fmt.Sprintf("http://127.0.0.1/api/v1/label/%s/values", tag[0]))
		} else {
			return []string{}, errors.Fatal("no tag specified")
		}
	}

	var r prometheusTagResponse

	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, errors.FromErr(err)
	}

	if r.Status != "success" {
		return []string{}, errors.Error(r.Status)
	}

	if isTagName {
		if v, ok := params["tagPrefix"]; ok {
			data := make([]string, 0)
			for _, t := range r.Data {
				if strings.HasPrefix(t, v[0]) {
					data = append(data, t)
				}
			}
			r.Data = data
		}
	} else {
		if v, ok := params["valuePrefix"]; ok {
			data := make([]string, 0)
			for _, t := range r.Data {
				if strings.HasPrefix(t, v[0]) {
					data = append(data, t)
				}
			}
			r.Data = data
		}
	}

	if limit > 0 && len(r.Data) > int(limit) {
		r.Data = r.Data[:int(limit)]
	}

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	return r.Data, nil
}

func (c *PrometheusGroup) doComplexTagQuery(ctx context.Context, isTagName bool, params map[string][]string, limit int64) ([]string, *errors.Errors) {
	logger := c.logger
	var rewrite *url.URL

	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		if _, ok := params["tag"]; !ok {
			return []string{}, errors.Fatal("no tag specified")
		}
	}

	matches := make([]string, 0, len(params["expr"]))
	for _, e := range params["expr"] {
		name, t := c.promethizeTagValue(e)
		matches = append(matches, "{" + name + t.OP + "\"" + t.TagValue + "\"}")
	}

	rewrite, _ = url.Parse(fmt.Sprintf("http://127.0.0.1/api/v1/series"))
	v := url.Values{
		"match[]": matches,
	}
	rewrite.RawQuery = v.Encode()

	result := make([]string, 0)
	var r prometheusFindResponse

	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, errors.FromErr(err)
	}

	if r.Status != "success" {
		return []string{}, errors.Error(r.Status)
	}

	var prefix string
	if isTagName {
		if prefixArr, ok := params["tagPrefix"]; ok {
			prefix = prefixArr[0]
		}

		uniqueTagNames := make(map[string]struct{})
		for _, d := range r.Data {
			for k := range d {
				if strings.HasPrefix(k, prefix) {
					uniqueTagNames[k] = struct{}{}
				}
			}
		}
		for k := range uniqueTagNames {
			result = append(result, k)
		}
	} else {
		if prefixArr, ok := params["valuePrefix"]; ok {
			prefix = prefixArr[0]
		}

		uniqueTagValues := make(map[string]struct{})
		tag := params["tag"][0]
		for _, d := range r.Data {
			if v, ok := d[tag]; ok {
				if strings.HasPrefix(v, prefix) {
					uniqueTagValues[v] = struct{}{}
				}
			}
		}
		for v := range uniqueTagValues {
			result = append(result, v)
		}
	}

	if limit > 0 && len(result) > int(limit) {
		result = result[:int(limit)]
	}

	logger.Debug("got client response",
		zap.Any("r", result),
	)

	return result, nil
}

// TODO: Handle 'expr' query as well
func (c *PrometheusGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, *errors.Errors) {
	logger := c.logger
	params := make(map[string][]string)
	queryDecoded, _ := url.QueryUnescape(query)
	querySplit := strings.Split(queryDecoded, "&")
	for _, qvRaw := range querySplit {
		idx := strings.Index(qvRaw, "=")
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
		zap.Any("params", params),
	)

	if _, ok := params["expr"]; !ok {
		return c.doSimpleTagQuery(ctx, isTagName, params, limit)
	}

	return c.doComplexTagQuery(ctx, isTagName, params, limit)
}

func (c *PrometheusGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, *errors.Errors) {
	return c.doTagQuery(ctx, true, query, limit)
}

func (c *PrometheusGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, *errors.Errors) {
	return c.doTagQuery(ctx, false, query, limit)
}

func (c *PrometheusGroup) ProbeTLDs(ctx context.Context) ([]string, *errors.Errors) {
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
