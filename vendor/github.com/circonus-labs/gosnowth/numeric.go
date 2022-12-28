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

// NumericAllValueResponse values represent numeric data responses from IRONdb.
type NumericAllValueResponse struct {
	Data []NumericAllValue
}

// UnmarshalJSON decodes a JSON format byte slice into a
// NumericAllValueResponse.
func (nv *NumericAllValueResponse) UnmarshalJSON(b []byte) error {
	nv.Data = []NumericAllValue{}
	values := [][]interface{}{}

	if err := json.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("failed to deserialize numeric average response: %w",
			err)
	}

	for _, entry := range values {
		nav := NumericAllValue{}

		if m, ok := entry[1].(map[string]interface{}); ok {
			valueBytes, err := json.Marshal(m)
			if err != nil {
				return fmt.Errorf(
					"failed to marshal intermediate value from tuple: %w", err)
			}

			if err := json.Unmarshal(valueBytes, &nav); err != nil {
				return fmt.Errorf("failed to unmarshal value from tuple: %w",
					err)
			}
		}

		// grab the timestamp
		if v, ok := entry[0].(float64); ok {
			nav.Time = time.Unix(int64(v), 0)
		}

		nv.Data = append(nv.Data, nav)
	}

	return nil
}

// NumericAllValue values represent numeric data.
type NumericAllValue struct {
	Time              time.Time `json:"-"`
	Count             int64     `json:"count"`
	Value             int64     `json:"value"`
	StdDev            int64     `json:"stddev"`
	Derivative        int64     `json:"derivative"`
	DerivativeStdDev  int64     `json:"derivative_stddev"`
	Counter           int64     `json:"counter"`
	CounterStdDev     int64     `json:"counter_stddev"`
	Derivative2       int64     `json:"derivative2"`
	Derivative2StdDev int64     `json:"derivative2_stddev"`
	Counter2          int64     `json:"counter2"`
	Counter2StdDev    int64     `json:"counter2_stddev"`
}

// NumericValueResponse values represent responses containing numeric data.
type NumericValueResponse struct {
	Data []NumericValue
}

// UnmarshalJSON decodes a JSON format byte slice into a NumericValueResponse.
func (nv *NumericValueResponse) UnmarshalJSON(b []byte) error {
	nv.Data = []NumericValue{}
	values := [][]int64{}

	if err := json.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("failed to deserialize numeric average response: %w",
			err)
	}

	for _, tuple := range values {
		nv.Data = append(nv.Data, NumericValue{
			Time:  time.Unix(tuple[0], 0),
			Value: tuple[1],
		})
	}

	return nil
}

// NumericValue values represent individual numeric data values.
type NumericValue struct {
	Time  time.Time
	Value int64
}

// NumericWrite values represent numeric data.
type NumericWrite struct {
	Count            int64        `json:"count"`
	Value            int64        `json:"value"`
	Derivative       int64        `json:"derivative"`
	Counter          int64        `json:"counter"`
	StdDev           int64        `json:"stddev"`
	DerivativeStdDev int64        `json:"derivative_stddev"`
	CounterStdDev    int64        `json:"counter_stddev"`
	Metric           string       `json:"metric"`
	ID               string       `json:"id"`
	Offset           int64        `json:"offset"`
	Parts            NumericParts `json:"parts"`
}

// NumericPartsData values represent numeric base data parts.
type NumericPartsData struct {
	Count            int64 `json:"count"`
	Value            int64 `json:"value"`
	Derivative       int64 `json:"derivative"`
	Counter          int64 `json:"counter"`
	StdDev           int64 `json:"stddev"`
	DerivativeStdDev int64 `json:"derivative_stddev"`
	CounterStdDev    int64 `json:"counter_stddev"`
}

// NumericParts values contain the NumericWrite submission parts of an
// numeric rollup.
type NumericParts struct {
	Period int64              `json:"period"`
	Data   []NumericPartsData `json:"data"`
}

// MarshalJSON marshals a NumericParts value into a JSON format byte slice.
func (p *NumericParts) MarshalJSON() ([]byte, error) {
	tuple := []interface{}{}
	tuple = append(tuple, p.Period, p.Data)
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)

	if err := enc.Encode(tuple); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// WriteNumeric writes numeric data to a node.
func (sc *SnowthClient) WriteNumeric(data []NumericWrite,
	nodes ...*SnowthNode,
) error {
	return sc.WriteNumericContext(context.Background(), data, nodes...)
}

// WriteNumericContext is the context aware version of WriteNumeric.
func (sc *SnowthClient) WriteNumericContext(ctx context.Context,
	data []NumericWrite, nodes ...*SnowthNode,
) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("failed to encode NumericWrite for write: %w", err)
	}

	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else if len(data) > 0 {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(data[0].ID,
			data[0].Metric))
	}

	_, _, err := sc.DoRequestContext(ctx, node, "POST",
		"/write/numeric", buf, nil)

	return err
}

// ReadNumericValues reads numeric data from a node.
func (sc *SnowthClient) ReadNumericValues(start, end time.Time, period int64,
	t, id, metric string, nodes ...*SnowthNode,
) ([]NumericValue, error) {
	return sc.ReadNumericValuesContext(context.Background(), start, end,
		period, t, id, metric, nodes...)
}

// ReadNumericValuesContext is the context aware version of ReadNumericValues.
func (sc *SnowthClient) ReadNumericValuesContext(ctx context.Context,
	start, end time.Time, period int64,
	t, id, metric string, nodes ...*SnowthNode,
) ([]NumericValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(id, metric))
	}

	r := &NumericValueResponse{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", path.Join("/read",
		strconv.FormatInt(start.Unix(), 10),
		strconv.FormatInt(end.Unix(), 10),
		strconv.FormatInt(period, 10), id, t, metric), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r.Data, nil
}

// ReadNumericAllValues reads all numeric data from a node.
func (sc *SnowthClient) ReadNumericAllValues(start, end time.Time, period int64,
	id, metric string, nodes ...*SnowthNode,
) ([]NumericAllValue, error) {
	return sc.ReadNumericAllValuesContext(context.Background(), start, end,
		period, id, metric, nodes...)
}

// ReadNumericAllValuesContext is the context aware version of
// ReadNumericAllValues.
func (sc *SnowthClient) ReadNumericAllValuesContext(ctx context.Context,
	start, end time.Time, period int64,
	id, metric string, nodes ...*SnowthNode,
) ([]NumericAllValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(id, metric))
	}

	r := &NumericAllValueResponse{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", path.Join("/read",
		strconv.FormatInt(start.Unix(), 10),
		strconv.FormatInt(end.Unix(), 10),
		strconv.FormatInt(period, 10), id, "all", metric), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r.Data, nil
}
