package gosnowth

import (
	"context"
	"fmt"
)

// GetStats retrieves the metrics about the status of an IRONdb node.
func (sc *SnowthClient) GetStats(nodes ...*SnowthNode) (*Stats, error) {
	return sc.GetStatsContext(context.Background(), nodes...)
}

// GetStatsContext is the context aware version of GetStats.
func (sc *SnowthClient) GetStatsContext(ctx context.Context,
	nodes ...*SnowthNode,
) (*Stats, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	r := &Stats{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", "/stats.json", nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// GetStatsNodeContext gets the status metrics for a single specified node.
func (sc *SnowthClient) GetStatsNodeContext(ctx context.Context,
	node *SnowthNode,
) (*Stats, error) {
	r := &Stats{}

	body, _, _, err := sc.do(ctx, node, "GET", "/stats.json", nil, nil, "")
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// Stats values represent a collection of metric data describing the status
// of an IRONdb node.
type Stats map[string]interface{}

// Identity returns the identity string value from a node state value.
func (ns *Stats) Identity() string {
	if ns == nil {
		return ""
	}

	m, ok := (*ns)["identity"].(map[string]interface{})
	if !ok {
		return ""
	}

	id, ok := m["_value"].(string)
	if !ok {
		return ""
	}

	return id
}

// SemVer returns the semantic version string value from a node state value.
func (ns *Stats) SemVer() string {
	if ns == nil {
		return ""
	}

	m, ok := (*ns)["semver"].(map[string]interface{})
	if !ok {
		return ""
	}

	ver, ok := m["_value"].(string)
	if !ok {
		return ""
	}

	return ver
}

// CurrentTopology returns the current topology string value from a node state
// value.
func (ns *Stats) CurrentTopology() string {
	if ns == nil {
		return ""
	}

	t, ok := (*ns)["topology"].(map[string]interface{})
	if !ok {
		return ""
	}

	m, ok := t["current"].(map[string]interface{})
	if !ok {
		return ""
	}

	current, ok := m["_value"].(string)
	if !ok {
		return ""
	}

	return current
}

// NextTopology returns the next topology string value from a node state value.
func (ns *Stats) NextTopology() string {
	if ns == nil {
		return ""
	}

	t, ok := (*ns)["topology"].(map[string]interface{})
	if !ok {
		return ""
	}

	m, ok := t["next"].(map[string]interface{})
	if !ok {
		return ""
	}

	next, ok := m["_value"].(string)
	if !ok {
		return ""
	}

	return next
}
