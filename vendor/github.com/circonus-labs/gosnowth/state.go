package gosnowth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetNodeState retrieves the state of an IRONdb node.
func (sc *SnowthClient) GetNodeState(nodes ...*SnowthNode) (*NodeState, error) {
	return sc.GetNodeStateContext(context.Background(), nodes...)
}

// GetNodeStateContext is the context aware version of GetNodeState.
func (sc *SnowthClient) GetNodeStateContext(ctx context.Context,
	nodes ...*SnowthNode,
) (*NodeState, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	r := &NodeState{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", "/state", nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// NodeState values represent the state of an IRONdb node.
type NodeState struct {
	Identity      string   `json:"identity"`
	Current       string   `json:"current"`
	Next          string   `json:"next"`
	NNT           Rollup   `json:"nnt"`
	NNTBS         *Rollup  `json:"nntbs"`
	Text          Rollup   `json:"text"`
	Histogram     Rollup   `json:"histogram"`
	BaseRollup    uint64   `json:"base_rollup"`
	Rollups       []uint64 `json:"rollups"`
	NNTCacheSize  uint64   `json:"nnt_cache_size"`
	RUsageUTime   float64  `json:"rusage.utime"`
	RUsageSTime   float64  `json:"rusage.stime"`
	RUsageMaxRSS  uint64   `json:"rusage.maxrss"`
	RUsageMinFLT  uint64   `json:"rusage.minflt"`
	RUsageMajFLT  uint64   `json:"rusage.majflt"`
	RUsageNSwap   uint64   `json:"rusage.nswap"`
	RUsageInBlock uint64   `json:"rusage.inblock"`
	RUsageOuBlock uint64   `json:"rusage.oublock"`

	RUsageMsgSnd   uint64  `json:"rusage.msgsnd"`
	RUsageMsgRcv   uint64  `json:"rusage.msgrcv"`
	RUsageNSignals uint64  `json:"rusage.nsignals"`
	RUsageNvcSW    uint64  `json:"rusage.nvcsw"`
	RUsageNivcSW   uint64  `json:"rusage.nivcsw"`
	MaxPeerLag     float64 `json:"max_peer_lag"`
	AvgPeerLag     float64 `json:"avg_peer_lag"`

	Features Features `json:"features"`

	Version     string `json:"version"`
	Application string `json:"application"`
}

// Rollup values represent node state rollups.
type Rollup struct {
	RollupEntries
	RollupList []uint64      `json:"rollups"`
	Aggregate  RollupDetails `json:"aggregate"`
}

// UnmarshalJSON populates a rollup value from a JSON format byte slice.
func (r *Rollup) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if rollups, ok := m["rollups"].([]interface{}); ok {
		for _, v := range rollups {
			vf, ok := v.(float64)
			if ok {
				r.RollupList = append(r.RollupList, uint64(vf))
			}

			delete(m, "rollup")
		}
	}

	if aggregate, ok := m["aggregate"].(RollupDetails); ok {
		r.Aggregate = aggregate

		delete(m, "aggregate")
	}

	rr := make(map[string]RollupDetails)

	for k, v := range m {
		if strings.HasPrefix(k, "rollup_") {
			b, _ := json.Marshal(v)
			rd := new(RollupDetails)

			if err := json.Unmarshal(b, rd); err != nil {
				return err
			}

			rr[k] = *rd
		}
	}

	r.RollupEntries = RollupEntries(rr)

	return nil
}

// RollupEntries values contain node state rollup information.
type RollupEntries map[string]RollupDetails

// RollupDetails values represent node state rollup information.
type RollupDetails struct {
	FilesSystem   FileSystemDetails `json:"fs"`
	PutCalls      uint64            `json:"put.calls"`
	PutElapsedUS  uint64            `json:"put.elapsed_us"`
	GetCalls      uint64            `json:"get.calls"`
	GetProxyCalls uint64            `json:"get.proxy_calls"`
	GetCount      uint64            `json:"get.count"`
	GetElapsedUS  uint64            `json:"get.elapsed_us"`
	ExtendCalls   uint64            `json:"extend.calls"`
}

// FileSystemDetails values represent details about a nodes file system.
type FileSystemDetails struct {
	ID      uint64  `json:"id"`
	TotalMB float64 `json:"totalMb"`
	FreeMB  float64 `json:"availMb"`
}

// Features values represent features supported by the node.
type Features struct {
	TextStore               bool `json:"text:store"`
	HistogramStore          bool `json:"histogram:store"`
	NNTSecondOrder          bool `json:"nnt:second_order"`
	HistogramDynamicRollups bool `json:"histogram:dynamic_rollups"`
	NNTStore                bool `json:"nnt:store"`
	FeatureFlags            bool `json:"features"`
}

// UnmarshalJSON populates a features value from a JSON format byte slice.
func (f *Features) UnmarshalJSON(b []byte) error {
	f.TextStore = false
	f.HistogramStore = false
	f.NNTSecondOrder = false
	f.HistogramDynamicRollups = false
	f.NNTStore = false
	f.FeatureFlags = false

	m := make(map[string]string)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

loop:
	for k, v := range m {
		if v == "1" {
			switch k {
			case "text:store":
				f.TextStore = true

				break loop
			case "histogram:store":
				f.HistogramStore = true

				break loop
			case "nnt:second_order":
				f.NNTSecondOrder = true

				break loop
			case "histogram:dynamic_rollups":
				f.HistogramDynamicRollups = true

				break loop
			case "nnt:store":
				f.NNTStore = true

				break loop
			case "features":
				f.FeatureFlags = true

				break loop
			}
		}
	}

	return nil
}
