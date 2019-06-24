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
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

// RoundRobin is used to connect to backends inside clientGroups, implements BackendServer interface
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

	step int64

	httpQuery *helper.HttpQuery
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, limiter *limiter.ServerLimiter) (types.BackendServer, *errors.Errors) {
	logger = logger.With(zap.String("type", "prometheus"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

	logger.Warn("support for this backend protocol is experimental, use with caution")

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

	step := int64(15)
	stepI, ok := config.BackendOptions["step"]
	if ok {
		stepNew, ok := stepI.(string)
		if ok {
			if stepNew[len(stepNew) - 1] >= '0' && stepNew[len(stepNew) - 1] <= '9' {
				stepNew += "s"
			}
			t, err := time.ParseDuration(stepNew)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "step"),
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

	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &PrometheusGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: config.MaxBatchSize,
		step: step,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}
	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2) (types.BackendServer, *errors.Errors) {
	if config.ConcurrencyLimit == nil {
		return nil, errors.Fatal("concurency limit is not set")
	}
	if len(config.Servers) == 0 {
		return nil, errors.Fatal("no servers specified")
	}
	l := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, l)
}

func (c *PrometheusGroup) Children() []types.BackendServer {
	return []types.BackendServer{c}
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
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/api/v1/query_range")

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	e := errors.Errors{}
	// TODO: Do something clever with "step"
	step := c.step
	stepStr := strconv.FormatInt(step, 10)
	for pathExpr, targets := range pathExprToTargets {
		for _, target := range targets {
			logger.Debug("got some target to query",
				zap.Any("pathExpr", pathExpr),
				zap.Any("target", target),
			)
			// rewrite metric for tag
			// Make local copy
			stepLocal := step
			stepLocalStr := stepStr
			if strings.HasPrefix(target, "seriesByTag") {
				stepLocalStr, target = c.convertGraphiteQueryToProm(stepLocalStr, target)
			}
			if stepLocalStr[len(stepLocalStr) - 1] >= '0' && stepLocalStr[len(stepLocalStr) - 1] <= '9' {
				stepLocalStr += "s"
			}
			t, err := time.ParseDuration(stepLocalStr)
			if err != nil {
				logger.Debug("failed to parse step",
					zap.String("step", stepLocalStr),
					zap.Error(err),
					)
				e.Add(err)
				continue
			}
			stepLocal = int64(t.Seconds())
			/*
			newStep, err3 := strToStep(stepStr)
			if err3 == nil {
				step = newStep
			}
			 */
			logger.Debug("will do query",
				zap.String("query", target),
				zap.Int64("start", request.Metrics[0].StartTime),
				zap.Int64("stop", request.Metrics[0].StopTime),
				)
			v := url.Values{
				"query": []string{target},
				"start": []string{strconv.Itoa(int(request.Metrics[0].StartTime))},
				"stop":  []string{strconv.Itoa(int(request.Metrics[0].StopTime))},
				"step":  []string{stepLocalStr},
			}
			rewrite.RawQuery = v.Encode()
			res, err2 := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
			if err2 != nil {
				err2.HaveFatalErrors = false
				e.Merge(err2)
				continue
			}

			var response prometheusResponse
			err = json.Unmarshal(res.Response, &response)
			if err != nil {
				c.logger.Debug("failed to unmarshal response",
					zap.Error(err),
				)
				e.Add(err)
				continue
			}

			if response.Status != "success" {
				e.Addf("query=%s, err=%s", target, response.Status)
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
					StartTime:         request.Metrics[0].StartTime,
					StopTime:          request.Metrics[0].StopTime,
					StepTime:          stepLocal,
					Values:            vals,
					XFilesFactor:      0.0,
				})
			}
		}
	}

	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
		return &r, stats, &e
	}
	return &r, stats, nil
}

func (c *PrometheusGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/api/v1/series")

	r := protov3.MultiGlobResponse{
		Metrics: make([]protov3.GlobResponse, 0),
	}
	e := errors.Errors{}
	uniqueMetrics := make(map[string]bool)
	for _, query := range request.Metrics {
		// Convert query to Prometheus-compatible regex
		reQuery := strings.Replace(query, ".", "\\\\.", -1)
		reQuery = strings.Replace(query, "*", "[^.][^.]*", -1)
		if reQuery[len(reQuery) - 1] == '*' {
			reQuery += ".*"
		}
		matchQuery := "{__name__=~\"" + reQuery + "\"}"
		v := url.Values{
			"match[]": []string{matchQuery},
		}
		rewrite.RawQuery = v.Encode()
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			e.Merge(err)
			continue
		}

		var pr prometheusFindResponse

		err2 := json.Unmarshal(res.Response, &pr)
		if err2 != nil {
			e.Add(err2)
			continue
		}

		if pr.Status != "success" {
			e.Addf("status=%s, errorType=%s, error=%s", pr.Status, pr.ErrorType, pr.Error)
			continue
		}

		querySplit := strings.Split(query, ".")
		resp := protov3.GlobResponse{
			Name: query,
			Matches: make([]protov3.GlobMatch, 0),
		}
		for _, m := range pr.Data {
			name, ok := m["__name__"]
			if !ok {
				continue
			}
			nameSplit := strings.Split(name, ".")

			if len(querySplit) > len(nameSplit) {
				continue
			}

			isLeaf := false
			if len(nameSplit) == len(querySplit) {
				isLeaf = true
			}

			uniqueMetrics[strings.Join(nameSplit[:len(querySplit)], ".")] = isLeaf
		}

		for k, v := range uniqueMetrics {
			resp.Matches = append(resp.Matches, protov3.GlobMatch{
				IsLeaf: v,
				Path: k,
			})
			r.Metrics = append(r.Metrics, resp)
		}
	}

	if len(r.Metrics) == 0 {
		e.Add(types.ErrNoResponseFetched)
	}
	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
		return &r, stats, &e
	}
	return &r, stats, nil
}

func (c *PrometheusGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotSupportedByBackend)
}

func (c *PrometheusGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (c *PrometheusGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotSupportedByBackend)
}

func (c *PrometheusGroup) doSimpleTagQuery(ctx context.Context, logger *zap.Logger, isTagName bool, params map[string][]string, limit int64) ([]string, *errors.Errors) {
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

	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, errors.FromErr(err)
	}

	if r.Status != "success" {
		return []string{}, errors.Errorf("status=%s, errorType=%s, error=%s", r.Status, r.ErrorType, r.Error)
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
		zap.Any("result", r),
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

	rewrite, _ = url.Parse("http://127.0.0.1/api/v1/series")
	v := url.Values{
		"match[]": matches,
	}
	rewrite.RawQuery = v.Encode()

	result := make([]string, 0)
	var r prometheusFindResponse

	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, errors.FromErr(err)
	}

	if r.Status != "success" {
		return []string{}, errors.Errorf("status=%s, errorType=%s, error=%s", r.Status, r.ErrorType, r.Error)
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
		zap.Any("result", result),
	)

	return result, nil
}

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
		return c.doSimpleTagQuery(ctx, logger, isTagName, params, limit)
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
