package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"
)

// TextValueResponse values represent text data responses.
type TextValueResponse []TextValue

// UnmarshalJSON decodes a JSON format byte slice into a TextValueResponse.
func (tvr *TextValueResponse) UnmarshalJSON(b []byte) error {
	*tvr = TextValueResponse{}
	values := [][]interface{}{}

	if err := json.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	for _, entry := range values {
		tv := TextValue{}
		if v, ok := entry[0].(float64); ok {
			tv.Time = time.Unix(int64(v), 0)
		}

		if v, ok := entry[1].(string); ok {
			tv.Value = new(string)
			*tv.Value = v
		}

		*tvr = append(*tvr, tv)
	}

	return nil
}

// TextValue values represent text data read from IRONdb.
type TextValue struct {
	Time  time.Time
	Value *string
}

// ReadTextValues reads text data values from an IRONdb node.
func (sc *SnowthClient) ReadTextValues(uuid, metric string,
	start, end time.Time, nodes ...*SnowthNode,
) ([]TextValue, error) {
	return sc.ReadTextValuesContext(context.Background(), uuid, metric,
		start, end, nodes...)
}

// ReadTextValuesContext is the context aware version of ReadTextValues.
func (sc *SnowthClient) ReadTextValuesContext(ctx context.Context,
	uuid, metric string, start, end time.Time,
	nodes ...*SnowthNode,
) ([]TextValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(uuid, metric))
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	r := TextValueResponse{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", path.Join("/read",
		strconv.FormatInt(start.Unix(), 10),
		strconv.FormatInt(end.Unix(), 10), uuid, metric), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// TextData values represent text data to be written to IRONdb.
type TextData struct {
	Metric string `json:"metric"`
	ID     string `json:"id"`
	Offset string `json:"offset"`
	Value  string `json:"value"`
}

// WriteText writes text data to an IRONdb node.
func (sc *SnowthClient) WriteText(data []TextData, nodes ...*SnowthNode) error {
	return sc.WriteTextContext(context.Background(), data, nodes...)
}

// WriteTextContext is the context aware version of WriteText.
func (sc *SnowthClient) WriteTextContext(ctx context.Context,
	data []TextData, nodes ...*SnowthNode,
) error {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else if len(data) > 0 {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(data[0].ID,
			data[0].Metric))
	}

	if node == nil {
		return fmt.Errorf("unable to get active node")
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("failed to encode TextData for write: %w", err)
	}

	_, _, err := sc.DoRequestContext(ctx, node, "POST", "/write/text", buf, nil)

	return err
}
