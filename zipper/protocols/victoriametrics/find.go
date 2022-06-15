package victoriametrics

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/grafana/carbonapi/zipper/types"
)

func (c *VictoriaMetricsGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	supportedFeatures, _ := c.featureSet.Load().(*vmSupportedFeatures)
	if !supportedFeatures.SupportGraphiteFindAPI {
		// VictoriaMetrics <1.41.0 doesn't support graphite find api, reverting back to prometheus code-path
		return c.BackendServer.Find(ctx, request)
	}
	var r protov3.MultiGlobResponse
	var e merry.Error

	logger := c.logger.With(
		zap.String("type", "find"),
		zap.Strings("request", request.Metrics),
		zap.Any("supported_features", supportedFeatures),
	)
	stats := &types.Stats{}
	var serverUrl string
	if len(c.vmClusterTenantID) > 0 {
		serverUrl = fmt.Sprintf("http://127.0.0.1/select/%s/graphite/metrics/find", c.vmClusterTenantID)
	} else {
		serverUrl = "http://127.0.0.1/metrics/find"
	}
	rewrite, _ := url.Parse(serverUrl)

	r.Metrics = make([]protov3.GlobResponse, 0)
	parser := c.parserPool.Get()
	defer c.parserPool.Put(parser)

	for _, query := range request.Metrics {
		v := url.Values{
			"query": []string{query},
		}

		rewrite.RawQuery = v.Encode()
		stats.FindRequests++
		res, queryErr := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
		if queryErr != nil {
			stats.FindErrors++
			if merry.Is(queryErr, types.ErrTimeoutExceeded) {
				stats.Timeouts++
				stats.FindTimeouts++
			}
			if e == nil {
				e = merry.Wrap(queryErr).WithValue("query", query)
			} else {
				e = e.WithCause(queryErr)
			}
			continue
		}
		parsedJSON, err := parser.ParseBytes(res.Response)
		if err != nil {
			if e == nil {
				e = merry.Wrap(err).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		globs, err := parsedJSON.Array()
		if err != nil {
			if e == nil {
				e = merry.Wrap(err).WithValue("query", query)
			} else {
				e = e.WithCause(err)
			}
			continue
		}

		stats.Servers = append(stats.Servers, res.Server)
		matches := make([]protov3.GlobMatch, 0, len(globs))
		for _, m := range globs {
			p := m.GetObject()
			isLeaf := p.Get("leaf").GetInt() == 1
			path := string(p.Get("id").GetStringBytes())
			if len(path) > 1 && path[len(path)-1] == '.' {
				path = path[0 : len(path)-1]
			}
			matches = append(matches, protov3.GlobMatch{
				Path:   path,
				IsLeaf: isLeaf,
			})
		}
		r.Metrics = append(r.Metrics, protov3.GlobResponse{
			Name:    query,
			Matches: matches,
		})
	}

	if e != nil {
		logger.Error("errors occurred while getting results",
			zap.Any("errors", e),
		)
		return &r, stats, e
	}
	return &r, stats, nil
}

func (c *VictoriaMetricsGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
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
