package gosnowth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"
)

// RollupValue values are individual data points of a rollup.
type RollupValue struct {
	Time  time.Time
	Value *float64
}

// MarshalJSON encodes a RollupValue value into a JSON format byte slice.
func (rv *RollupValue) MarshalJSON() ([]byte, error) {
	v := []interface{}{}
	fv, err := strconv.ParseFloat(formatTimestamp(rv.Time), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid rollup value time: " +
			formatTimestamp(rv.Time))
	}

	v = append(v, fv)
	v = append(v, rv.Value)
	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a RollupValue value.
func (rv *RollupValue) UnmarshalJSON(b []byte) error {
	v := []interface{}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if len(v) != 2 {
		return fmt.Errorf("rollup value should contain two entries: " +
			string(b))
	}

	if fv, ok := v[0].(float64); ok {
		tv, err := parseTimestamp(strconv.FormatFloat(fv, 'f', 3, 64))
		if err != nil {
			return err
		}

		rv.Time = tv
	}

	if fv, ok := v[1].(float64); ok {
		rv.Value = new(float64)
		*rv.Value = fv
	}

	return nil
}

// Timestamp returns the RollupValue time as a string in the IRONdb timestamp
// format.
func (rv *RollupValue) Timestamp() string {
	return formatTimestamp(rv.Time)
}

// RollupAllData values contain the data values of an individual rollup data
// point.
type RollupAllData struct {
	Count             int64
	Counter           float64
	Counter2          float64
	CounterStddev     float64
	Counter2Stddev    float64
	Derivative        float64
	Derivative2       float64
	DerivativeStddev  float64
	Derivative2Stddev float64
	Stddev            float64
	Value             float64
}

// RollupAllValue values contain all parts of an individual rollup data point.
type RollupAllValue struct {
	Time time.Time
	Data *RollupAllData
}

// MarshalJSON encodes a RollupValue value into a JSON format byte slice.
func (rv *RollupAllValue) MarshalJSON() ([]byte, error) {
	v := []interface{}{}
	fv, err := strconv.ParseFloat(formatTimestamp(rv.Time), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid rollup value time: " +
			formatTimestamp(rv.Time))
	}

	v = append(v, fv)
	if rv.Data == nil {
		v = append(v, nil)
	} else {
		v = append(v, map[string]interface{}{
			"count":              rv.Data.Count,
			"value":              rv.Data.Value,
			"stddev":             rv.Data.Stddev,
			"derivative":         rv.Data.Derivative,
			"derivative_stddev":  rv.Data.DerivativeStddev,
			"counter":            rv.Data.Counter,
			"counter_stddev":     rv.Data.CounterStddev,
			"derivative2":        rv.Data.Derivative2,
			"derivative2_stddev": rv.Data.Derivative2Stddev,
			"counter2":           rv.Data.Counter2,
			"counter2_stddev":    rv.Data.Counter2Stddev,
		})
	}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a RollupValue value.
func (rv *RollupAllValue) UnmarshalJSON(b []byte) error {
	v := []interface{}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if len(v) != 2 {
		return fmt.Errorf("rollup value should contain two entries: " +
			string(b))
	}

	if fv, ok := v[0].(float64); ok {
		tv, err := parseTimestamp(strconv.FormatFloat(fv, 'f', 3, 64))
		if err != nil {
			return err
		}

		rv.Time = tv
	}

	if m, ok := v[1].(map[string]interface{}); ok {
		rv.Data = &RollupAllData{}
		for key, val := range m {
			if fv, ok := val.(float64); ok {
				switch key {
				case "count":
					rv.Data.Count = int64(fv)
				case "value":
					rv.Data.Value = fv
				case "stddev":
					rv.Data.Stddev = fv
				case "derivative":
					rv.Data.Derivative = fv
				case "derivative_stddev":
					rv.Data.DerivativeStddev = fv
				case "counter":
					rv.Data.Counter = fv
				case "counter_stddev":
					rv.Data.CounterStddev = fv
				case "derivative2":
					rv.Data.Derivative2 = fv
				case "derivative2_stddev":
					rv.Data.Derivative2Stddev = fv
				case "counter2":
					rv.Data.Counter2 = fv
				case "counter2_stddev":
					rv.Data.Counter2Stddev = fv
				}
			}
		}
	}

	return nil
}

// Timestamp returns the RollupAllValue time as a string in the IRONdb
// timestamp format.
func (rv *RollupAllValue) Timestamp() string {
	return formatTimestamp(rv.Time)
}

// ReadRollupValues reads rollup data from a node.
func (sc *SnowthClient) ReadRollupValues(uuid, metric string, period time.Duration,
	start, end time.Time, dataType string, nodes ...*SnowthNode) ([]RollupValue, error) {
	return sc.ReadRollupValuesContext(context.Background(), uuid, metric,
		period, start, end, dataType, nodes...)
}

// ReadRollupValuesContext is the context aware version of ReadRollupValues.
func (sc *SnowthClient) ReadRollupValuesContext(ctx context.Context,
	uuid, metric string, period time.Duration, start, end time.Time,
	dataType string, nodes ...*SnowthNode) ([]RollupValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(uuid, metric))
	}

	if dataType == "" {
		dataType = "average"
	}

	switch dataType {
	case "count", "average", "derive", "counter", "average_stddev",
		"derive_stddev", "counter_stddev", "derive2", "counter2",
		"derive2_stddev", "counter2_stddev":
	default:
		return nil, fmt.Errorf("invalid rollup data type: " + dataType)
	}

	startTS := start.Unix() - start.Unix()%int64(period/time.Second)
	endTS := end.Unix() - end.Unix()%int64(period/time.Second) +
		int64(period/time.Second)
	r := []RollupValue{}
	body, _, err := sc.DoRequestContext(ctx, node, "GET",
		fmt.Sprintf("%s?start_ts=%d&end_ts=%d&rollup_span=%ds&type=%s",
			path.Join("/rollup", uuid, url.QueryEscape(metric)),
			startTS, endTS, int64(period/time.Second), dataType), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// ReadRollupAllValues reads rollup data from a node.
func (sc *SnowthClient) ReadRollupAllValues(
	uuid, metric string, period time.Duration,
	start, end time.Time, nodes ...*SnowthNode) ([]RollupAllValue, error) {
	return sc.ReadRollupAllValuesContext(context.Background(), uuid,
		metric, period, start, end, nodes...)
}

// ReadRollupAllValuesContext is the context aware version of ReadRollupValues.
func (sc *SnowthClient) ReadRollupAllValuesContext(ctx context.Context,
	uuid, metric string, period time.Duration,
	start, end time.Time, nodes ...*SnowthNode) ([]RollupAllValue, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode(sc.FindMetricNodeIDs(uuid, metric))
	}

	startTS := start.Unix() - start.Unix()%int64(period/time.Second)
	endTS := end.Unix() - end.Unix()%int64(period/time.Second) +
		int64(period/time.Second)
	r := []RollupAllValue{}
	body, _, err := sc.DoRequestContext(ctx, node, "GET",
		fmt.Sprintf("%s?start_ts=%d&end_ts=%d&rollup_span=%ds&type=all",
			path.Join("/rollup", uuid, url.QueryEscape(metric)),
			startTS, endTS, int64(period/time.Second)), nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}
