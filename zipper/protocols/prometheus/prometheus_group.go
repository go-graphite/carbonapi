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
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
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

type tag struct {
	TagValue string
	OP string
}

func (c *PrometheusGroup) splitTagValues(query string) map[string]tag {
	tags := strings.Split(query, ",")
	result := make(map[string]tag)
	for _, tvString := range tags {
		tvString = strings.TrimSpace(tvString)
		// Handle = and =~
		t := tag{}
		idx := strings.Index(tvString, "=")
		if idx != -1 {
			if tvString[idx+1] == '~' {
				t.OP = "=~"
				t.TagValue = tvString[idx+2:len(tvString)-1]
			} else {
				t.OP = "="
				t.TagValue = tvString[idx+1:len(tvString)-1]
			}
		} else {
			// Handle != and !=~
			idx = strings.Index(tvString, "!")
			if tvString[idx+2] == '~' {
				t.OP = "!~"
				t.TagValue = tvString[idx+3:len(tvString)-1]
			} else {
				t.OP = "!="
				t.TagValue = tvString[idx+2:len(tvString)-1]
			}
		}
		// Skip first \" sign
		result[tvString[1:idx]] = t
	}
	return result
}

func (c *PrometheusGroup) promMetricToGraphite(metric map[string]string) string {
	var res strings.Builder

	res.WriteString(metric["__name__"])
	delete(metric, "__name__")

	keys := make([]string, 0, len(metric))
	for k := range metric {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		res.WriteString(";" + k + "=" + metric[k])
	}

	return res.String()
}

type prometheusValue struct {
	Timestamp int
	Value float64
}

func (p *prometheusValue) UnmarshalJSON(data []byte) error {
	arr := make([]interface{}, 0)
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}
	if len(arr) != 2 {
		return fmt.Errorf("length mismatch, got %v, expected 2", len(arr))
	}
	ts, ok := arr[0].(float64)
	if !ok {
		return fmt.Errorf("type mismatch for element[0/1], expected 'float64', got '%T', str=%v", arr[0], string(data))
	}
	p.Timestamp = int(ts)

	str, ok := arr[1].(string)
	if !ok {
		return fmt.Errorf("type mismatch for element[1/1], expected 'string', got '%T', str=%v", arr[1], string(data))
	}

	switch str {
	case "NaN":
		p.Value = math.NaN()
		return nil
	case "+Inf":
		p.Value = math.Inf(1)
		return nil
	case "-Inf":
		p.Value = math.Inf(-1)
		return nil
	default:
		p.Value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
	}

	return nil
}

type prometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values []prometheusValue `json:"values"`
}

type prometheusData struct {
	Result []prometheusResult `json:"result"`
	ResultType string `json:"resultType"`
}

type prometheusResponse struct {
	Status string `json:"status"`
	Data prometheusData `json:"data"`
}

func (c *PrometheusGroup) convertGraphiteQueryToProm(target string) string {
	firstTag := true
	var queryBuilder strings.Builder
	tagsString := target[len("seriesByTag("):len(target)-1]
	tvs := c.splitTagValues(tagsString)
	// It's ok to have empty "__name__"
	if v, ok := tvs["__name__"]; ok {
		if v.OP == "=" {
			queryBuilder.WriteString(v.TagValue)
		} else {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString("__name__"+v.OP+"\""+v.TagValue+"\"")
		}

		delete(tvs, "__name__")
	}
	for tagName, t := range tvs {
		if firstTag {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString(tagName+t.OP+"\""+t.TagValue+"\"")
		} else {
			queryBuilder.WriteString(", " + tagName+t.OP+"\""+t.TagValue+"\"")
		}

	}
	if !firstTag {
		queryBuilder.WriteByte('}')
	}
	return queryBuilder.String()
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
	logger := c.logger.With(zap.String("type", "info"))
	stats := &types.Stats{}
	rewrite, _ := url.Parse("http://127.0.0.1/info/")

	var r protov3.ZipperInfoResponse
	var e errors.Errors
	r.Info = make(map[string]protov3.MultiMetricsInfoResponse)
	data := protov3.MultiMetricsInfoResponse{}
	server := c.groupName
	if len(c.servers) == 1 {
		server = c.servers[0]
	}

	for _, query := range request.Names {
		v := url.Values{
			"target": []string{query},
			"format": []string{c.protocol},
		}
		rewrite.RawQuery = v.Encode()
		res, e2 := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
		if e2 != nil {
			e.Merge(e2)
			continue
		}

		var info protov2.InfoResponse
		err := info.Unmarshal(res.Response)
		if err != nil {
			e.Add(err)
			continue
		}
		stats.Servers = append(stats.Servers, res.Server)

		if info.AggregationMethod == "" {
			info.AggregationMethod = "average"
		}
		infoV3 := protov3.MetricsInfoResponse{
			Name:              info.Name,
			ConsolidationFunc: info.AggregationMethod,
			XFilesFactor:      info.XFilesFactor,
			MaxRetention:      int64(info.MaxRetention),
		}

		for _, r := range info.Retentions {
			newR := protov3.Retention{
				SecondsPerPoint: int64(r.SecondsPerPoint),
				NumberOfPoints:  int64(r.NumberOfPoints),
			}
			infoV3.Retentions = append(infoV3.Retentions, newR)
		}

		data.Metrics = append(data.Metrics, infoV3)
	}
	r.Info[server] = data

	if len(e.Errors) != 0 {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e.Errors),
		)
	}

	if len(r.Info[server].Metrics) == 0 {
		return nil, stats, errors.FromErr(types.ErrNoResponseFetched)
	}

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	return &r, stats, nil
}

func (c *PrometheusGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}
func (c *PrometheusGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.FromErr(types.ErrNotImplementedYet)
}

func (c *PrometheusGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64) ([]string, *errors.Errors) {
	logger := c.logger
	var rewrite *url.URL
	if isTagName {
		logger = logger.With(zap.String("type", "tagName"))
		rewrite, _ = url.Parse("http://127.0.0.1/tags/autoComplete/tags")
	} else {
		logger = logger.With(zap.String("type", "tagValues"))
		rewrite, _ = url.Parse("http://127.0.0.1/tags/autoComplete/values")
	}

	var r []string

	rewrite.RawQuery = query
	res, e := c.httpQuery.DoQuery(ctx, rewrite.RequestURI(), nil)
	if e != nil {
		return r, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		e.Add(err)
		return r, e
	}

	logger.Debug("got client response",
		zap.Any("r", r),
	)

	return r, nil
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
