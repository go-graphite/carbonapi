package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/circonus-labs/gosnowth/fb/fetch"
	flatbuffers "github.com/google/flatbuffers/go"
)

// FetchStream values represent queries for individual data streams in an
// IRONdb fetch request.
type FetchStream struct {
	UUID            string   `json:"uuid"`
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	Label           string   `json:"label,omitempty"`
	Transform       string   `json:"transform"`
	TransformParams []string `json:"transform_params,omitempty"`
}

// FetchReduce values represent reduce operations to perform on specified
// data streams in an IRONdb fetch request.
type FetchReduce struct {
	Label        string   `json:"label"`
	Method       string   `json:"method"`
	MethodParams []string `json:"method_params,omitempty"`
}

// FetchQuery values represent queries used to fetch IRONdb data.
type FetchQuery struct {
	Start   time.Time     `json:"start"`
	Period  time.Duration `json:"period"`
	Count   int64         `json:"count"`
	Streams []FetchStream `json:"streams"`
	Reduce  []FetchReduce `json:"reduce"`
}

// MarshalJSON encodes a FetchQuery value into a JSON format byte slice.
func (fq *FetchQuery) MarshalJSON() ([]byte, error) {
	v := struct {
		Start   float64       `json:"start"`
		Period  float64       `json:"period"`
		Count   int64         `json:"count"`
		Streams []FetchStream `json:"streams"`
		Reduce  []FetchReduce `json:"reduce"`
	}{}

	fv, err := strconv.ParseFloat(formatTimestamp(fq.Start), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid fetch start value: " +
			formatTimestamp(fq.Start))
	}

	v.Start = fv
	v.Period = fq.Period.Seconds()
	v.Count = fq.Count

	if len(fq.Streams) > 0 {
		v.Streams = fq.Streams
	}

	if len(fq.Reduce) > 0 {
		v.Reduce = fq.Reduce
	}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a HistogramValue value.
func (fq *FetchQuery) UnmarshalJSON(b []byte) error {
	v := struct {
		Start   float64       `json:"start"`
		Period  float64       `json:"period"`
		Count   int64         `json:"count"`
		Streams []FetchStream `json:"streams"`
		Reduce  []FetchReduce `json:"reduce"`
	}{}

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v.Start == 0 {
		return fmt.Errorf("fetch query missing start: " + string(b))
	}

	fq.Start, err = parseTimestamp(strconv.FormatFloat(v.Start, 'f', 3, 64))
	if err != nil {
		return err
	}

	if v.Period == 0 {
		return fmt.Errorf("fetch query missing period: " + string(b))
	}

	fq.Period = time.Duration(v.Period*1000) * time.Millisecond

	if v.Count == 0 {
		return fmt.Errorf("fetch query missing count: " + string(b))
	}

	fq.Count = v.Count

	if len(v.Streams) < 1 {
		return fmt.Errorf("fetch query requires at least one stream: " +
			string(b))
	}

	fq.Streams = v.Streams

	if len(v.Reduce) < 1 {
		return fmt.Errorf("fetch query requires at least one reduce: " +
			string(b))
	}

	fq.Reduce = v.Reduce

	return nil
}

// Timestamp returns the FetchQuery start time as a string in the IRONdb
// timestamp format.
func (fq *FetchQuery) Timestamp() string {
	return formatTimestamp(fq.Start)
}

// FetchValues retrieves data values using the IRONdb fetch API.
func (sc *SnowthClient) FetchValues(q *FetchQuery, nodes ...*SnowthNode) (*DF4Response, error) {
	return sc.FetchValuesContext(context.Background(), q, nodes...)
}

// FetchValuesContext is the context aware version of FetchValues.
func (sc *SnowthClient) FetchValuesContext(ctx context.Context,
	q *FetchQuery, nodes ...*SnowthNode,
) (*DF4Response, error) {
	var node *SnowthNode

	switch {
	case len(nodes) > 0 && nodes[0] != nil:
		node = nodes[0]
	case len(q.Streams) > 0:
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(q.Streams[0].UUID, q.Streams[0].Name))
	default:
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&q); err != nil {
		return nil, err
	}

	hdrs := http.Header{"Content-Type": {"application/json"}}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", "/fetch", buf, hdrs)
	if err != nil {
		return nil, err
	}

	rb, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("unable to read IRONdb response body: %w", err)
	}

	rb = replaceInf(rb)

	r := &DF4Response{}
	if err := decodeJSON(bytes.NewBuffer(rb), &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// FetchFlatbufferContentType is the content type header for flatbuffer fetch data.
const FetchFlatbufferContentType = "x-irondb-fetch-flatbuffer"

// Df4FlatbufferAccept is the accept header for flatbuffer df4 data.
const Df4FlatbufferAccept = "x-irondb-df4-flatbuffer"

// FetchValuesFb retrieves data values using the IRONdb fetch API with FlatBuffers.
func (sc *SnowthClient) FetchValuesFb(node *SnowthNode,
	q *fetch.FetchT,
) (*fetch.DF4T, error) {
	return sc.FetchValuesFbContext(context.Background(), node, q)
}

// FetchValuesFbContext is the context aware version of FetchValuesFb.
func (sc *SnowthClient) FetchValuesFbContext(ctx context.Context,
	node *SnowthNode, q *fetch.FetchT,
) (*fetch.DF4T, error) {
	builder := flatbuffers.NewBuilder(8192)
	qOffset := fetch.FetchPack(builder, q)
	builder.Finish(qOffset)
	buf := bytes.NewBuffer(builder.FinishedBytes())

	hdrs := http.Header{
		"Content-Type": {FetchFlatbufferContentType},
		"Accept":       {Df4FlatbufferAccept},
	}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", "/fetch", buf, hdrs)
	if err != nil {
		return nil, err
	}

	df4Buf, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	df4 := fetch.GetRootAsDF4(df4Buf, flatbuffers.UOffsetT(0))
	r := df4.UnPack()

	return r, nil
}
