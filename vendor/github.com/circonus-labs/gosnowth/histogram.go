package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/openhistogram/circonusllhist"
)

// HistogramValue values are individual data points of a histogram metric.
type HistogramValue struct {
	Time   time.Time
	Period time.Duration
	Data   map[string]int64
}

// MarshalJSON encodes a HistogramValue value into a JSON format byte slice.
func (hv *HistogramValue) MarshalJSON() ([]byte, error) {
	v := []interface{}{}
	fv, err := strconv.ParseFloat(formatTimestamp(hv.Time), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid histogram value time: " +
			formatTimestamp(hv.Time))
	}

	v = append(v, fv)
	v = append(v, hv.Period.Seconds())
	v = append(v, hv.Data)
	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a HistogramValue value.
func (hv *HistogramValue) UnmarshalJSON(b []byte) error {
	v := []interface{}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if len(v) != 3 {
		return fmt.Errorf("histogram value should contain three entries: " +
			string(b))
	}

	if fv, ok := v[0].(float64); ok {
		tv, err := parseTimestamp(strconv.FormatFloat(fv, 'f', 3, 64))
		if err != nil {
			return err
		}

		hv.Time = tv
	}

	if fv, ok := v[1].(float64); ok {
		hv.Period = time.Duration(fv) * time.Second
	}

	if m, ok := v[2].(map[string]interface{}); ok {
		hv.Data = make(map[string]int64, len(m))
		for k, iv := range m {
			if fv, ok := iv.(float64); ok {
				hv.Data[k] = int64(fv)
			}
		}
	}

	return nil
}

// Timestamp returns the HistogramValue time as a string in the IRONdb
// timestamp format.
func (hv *HistogramValue) Timestamp() string {
	return formatTimestamp(hv.Time)
}

// ReadHistogramValues reads histogram data from a node.
func (sc *SnowthClient) ReadHistogramValues(
	uuid, metric string, period time.Duration,
	start, end time.Time, nodes ...*SnowthNode) ([]HistogramValue, error) {
	return sc.ReadHistogramValuesContext(context.Background(), uuid,
		metric, period, start, end, nodes...)
}

// ReadHistogramValuesContext is the context aware version of
// ReadHistogramValues.
func (sc *SnowthClient) ReadHistogramValuesContext(ctx context.Context,
	uuid, metric string, period time.Duration,
	start, end time.Time, nodes ...*SnowthNode) ([]HistogramValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(uuid, metric))
	}

	startTS := start.Unix() - start.Unix()%int64(period.Seconds())
	endTS := end.Unix() - end.Unix()%int64(period.Seconds()) +
		int64(period.Seconds())
	r := []HistogramValue{}
	body, _, err := sc.DoRequestContext(ctx, node, "GET",
		path.Join("/histogram", strconv.FormatInt(startTS, 10),
			strconv.FormatInt(endTS, 10),
			strconv.FormatInt(int64(period.Seconds()), 10), uuid,
			url.QueryEscape(metric)), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// HistogramData values represent histogram data records in IRONdb.
type HistogramData struct {
	AccountID int64                     `json:"account_id"`
	Metric    string                    `json:"metric"`
	ID        string                    `json:"id"`
	CheckName string                    `json:"check_name"`
	Offset    int64                     `json:"offset"`
	Period    int64                     `json:"period"`
	Histogram *circonusllhist.Histogram `json:"histogram"`
}

// WriteHistogram sends a list of histogram data values to be written
// to an IRONdb node.
func (sc *SnowthClient) WriteHistogram(data []HistogramData,
	nodes ...*SnowthNode) error {
	return sc.WriteHistogramContext(context.Background(), data, nodes...)
}

// WriteHistogramContext is the context aware version of WriteHistogram.
func (sc *SnowthClient) WriteHistogramContext(ctx context.Context,
	data []HistogramData, nodes ...*SnowthNode) error {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else if len(data) > 0 {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(data[0].ID,
			data[0].Metric))
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("failed to encode HistogramData for write: %w", err)
	}

	_, _, err := sc.DoRequestContext(ctx, node, "POST", "/histogram/write", buf, nil)
	return err
}
