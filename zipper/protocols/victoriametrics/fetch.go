package victoriametrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus/helpers"
	prometheusTypes "github.com/go-graphite/carbonapi/zipper/protocols/prometheus/types"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"
)

type fetchTarget struct {
	name  string
	start int64
	stop  int64
	step  string
}

func (c *VictoriaMetricsGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	supportedFeatures, _ := c.featureSet.Load().(*vmSupportedFeatures)
	if !supportedFeatures.SupportOptimizedGraphiteFetch {
		// VictoriaMetrics <1.53.1 doesn't support graphite find api, reverting back to prometheus code-path
		return c.BackendServer.Fetch(ctx, request)
	}
	logger := c.logger.With(zap.String("type", "fetch"), zap.String("request", request.String()))
	stats := &types.Stats{}
	var serverUrl string
	if len(c.vmClusterTenantID) > 0 {
		serverUrl = fmt.Sprintf("http://127.0.0.1/select/%s/prometheus/api/v1/query_range", c.vmClusterTenantID)
	} else {
		serverUrl = "http://127.0.0.1/api/v1/query_range"
	}
	rewrite, _ := url.Parse(serverUrl)

	pathExprToTargets := make(map[string][]*fetchTarget)
	for _, m := range request.Metrics {
		var maxPointsPerQuery int64
		if m.MaxDataPoints != 0 {
			maxPointsPerQuery = m.MaxDataPoints
		} else {
			maxPointsPerQuery = c.maxPointsPerQuery
		}

		step := helpers.AdjustStep(m.StartTime, m.StopTime, maxPointsPerQuery, c.step, c.forceMinStepInterval)
		stepStr := strconv.FormatInt(step, 10)

		t := &fetchTarget{
			name:  m.Name,
			start: m.StartTime,
			stop:  m.StopTime,
			step:  stepStr,
		}
		targets := pathExprToTargets[m.PathExpression]
		pathExprToTargets[m.PathExpression] = append(targets, t)
	}

	var r protov3.MultiFetchResponse
	var e merry.Error

	for pathExpr, targets := range pathExprToTargets {
		for _, target := range targets {
			logger.Debug("got some target to query",
				zap.Any("pathExpr", pathExpr),
				zap.Any("target", target.name),
			)
			// rewrite metric for Tag
			// Make local copy
			stepLocalStr := target.step
			if strings.HasPrefix(target.name, "seriesByTag") {
				target.name = strings.ReplaceAll(target.name, "'name=", "'__name__=")
				stepLocalStr, target.name = helpers.SeriesByTagToPromQL(stepLocalStr, target.name)
			} else {
				target.name = fmt.Sprintf("{__graphite__=%q}", target.name)
			}
			if stepLocalStr[len(stepLocalStr)-1] >= '0' && stepLocalStr[len(stepLocalStr)-1] <= '9' {
				stepLocalStr += "s"
			}
			t, err := time.ParseDuration(stepLocalStr)
			if err != nil {
				stats.RenderErrors += 1
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

			logger.Debug("will do query",
				zap.String("query", target.name),
				zap.Int64("start", target.start),
				zap.Int64("stop", target.stop),
				zap.String("step", stepLocalStr),
				zap.String("max_lookback", stepLocalStr),
			)
			v := url.Values{
				"query":        []string{target.name},
				"start":        []string{strconv.Itoa(int(target.start))},
				"end":          []string{strconv.Itoa(int(target.stop))},
				"step":         []string{stepLocalStr},
				"max_lookback": []string{stepLocalStr},
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
				stats.RenderErrors += 1
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
				stats.RenderErrors += 1
				if e == nil {
					e = types.ErrFailedToFetch.WithMessage(response.Status).WithValue("query", target.name).WithValue("status", response.Status)
				} else {
					e = e.WithCause(err2).WithValue("query", target.name).WithValue("status", response.Status)
				}
				continue
			}

			for _, m := range response.Data.Result {
				// We always should trust backend's response (to mimic behavior of graphite for grahpite native protoocols)
				// See https://github.com/go-graphite/carbonapi/issues/504 and https://github.com/go-graphite/carbonapi/issues/514
				realStart := target.start
				realStop := target.stop
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
					RequestStartTime:  target.start,
					RequestStopTime:   target.stop,
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
