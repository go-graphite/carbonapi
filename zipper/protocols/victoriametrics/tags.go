package victoriametrics

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
)

func (c *VictoriaMetricsGroup) doTagQuery(ctx context.Context, isTagName bool, query string, limit int64, supportedFeatures *vmSupportedFeatures) ([]string, merry.Error) {
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
	res, e := c.httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), nil)
	if e != nil {
		return r, e
	}

	err := json.Unmarshal(res.Response, &r)
	if err != nil {
		return r, merry.Wrap(err)
	}

	if supportedFeatures.GraphiteTagsAPIRequiresDedupe {
		// Current versions of VictoriaMetrics can return duplicate results.
		// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/942
		seen := make(map[string]struct{}, len(r))
		i := 0
		for _, v := range r {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			r[i] = v
			i++
		}
		r = r[:i]
	}

	logger.Debug("got client response",
		zap.Strings("response", r),
	)

	return r, nil
}

func (c *VictoriaMetricsGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	supportedFeatures, _ := c.featureSet.Load().(*vmSupportedFeatures)
	if !supportedFeatures.SupportGraphiteTagsAPI {
		// VictoriaMetrics < 1.47.0 doesn't support graphite tags api, reverting back to prometheus code-path
		return c.BackendServer.TagNames(ctx, query, limit)
	}
	return c.doTagQuery(ctx, true, query, limit, supportedFeatures)
}

func (c *VictoriaMetricsGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	supportedFeatures, _ := c.featureSet.Load().(*vmSupportedFeatures)
	if !supportedFeatures.SupportGraphiteTagsAPI {
		// VictoriaMetrics < 1.47.0 doesn't support graphite tags api, reverting back to prometheus code-path
		return c.BackendServer.TagValues(ctx, query, limit)
	}
	return c.doTagQuery(ctx, false, query, limit, supportedFeatures)
}
