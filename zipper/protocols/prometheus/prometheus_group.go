package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus/helpers"
	prometheusTypes "github.com/go-graphite/carbonapi/zipper/protocols/prometheus/types"

	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"

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

type StartDelay struct {
	IsSet      bool
	IsDuration bool
	D          time.Duration
	T          int64
	S          string
}

func (s *StartDelay) String() string {
	if s.IsDuration {
		return strconv.FormatInt(time.Now().Add(s.D).Unix(), 10)
	}
	if s.S == "" {
		s.S = strconv.FormatInt(s.T, 10)
	}
	return s.S
}

// RoundRobin is used to connect to backends inside clientGroups, implements BackendServer interface
type PrometheusGroup struct {
	groupName string
	servers   []string
	protocol  string

	client *http.Client

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	step                 int64
	maxPointsPerQuery    int64
	forceMinStepInterval time.Duration

	startDelay StartDelay

	httpQuery *helper.HttpQuery
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "prometheus"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))

	logger.Warn("support for this backend protocol is experimental, use with caution")
	httpClient := helper.GetHTTPClient(logger, config)

	step := int64(15)
	stepI, ok := config.BackendOptions["step"]
	if ok {
		stepNew, ok := stepI.(string)
		if ok {
			if stepNew[len(stepNew)-1] >= '0' && stepNew[len(stepNew)-1] <= '9' {
				stepNew += "s"
			}
			t, err := time.ParseDuration(stepNew)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "step"),
					zap.String("option_value", stepNew),
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

	maxPointsPerQuery := int64(11000)
	mppqI, ok := config.BackendOptions["max_points_per_query"]
	if ok {
		mppq, ok := mppqI.(int)
		if !ok {
			logger.Fatal("failed to parse max_points_per_query",
				zap.String("type_parsed", fmt.Sprintf("%T", mppqI)),
				zap.String("type_expected", "int"),
			)
		}

		maxPointsPerQuery = int64(mppq)
	}

	var forceMinStepInterval time.Duration
	fmsiI, ok := config.BackendOptions["force_min_step_interval"]
	if ok {
		fmsiS, ok := fmsiI.(string)
		if !ok {
			logger.Fatal("failed to parse force_min_step_interval",
				zap.String("type_parsed", fmt.Sprintf("%T", fmsiI)),
				zap.String("type_expected", "time.Duration"),
			)
		}
		var err error
		forceMinStepInterval, err = time.ParseDuration(fmsiS)
		if err != nil {
			logger.Fatal("failed to parse force_min_step_interval",
				zap.String("value_provided", fmsiS),
				zap.String("type_expected", "time.Duration"),
			)
		}
	}

	delay := StartDelay{
		IsSet:      false,
		IsDuration: false,
		T:          -1,
	}
	startI, ok := config.BackendOptions["start"]
	if ok {
		delay.IsSet = true
		startNew, ok := startI.(string)
		if ok {
			startNewInt, err := strconv.Atoi(startNew)
			if err != nil {
				d, err2 := time.ParseDuration(startNew)
				if err2 != nil {
					logger.Fatal("failed to parse option",
						zap.String("option_name", "start"),
						zap.String("option_value", startNew),
						zap.Errors("errors", []error{err, err2}),
					)
				}
				delay.IsDuration = true
				delay.D = d
			} else {
				delay.T = int64(startNewInt)
			}
		}
	}

	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	return NewWithEverythingInitialized(logger, config, tldCacheDisabled, requireSuccessAll, limiter, step, maxPointsPerQuery, forceMinStepInterval, delay, httpQuery, httpClient)
}

func NewWithEverythingInitialized(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter, step, maxPointsPerQuery int64, forceMinStepInterval time.Duration, delay StartDelay, httpQuery *helper.HttpQuery, httpClient *http.Client) (types.BackendServer, merry.Error) {
	c := &PrometheusGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: *config.MaxBatchSize,
		step:                 step,
		forceMinStepInterval: forceMinStepInterval,
		maxPointsPerQuery:    maxPointsPerQuery,
		startDelay:           delay,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
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

func (c *PrometheusGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/api/v1/query_range")

	pathExprToTargets := make(map[string][]string)
	for _, m := range request.Metrics {
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, m.Name)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error

	start := request.Metrics[0].StartTime
	stop := request.Metrics[0].StopTime

	maxPointsPerQuery := c.maxPointsPerQuery
	if len(request.Metrics) > 0 && request.Metrics[0].MaxDataPoints != 0 {
		maxPointsPerQuery = request.Metrics[0].MaxDataPoints
	}
	step := helpers.AdjustStep(start, stop, maxPointsPerQuery, c.step, c.forceMinStepInterval)

	stepStr := strconv.FormatInt(step, 10)
	for pathExpr, targets := range pathExprToTargets {
		for _, target := range targets {
			logger.Debug("got some target to query",
				zap.Any("pathExpr", pathExpr),
				zap.Any("target", target),
			)
			// rewrite metric for Tag
			// Make local copy
			stepLocalStr := stepStr
			if strings.HasPrefix(target, "seriesByTag") {
				stepLocalStr, target = helpers.SeriesByTagToPromQL(stepLocalStr, target)
			} else {
				reQuery := helpers.ConvertGraphiteTargetToPromQL(target)
				target = fmt.Sprintf("{__name__=~%q}", reQuery)
			}
			if stepLocalStr[len(stepLocalStr)-1] >= '0' && stepLocalStr[len(stepLocalStr)-1] <= '9' {
				stepLocalStr += "s"
			}
			t, err := time.ParseDuration(stepLocalStr)
			if err != nil {
				stats.RenderErrors++
				logger.Debug("failed to parse step",
					zap.String("step", stepLocalStr),
					zap.Error(err),
				)
				if e == nil {
					e = merry.Wrap(err)
				}
				continue
			}
			stepLocal := int64(t.Seconds())
			/*
				newStep, err3 := strToStep(stepStr)
				if err3 == nil {
					step = newStep
				}
			*/
			logger.Debug("will do query",
				zap.String("query", target),
				zap.Int64("start", start),
				zap.Int64("stop", stop),
				zap.String("step", stepLocalStr),
			)
			v := url.Values{
				"query": []string{target},
				"start": []string{strconv.Itoa(int(start))},
				"end":   []string{strconv.Itoa(int(stop))},
				"step":  []string{stepLocalStr},
			}

			rewrite.RawQuery = v.Encode()
			stats.RenderRequests++
			res, err2 := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
			if err2 != nil {
				stats.RenderErrors++
				if merry.Is(err, types.ErrTimeoutExceeded) {
					stats.Timeouts++
					stats.RenderTimeouts++
				}
				if e == nil {
					e = err2
				} else {
					e = e.WithCause(err2)
				}
				continue
			}

			var response prometheusTypes.HTTPResponse
			err = json.Unmarshal(res.Response, &response)
			if err != nil {
				stats.RenderErrors++
				c.logger.Debug("failed to unmarshal response",
					zap.Error(err),
				)
				if e == nil {
					e = err2
				} else {
					e = e.WithCause(err2)
				}
				continue
			}

			if response.Status != "success" {
				stats.RenderErrors++
				if e == nil {
					e = types.ErrFailedToFetch.WithMessage(response.Status).WithValue("query", target).WithValue("status", response.Status)
				} else {
					e = e.WithCause(err2).WithValue("query", target).WithValue("status", response.Status)
				}
				continue
			}

			for _, m := range response.Data.Result {
				// We always should trust backend's response (to mimic behavior of graphite for grahpite native protoocols)
				// See https://github.com/go-graphite/carbonapi/issues/504 and https://github.com/go-graphite/carbonapi/issues/514
				realStart := start
				realStop := stop
				if len(m.Values) > 0 {
					realStart = int64(m.Values[0].Timestamp)
					realStop = int64(m.Values[len(m.Values)-1].Timestamp)
				}
				alignedValues := helpers.AlignValues(realStart, realStop, stepLocal, m.Values)

				r.Metrics = append(r.Metrics, protov3.FetchResponse{
					Name:              helpers.PromMetricToGraphite(m.Metric),
					PathExpression:    pathExpr,
					ConsolidationFunc: "Average",
					StartTime:         realStart,
					StopTime:          realStop,
					StepTime:          stepLocal,
					Values:            alignedValues,
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

func (c *PrometheusGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	logger := c.logger.With(zap.String("type", "find"), zap.Strings("request", request.Metrics))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/api/v1/series")

	r := protov3.MultiGlobResponse{
		Metrics: make([]protov3.GlobResponse, 0),
	}
	var e merry.Error
	uniqueMetrics := make(map[string]bool)
	for _, query := range request.Metrics {
		// Convert query to Prometheus-compatible regex
		if !strings.HasSuffix(query, "*") {
			query = query + "*"
		}

		reQuery := helpers.ConvertGraphiteTargetToPromQL(query)
		matchQuery := fmt.Sprintf("{__name__=~%q}", reQuery)
		v := url.Values{
			"match[]": []string{matchQuery},
		}

		if c.startDelay.IsSet {
			v.Add("start", c.startDelay.String())
		}

		rewrite.RawQuery = v.Encode()
		stats.FindRequests += 1
		res, err := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if err != nil {
			stats.FindErrors += 1
			if merry.Is(err, types.ErrTimeoutExceeded) {
				stats.Timeouts += 1
				stats.FindTimeouts += 1
			}
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		var pr prometheusTypes.PrometheusFindResponse

		err2 := json.Unmarshal(res.Response, &pr)
		if err2 != nil {
			stats.FindErrors += 1
			if e == nil {
				e = err
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		if pr.Status != "success" {
			stats.FindErrors += 1
			if e == nil {
				e = types.ErrFailedToFetch.WithMessage(pr.Error).WithValue("query", matchQuery).WithValue("error_type", pr.ErrorType).WithValue("error", pr.Error)
			} else {
				e = e.WithCause(err2).WithValue("query", matchQuery).WithValue("error_type", pr.ErrorType).WithValue("error", pr.Error)
			}
			continue
		}

		querySplit := strings.Split(query, ".")
		resp := protov3.GlobResponse{
			Name:    query,
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
				Path:   k,
			})
			r.Metrics = append(r.Metrics, resp)
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

func (c *PrometheusGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotSupportedByBackend
}

func (c *PrometheusGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotImplementedYet
}
func (c *PrometheusGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, types.ErrNotSupportedByBackend
}

func (c *PrometheusGroup) doSimpleTagQuery(ctx context.Context, logger *zap.Logger, isTagName bool, params map[string][]string, limit int64) ([]string, merry.Error) {
	var rewrite *url.URL

	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
		rewrite, _ = url.Parse("http://127.0.0.1/api/v1/labels")
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		if tag, ok := params["Tag"]; ok {
			rewrite, _ = url.Parse(fmt.Sprintf("http://127.0.0.1/api/v1/label/%s/values", tag[0]))
		} else {
			return []string{}, types.ErrNoTagSpecified
		}
	}

	var r prometheusTypes.PrometheusTagResponse

	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, merry.Wrap(err)
	}

	if r.Status != "success" {
		return []string{}, merry.New("request returned an error").WithValue("status", r.Status).WithValue("error_type", r.ErrorType).WithValue("error", r.Error)
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

func (c *PrometheusGroup) doComplexTagQuery(ctx context.Context, isTagName bool, params map[string][]string, limit int64) ([]string, merry.Error) {
	logger := c.logger
	var rewrite *url.URL

	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		if _, ok := params["Tag"]; !ok {
			return []string{}, types.ErrNoTagSpecified
		}
	}

	matches := make([]string, 0, len(params["expr"]))
	for _, e := range params["expr"] {
		name, t := helpers.PromethizeTagValue(e)
		matches = append(matches, "{"+name+t.OP+"\""+t.TagValue+"\"}")
	}

	rewrite, _ = url.Parse("http://127.0.0.1/api/v1/series")
	v := url.Values{
		"match[]": matches,
	}
	rewrite.RawQuery = v.Encode()

	result := make([]string, 0)
	var r prometheusTypes.PrometheusFindResponse

	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return []string{}, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return []string{}, merry.Wrap(err)
	}

	if r.Status != "success" {
		return []string{}, merry.New("request returned an error").WithValue("status", r.Status).WithValue("error_type", r.ErrorType).WithValue("error", r.Error)
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
		tag := params["Tag"][0]
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

func (c *PrometheusGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, merry.Error) {
	logger := c.logger
	params := make(map[string][]string)
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
		zap.Any("params", params),
	)

	if _, ok := params["expr"]; !ok {
		return c.doSimpleTagQuery(ctx, logger, isTagName, params, limit)
	}

	return c.doComplexTagQuery(ctx, isTagName, params, limit)
}

func (c *PrometheusGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, true, query, limit)
}

func (c *PrometheusGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return c.doTagQuery(ctx, false, query, limit)
}

func (c *PrometheusGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
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
