package gosnowth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FindTagsItem values represent results returned from IRONdb tag queries.
type FindTagsItem struct {
	UUID       string          `json:"uuid"`
	CheckTags  []string        `json:"check_tags,omitempty"`
	MetricName string          `json:"metric_name"`
	Type       string          `type:"type"`
	AccountID  int64           `json:"account_id"`
	Activity   [][]int64       `json:"activity,omitempty"`
	Latest     *FindTagsLatest `json:"latest,omitempty"`
}

// FindTagsResult values contain the results of a find tags request.
type FindTagsResult struct {
	Items    []FindTagsItem `json:"items,omitempty"`
	Count    int64          `json:"count"`
	Estimate bool           `json:"estimate"`
}

// FindTagsOptions values contain optional parameters to be passed to the
// IRONdb find tags call by a FindTags operation.
type FindTagsOptions struct {
	Start     time.Time
	End       time.Time
	StartStr  string `json:"activity_start_secs"`
	EndStr    string `json:"activity_end_secs"`
	Activity  int64  `json:"activity"`
	Latest    int64  `json:"latest"`
	CountOnly int64  `json:"count_only"`
	Limit     int64  `json:"limit"`
}

// FindTagsLatest values contain the most recent data values for a metric.
type FindTagsLatest struct {
	Numeric   []FindTagsLatestNumeric   `json:"numeric,omitempty"`
	Text      []FindTagsLatestText      `json:"text,omitempty"`
	Histogram []FindTagsLatestHistogram `json:"histogram,omitempty"`
}

// FindTagsLatestNumeric values contain recent metric numeric data.
type FindTagsLatestNumeric struct {
	Time  int64
	Value *float64
}

// MarshalJSON encodes a FindTagsLatestNumeric value into a JSON format byte
// slice.
func (ftl *FindTagsLatestNumeric) MarshalJSON() ([]byte, error) {
	v := []interface{}{ftl.Time, ftl.Value}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a FindTagsLatestNumeric
// value.
func (ftl *FindTagsLatestNumeric) UnmarshalJSON(b []byte) error {
	v := []interface{}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if len(v) != 2 {
		return fmt.Errorf("unable to decode latest numeric value, "+
			"invalid length: %v", string(b))
	}

	if fv, ok := v[0].(float64); ok {
		ftl.Time = int64(fv)
	} else {
		return fmt.Errorf("unable to decode latest numeric value, "+
			"invalid timestamp: %v", string(b))
	}

	if v[1] != nil {
		if fv, ok := v[1].(float64); ok {
			ftl.Value = &fv
		} else {
			return fmt.Errorf("unable to decode latest numeric value, "+
				"invalid value: %v", string(b))
		}
	}

	return nil
}

// FindTagsLatestText values contain recent metric text data.
type FindTagsLatestText struct {
	Time  int64
	Value *string
}

// MarshalJSON encodes a FindTagsLatestText value into a JSON format byte slice.
func (ftl *FindTagsLatestText) MarshalJSON() ([]byte, error) {
	v := []interface{}{ftl.Time, ftl.Value}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a FindTagsLatestText
// value.
func (ftl *FindTagsLatestText) UnmarshalJSON(b []byte) error {
	v := []interface{}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if len(v) != 2 {
		return fmt.Errorf("unable to decode latest text value, "+
			"invalid length: %v", string(b))
	}

	if fv, ok := v[0].(float64); ok {
		ftl.Time = int64(fv)
	} else {
		return fmt.Errorf("unable to decode latest text value, "+
			"invalid timestamp: %v", string(b))
	}

	if v[1] != nil {
		if sv, ok := v[1].(string); ok {
			ftl.Value = &sv
		} else {
			return fmt.Errorf("unable to decode latest text value, "+
				"invalid value: %v", string(b))
		}
	}

	return nil
}

// FindTagsLatestHistogram values contain recent metric histogram data.
type FindTagsLatestHistogram struct {
	Time  int64
	Value *string
}

// MarshalJSON encodes a FindTagsLatestHistogram value into a JSON format byte
// slice.
func (ftl *FindTagsLatestHistogram) MarshalJSON() ([]byte, error) {
	v := []interface{}{ftl.Time, ftl.Value}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON format byte slice into a
// FindTagsLatestHistogram value.
func (ftl *FindTagsLatestHistogram) UnmarshalJSON(b []byte) error {
	v := []interface{}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if len(v) != 2 {
		return fmt.Errorf("unable to decode latest histogram value, "+
			"invalid length: %v", string(b))
	}

	if fv, ok := v[0].(float64); ok {
		ftl.Time = int64(fv)
	} else {
		return fmt.Errorf("unable to decode latest histogram value, "+
			"invalid timestamp: %v", string(b))
	}

	if v[1] != nil {
		if sv, ok := v[1].(string); ok {
			ftl.Value = &sv
		} else {
			return fmt.Errorf("unable to decode latest histogram value, "+
				"invalid value: %v", string(b))
		}
	}

	return nil
}

// FindTags retrieves metrics that are associated with the provided tag query.
func (sc *SnowthClient) FindTags(accountID int64, query string,
	options *FindTagsOptions, nodes ...*SnowthNode,
) (*FindTagsResult, error) {
	return sc.FindTagsContext(context.Background(), accountID, query,
		options, nodes...)
}

// FindTagsContext is the context aware version of FindTags.
func (sc *SnowthClient) FindTagsContext(ctx context.Context, accountID int64,
	query string, options *FindTagsOptions,
	nodes ...*SnowthNode,
) (*FindTagsResult, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := fmt.Sprintf("%s?query=%s",
		fmt.Sprintf("/find/%d/tags", accountID),
		url.QueryEscape(query))

	starts, ends := "", ""
	if !options.Start.IsZero() && options.Start.Unix() != 0 {
		starts = formatTimestamp(options.Start)
	}

	if !options.End.IsZero() && options.End.Unix() != 0 {
		ends = formatTimestamp(options.End)
	}

	if options.StartStr != "" {
		starts = url.QueryEscape(options.StartStr)
	}

	if options.EndStr != "" {
		ends = url.QueryEscape(options.EndStr)
	}

	if starts != "" {
		u += fmt.Sprintf("&activity_start_secs=%s", starts)
	}

	if ends != "" {
		u += fmt.Sprintf("&activity_end_secs=%s", ends)
	}

	u += fmt.Sprintf("&activity=%d", options.Activity)
	u += fmt.Sprintf("&latest=%d", options.Latest)

	if options.CountOnly != 0 {
		u += fmt.Sprintf("&count_only=%d", options.CountOnly)
	}

	hdrs := http.Header{}
	if options.Limit != 0 {
		hdrs.Set("X-Snowth-Advisory-Limit",
			strconv.FormatInt(options.Limit, 10))
	}

	r := &FindTagsResult{}

	body, header, err := sc.DoRequestContext(ctx, node, "GET", u, nil, hdrs)
	if err != nil {
		return nil, err
	}

	if options.CountOnly != 0 {
		if err := decodeJSON(body, &r); err != nil {
			return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
		}

		return r, nil
	}

	if err := decodeJSON(body, &r.Items); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	r.Count = int64(len(r.Items))

	if header == nil {
		return r, nil
	}

	if v := header.Get("X-Snowth-Search-Result-Count"); v != "" {
		if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
			r.Count = iv
		}
	}

	if v := header.Get("X-Snowth-Search-Result-Count-Is-Estimate"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			r.Estimate = bv
		}
	}

	return r, nil
}

// FindTagCats retrieves tag categories that are associated with the
// provided query.
func (sc *SnowthClient) FindTagCats(accountID int64, query string,
	nodes ...*SnowthNode,
) ([]string, error) {
	return sc.FindTagCatsContext(context.Background(), accountID, query,
		nodes...)
}

// FindTagCatsContext is the context aware version of FindTagCats.
func (sc *SnowthClient) FindTagCatsContext(ctx context.Context,
	accountID int64, query string, nodes ...*SnowthNode,
) ([]string, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := fmt.Sprintf("%s?query=%s",
		fmt.Sprintf("/find/%d/tag_cats", accountID),
		url.QueryEscape(query))

	r := []string{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// FindTagVals retrieves tag values that are associated with the
// provided query.
func (sc *SnowthClient) FindTagVals(accountID int64, query, category string,
	nodes ...*SnowthNode,
) ([]string, error) {
	return sc.FindTagValsContext(context.Background(), accountID, query,
		category, nodes...)
}

// FindTagValsContext is the context aware version of FindTagVals.
func (sc *SnowthClient) FindTagValsContext(ctx context.Context,
	accountID int64, query, category string,
	nodes ...*SnowthNode,
) ([]string, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := fmt.Sprintf("%s?query=%s&category=%s",
		fmt.Sprintf("/find/%d/tag_vals", accountID),
		url.QueryEscape(query), url.QueryEscape(category))

	r := []string{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// CheckTags values contain check tag data from IRONdb.
type CheckTags map[string][]string

// ModifyTags values contain lists of check tags to add and remove.
type ModifyTags struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// GetCheckTags retrieves check tags from IRONdb for a specified check.
func (sc *SnowthClient) GetCheckTags(checkUUID string,
	nodes ...*SnowthNode,
) (CheckTags, error) {
	return sc.GetCheckTagsContext(context.Background(), checkUUID, nodes...)
}

// GetCheckTagsContext is the context aware version of GetCheckTags.
func (sc *SnowthClient) GetCheckTagsContext(ctx context.Context,
	checkUUID string, nodes ...*SnowthNode,
) (CheckTags, error) {
	if _, err := uuid.Parse(checkUUID); err != nil {
		return nil, fmt.Errorf("invalid check uuid: %w", err)
	}

	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := fmt.Sprintf("/meta/check/tag/%s", checkUUID)

	r := CheckTags{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, nil
}

// DeleteCheckTags removes check tags from IRONdb for a specified check.
func (sc *SnowthClient) DeleteCheckTags(checkUUID string,
	nodes ...*SnowthNode,
) error {
	return sc.DeleteCheckTagsContext(context.Background(), checkUUID, nodes...)
}

// DeleteCheckTagsContext is the context aware version of DeleteCheckTags.
func (sc *SnowthClient) DeleteCheckTagsContext(ctx context.Context,
	checkUUID string, nodes ...*SnowthNode,
) error {
	if _, err := uuid.Parse(checkUUID); err != nil {
		return fmt.Errorf("invalid check uuid: %w", err)
	}

	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return fmt.Errorf("unable to get active node")
	}

	u := fmt.Sprintf("/meta/check/tag/%s", checkUUID)

	_, _, err := sc.DoRequestContext(ctx, node, "DELETE", u, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCheckTags adds and removes tags for a specified check.
// DANGER: Ths function should not be used to set IRONdb check tags
// independently of the Circonus API. Doing so could result in tag data
// corruption, and malfunctioning metric searches.
func (sc *SnowthClient) UpdateCheckTags(checkUUID string,
	tags []string, nodes ...*SnowthNode,
) (int64, error) {
	return sc.UpdateCheckTagsContext(context.Background(), checkUUID, tags,
		nodes...)
}

// UpdateCheckTagsContext is the context aware version of UpdateCheckTags.
func (sc *SnowthClient) UpdateCheckTagsContext(ctx context.Context,
	checkUUID string, tags []string, nodes ...*SnowthNode,
) (int64, error) {
	if _, err := uuid.Parse(checkUUID); err != nil {
		return 0, fmt.Errorf("invalid check uuid: %w", err)
	}

	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return 0, fmt.Errorf("unable to get active node")
	}

	old, err := sc.GetCheckTagsContext(ctx, checkUUID, node)
	if err != nil {
		return 0, err
	}

	ex, ok := old[checkUUID]
	if !ok {
		return 0, fmt.Errorf("failed to retrieve existing check tags: %v",
			checkUUID)
	}

	if len(tags) == 0 {
		if err := sc.DeleteCheckTagsContext(ctx, checkUUID); err != nil {
			return 0, err
		}

		return 0, nil
	}

	tags, err = encodeTags(tags)
	if err != nil {
		return 0, err
	}

	del := []string{}

	for _, oldTag := range ex {
		d := true

		for _, newTag := range tags {
			if oldTag == newTag {
				d = false

				break
			}
		}

		if d {
			del = append(del, oldTag)
		}
	}

	mod := ModifyTags{
		Add:    tags,
		Remove: del,
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&mod); err != nil {
		return 0, fmt.Errorf("unable to encode payload: %w", err)
	}

	u := fmt.Sprintf("/meta/check/tag/%s", checkUUID)

	_, _, err = sc.DoRequestContext(ctx, node, "POST", u, buf, nil)
	if err != nil {
		return 0, fmt.Errorf("unable to post tags update: %w", err)
	}

	return int64(len(tags) + len(del)), nil
}

// encodeTags performs base64 encoding on tags when needed.
func encodeTags(tags []string) ([]string, error) {
	catTest := regexp.MustCompile("^[`+A-Za-z0-9!@#\\$%^&\"'\\/\\?\\._\\-]*$")
	valTest := regexp.MustCompile("^[`+A-Za-z0-9!@#\\$%^&\"'\\/\\?\\._\\-:=]*$")
	res := []string{}

	for _, tag := range tags {
		if tag == "" {
			continue
		}

		parts := strings.SplitN(tag, ":", 2)

		cat := parts[0]

		if strings.HasPrefix(cat, `b"`) && strings.HasSuffix(cat, `"`) {
			cat = strings.TrimPrefix(strings.TrimSuffix(cat, `"`), `b"`)

			cat, err := base64.StdEncoding.DecodeString(cat)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 tag category: %v %w",
					cat, err)
			}
		}

		if !catTest.MatchString(cat) {
			cat = `b"` + base64.StdEncoding.EncodeToString([]byte(cat)) + `"`
		}

		if cat == "" {
			return nil, fmt.Errorf("invalid tag passed: %v", tag)
		}

		if len(parts) < 2 {
			res = append(res, cat+":")

			continue
		}

		val := parts[1]

		if strings.HasPrefix(val, `b"`) && strings.HasSuffix(val, `"`) {
			val = strings.TrimPrefix(strings.TrimSuffix(val, `"`), `b"`)

			val, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 tag value: %v %w",
					val, err)
			}
		}

		if !valTest.MatchString(val) {
			val = `b"` + base64.StdEncoding.EncodeToString([]byte(val)) + `"`
		}

		res = append(res, cat+":"+val)
	}

	return res, nil
}
