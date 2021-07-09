package gosnowth

import (
	"context"
	"fmt"
	"path"
)

// LocateMetric returns a list of nodes owning the specified metric
func (sc *SnowthClient) LocateMetric(uuid string, metric string,
	node ...*SnowthNode) ([]TopologyNode, error) {
	if len(node) > 0 {
		return sc.LocateMetricRemote(uuid, metric, node[0])
	}

	topo, err := sc.Topology()
	if err != nil {
		return nil, err
	}

	return topo.FindMetric(uuid, metric)
}

// LocateMetricContext is the context aware version of LocateMetric
func (sc *SnowthClient) LocateMetricContext(ctx context.Context, uuid string,
	metric string, node ...*SnowthNode) ([]TopologyNode, error) {
	if len(node) > 0 {
		return sc.LocateMetricRemoteContext(ctx, uuid, metric, node[0])
	}

	topo, err := sc.Topology()
	if err != nil {
		return nil, err
	}

	return topo.FindMetric(uuid, metric)
}

// LocateMetricRemote locates which nodes contain specified metric data.
func (sc *SnowthClient) LocateMetricRemote(uuid string, metric string,
	node *SnowthNode) ([]TopologyNode, error) {
	return sc.LocateMetricRemoteContext(context.Background(),
		uuid, metric, node)
}

// LocateMetricRemoteContext is the context aware version of LocateMetricRemote.
func (sc *SnowthClient) LocateMetricRemoteContext(ctx context.Context,
	uuid string, metric string, node *SnowthNode) ([]TopologyNode, error) {
	r := &Topology{}
	if node == nil {
		nodes := sc.ListActiveNodes()
		if len(nodes) == 0 {
			return nil, fmt.Errorf("no active nodes")
		}

		node = nodes[0]
	}

	body, _, err := sc.DoRequestContext(ctx, node, "GET",
		path.Join("/locate/xml", uuid, metric), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeXML(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	if r.WriteCopies == 0 {
		r.WriteCopies = r.OldWriteCopies
	}

	r.OldWriteCopies = r.WriteCopies

	return r.Nodes, nil
}
