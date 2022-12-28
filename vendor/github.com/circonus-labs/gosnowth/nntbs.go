package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/circonus-labs/gosnowth/fb/nntbs"
	flatbuffers "github.com/google/flatbuffers/go"
)

const metricSourceGraphite = 0x2

func nntMergeFileIdentifier() []byte {
	return []byte("CINN")
}

// WriteNNTBSFlatbuffer writes flatbuffer format NNTBS data to an IRONdb node.
func (sc *SnowthClient) WriteNNTBSFlatbuffer(merge *nntbs.NNTMergeT,
	builder *flatbuffers.Builder, nodes ...*SnowthNode,
) error {
	return sc.WriteNNTBSFlatbufferContext(context.Background(), merge,
		builder, nodes...)
}

// WriteNNTBSFlatbufferContext is the context aware version of
// WriteNNTBSFlatbuffer.
func (sc *SnowthClient) WriteNNTBSFlatbufferContext(ctx context.Context,
	merge *nntbs.NNTMergeT, builder *flatbuffers.Builder,
	nodes ...*SnowthNode,
) error {
	if merge == nil {
		return fmt.Errorf("NNTBS merge data must not be null")
	}

	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else if len(merge.Ops) > 0 {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(
			string(merge.Ops[0].Metric.MetricLocator.CheckUuid),
			merge.Ops[0].Metric.MetricLocator.MetricName))
	}

	if builder == nil {
		builder = flatbuffers.NewBuilder(1024)
	} else {
		builder.Reset()
	}

	offset := nntbs.NNTMergePack(builder, merge)
	builder.FinishWithFileIdentifier(offset, nntMergeFileIdentifier())

	data := builder.FinishedBytes()
	hdrs := http.Header{"Content-Type": {"application/snowth-nntbs"}}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", "/nntbs",
		bytes.NewReader(data), hdrs)
	if err != nil {
		return err
	}

	res := &IRONdbPutResponse{}
	if err := json.NewDecoder(body).Decode(res); err != nil {
		return fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	if res.Errors != 0 || res.Misdirected != 0 || res.Records != 1 ||
		res.Updated != 1 {
		return fmt.Errorf("failed to write nntbs data: %v", res)
	}

	return nil
}
