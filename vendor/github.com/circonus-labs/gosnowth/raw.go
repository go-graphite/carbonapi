package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/circonus-labs/gosnowth/fb/noit"
	flatbuffers "github.com/google/flatbuffers/go"
)

// MetriclistFlatbufferContentType is the content type header for flatbuffer
// raw data.
const MetriclistFlatbufferContentType = "application/x-circonus-metric-list-flatbuffer"

// RawNumericValueResponse values represent raw numeric data responses
// from IRONdb.
type RawNumericValueResponse struct {
	Data []RawNumericValue
}

// UnmarshalJSON decodes a JSON format byte slice into a
// RawNumericValueResponse.
func (rv *RawNumericValueResponse) UnmarshalJSON(b []byte) error {
	rv.Data = []RawNumericValue{}
	values := [][]interface{}{}

	if err := json.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("failed to deserialize raw numeric response %w", err)
	}

	for _, entry := range values {
		rnv := RawNumericValue{}
		if m, ok := entry[1].(float64); ok {
			rnv.Value = m
		}

		// grab the timestamp
		if v, ok := entry[0].(float64); ok {
			rnv.Time = time.Unix(int64(v/1000), 0)
		}

		rv.Data = append(rv.Data, rnv)
	}

	return nil
}

// RawNumericValue values represent raw numeric data.
type RawNumericValue struct {
	Time  time.Time
	Value float64
}

// ReadRawNumericValues reads raw numeric data from a node.
func (sc *SnowthClient) ReadRawNumericValues(start time.Time, end time.Time,
	uuid string, metric string,
	nodes ...*SnowthNode,
) ([]RawNumericValue, error) {
	return sc.ReadRawNumericValuesContext(context.Background(), start, end,
		uuid, metric, nodes...)
}

// ReadRawNumericValuesContext is the context aware version of
// ReadRawNumericValues.
func (sc *SnowthClient) ReadRawNumericValuesContext(ctx context.Context,
	start, end time.Time, uuid, metric string,
	nodes ...*SnowthNode,
) ([]RawNumericValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(uuid, metric))
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	qp := url.Values{}
	qp.Add("start_ts", formatTimestamp(start))
	qp.Add("end_ts", formatTimestamp(end))

	r := &RawNumericValueResponse{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", path.Join("/raw",
		uuid, metric)+"?"+qp.Encode(), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r.Data, nil
}

// WriteRaw writes raw IRONdb data to a node.
func (sc *SnowthClient) WriteRaw(data io.Reader,
	fb bool, dataPoints uint64,
	nodes ...*SnowthNode,
) (*IRONdbPutResponse, error) {
	return sc.WriteRawContext(context.Background(), data, fb, dataPoints,
		nodes...)
}

// WriteRawContext is the context aware version of WriteRaw.
func (sc *SnowthClient) WriteRawContext(ctx context.Context,
	data io.Reader, fb bool, dataPoints uint64,
	nodes ...*SnowthNode,
) (*IRONdbPutResponse, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	hdrs := http.Header{
		"X-Snowth-Datapoints": {strconv.FormatUint(dataPoints, 10)},
	}

	if fb { // is flatbuffer?
		hdrs["Content-Type"] = []string{MetriclistFlatbufferContentType}
	}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", "/raw", data, hdrs)
	if err != nil {
		return nil, err
	}

	r := &IRONdbPutResponse{}
	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// WriteRawMetricList writes raw IRONdb data to a node with FlatBuffers.
func (sc *SnowthClient) WriteRawMetricList(metricList *noit.MetricListT,
	builder *flatbuffers.Builder,
	nodes ...*SnowthNode,
) (*IRONdbPutResponse, error) {
	return sc.WriteRawMetricListContext(context.Background(),
		metricList, builder, nodes...)
}

// WriteRawMetricListContext is the context aware version of WriteRawMetricList.
func (sc *SnowthClient) WriteRawMetricListContext(ctx context.Context,
	metricList *noit.MetricListT, builder *flatbuffers.Builder,
	nodes ...*SnowthNode,
) (*IRONdbPutResponse, error) {
	if metricList == nil {
		return nil, fmt.Errorf("metric list cannot be nil")
	}

	datapoints := uint64(len(metricList.Metrics))
	if datapoints == 0 {
		return nil, fmt.Errorf("metric list cannot be empty")
	}

	if builder == nil {
		builder = flatbuffers.NewBuilder(1024)
	} else {
		builder.Reset()
	}

	offset := noit.MetricListPack(builder, metricList)
	builder.FinishWithFileIdentifier(offset, []byte("CIML"))
	reader := bytes.NewReader(builder.FinishedBytes())

	return sc.WriteRawContext(ctx, reader, true, datapoints, nodes...)
}
