package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// GraphiteOptions values contain optional parameters to be passed to the
// IRONdb graphite API call operations.
type GraphiteOptions struct {
	Limit int64 `json:"limit"`
}

// GraphiteMetric values are used to describe metrics in graphite format.
type GraphiteMetric struct {
	Leaf bool              `json:"leaf"`
	Name string            `json:"name"`
	Data map[string]string `json:"leaf_data,omitempty"`
}

// GraphiteLookup values are used to provide information needed to retrieve
// graphite datapoints.
type GraphiteLookup struct {
	Start int64    `json:"start"`
	End   int64    `json:"end"`
	Names []string `json:"names"`
}

// GraphiteDatapoints values are used to represent series of graphite
// datapoints.
type GraphiteDatapoints struct {
	From   int64                 `json:"from"`
	To     int64                 `json:"to"`
	Step   int64                 `json:"step"`
	Series map[string][]*float64 `json:"series"`
}

// GraphiteFindMetrics retrieves metrics that are associated with the provided
// graphite metric search query.
func (sc *SnowthClient) GraphiteFindMetrics(accountID int64,
	prefix, query string, options *GraphiteOptions,
	nodes ...*SnowthNode) ([]GraphiteMetric, error) {
	return sc.GraphiteFindMetricsContext(context.Background(), accountID,
		prefix, query, options, nodes...)
}

// GraphiteFindMetricsContext is the context aware version of
// GraphiteFindMetrics.
func (sc *SnowthClient) GraphiteFindMetricsContext(ctx context.Context,
	accountID int64, prefix, query string, options *GraphiteOptions,
	nodes ...*SnowthNode) ([]GraphiteMetric, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	u := fmt.Sprintf("%s?query=%s",
		sc.getURL(node, fmt.Sprintf("/graphite/%d/%s/metrics/find",
			accountID, url.QueryEscape(prefix))), url.QueryEscape(query))

	hdrs := http.Header{}
	if options != nil && options.Limit != 0 {
		hdrs.Set("X-Snowth-Advisory-Limit",
			strconv.FormatInt(options.Limit, 10))
	}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, hdrs)
	if err != nil {
		return nil, err
	}

	r := []GraphiteMetric{}
	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// GraphiteFindTags retrieves metrics that are associated with the provided
// graphite tag search query.
func (sc *SnowthClient) GraphiteFindTags(accountID int64,
	prefix, query string, options *GraphiteOptions,
	nodes ...*SnowthNode) ([]GraphiteMetric, error) {
	return sc.GraphiteFindTagsContext(context.Background(), accountID,
		prefix, query, options, nodes...)
}

// GraphiteFindTagsContext is the context aware version of
// GraphiteFindTags.
func (sc *SnowthClient) GraphiteFindTagsContext(ctx context.Context,
	accountID int64, prefix, query string, options *GraphiteOptions,
	nodes ...*SnowthNode) ([]GraphiteMetric, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	u := fmt.Sprintf("%s?query=%s",
		sc.getURL(node, fmt.Sprintf("/graphite/%d/%s/tags/find",
			accountID, url.QueryEscape(prefix))), url.QueryEscape(query))

	hdrs := http.Header{}
	if options != nil && options.Limit != 0 {
		hdrs.Set("X-Snowth-Advisory-Limit",
			strconv.FormatInt(options.Limit, 10))
	}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, hdrs)
	if err != nil {
		return nil, err
	}

	r := []GraphiteMetric{}
	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// GraphiteGetDatapoints retrieves graphite datapoint series for specified
// metrics for a specified time range.
func (sc *SnowthClient) GraphiteGetDatapoints(accountID int64,
	prefix string, lookup *GraphiteLookup, options *GraphiteOptions,
	nodes ...*SnowthNode) (*GraphiteDatapoints, error) {
	return sc.GraphiteGetDatapointsContext(context.Background(), accountID,
		prefix, lookup, options, nodes...)
}

// GraphiteGetDatapointsContext is the context aware version of
// GraphiteGetDatapoints.
func (sc *SnowthClient) GraphiteGetDatapointsContext(ctx context.Context,
	accountID int64, prefix string, lookup *GraphiteLookup,
	options *GraphiteOptions,
	nodes ...*SnowthNode) (*GraphiteDatapoints, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	u := sc.getURL(node, fmt.Sprintf("/graphite/%d/%s/series_multi",
		accountID, url.QueryEscape(prefix)))

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&lookup); err != nil {
		return nil, err
	}

	hdrs := http.Header{"Content-Type": {"application/json"}}
	if options != nil && options.Limit != 0 {
		hdrs.Set("X-Snowth-Advisory-Limit",
			strconv.FormatInt(options.Limit, 10))
	}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", u, buf, hdrs)
	if err != nil {
		return nil, err
	}

	r := &GraphiteDatapoints{}
	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}
