package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// CAQLQuery values represent CAQL queries and associated parameters.
type CAQLQuery struct {
	Explain              bool     `json:"explain"`
	IgnoreDurationLimits bool     `json:"ignore_duration_limits,omitempty"`
	Debug                byte     `json:"_debug,omitempty"`
	Period               int64    `json:"period,omitempty"`
	ID                   int64    `json:"_id,omitempty"`
	AccountID            int64    `json:"account_id,string,omitempty"`
	Start                int64    `json:"start_time,omitempty"`
	Timeout              int64    `json:"_timeout,omitempty"`
	MinPrefill           int64    `json:"min_prefill,omitempty"`
	End                  int64    `json:"end_time,omitempty"`
	Format               string   `json:"format,omitempty"`
	Query                string   `json:"q,omitempty"`
	PrepareResults       string   `json:"prepare_results,omitempty"`
	Method               string   `json:"method,omitempty"`
	Expansion            []string `json:"expansion,omitempty"`
}

// CAQLErrorArgs values represent CAQL request arguments returned in an error.
type CAQLErrorArgs struct {
	IgnoreDurationLimits bool     `json:"ignore_duration_limits"`
	Debug                byte     `json:"_debug"`
	Period               int64    `json:"period"`
	ID                   int64    `json:"_id"`
	AccountID            int64    `json:"account_id,string"`
	Start                int64    `json:"start_time"`
	Timeout              int64    `json:"_timeout"`
	MinPrefill           int64    `json:"min_prefill"`
	End                  int64    `json:"end_time"`
	Format               string   `json:"format"`
	Query                string   `json:"q"`
	PrepareResults       string   `json:"prepare_results"`
	Method               string   `json:"method"`
	Expansion            []string `json:"expansion"`
}

// CAQLUserError values contain messages describing a CAQL error for a user.
type CAQLUserError struct {
	Message string `json:"message,omitempty"`
}

// CAQLError values contain information about an error returned by the CAQL
// extension.
type CAQLError struct {
	Locals    []string      `json:"locals"`
	Method    string        `json:"method"`
	Trace     []string      `json:"trace"`
	UserError CAQLUserError `json:"user_error"`
	Status    string        `json:"status"`
	Arguments CAQLErrorArgs `json:"arguments"`
	Success   bool          `json:"success"`
}

// String returns this value as a JSON format string.
func (ce *CAQLError) String() string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(ce); err != nil {
		return "unable to encode JSON: " + err.Error()
	}

	return buf.String()
}

// Error returns this error as a JSON format string.
func (ce *CAQLError) Error() string {
	return ce.String()
}

// GetCAQLQuery retrieves data values for metrics matching a CAQL format.
func (sc *SnowthClient) GetCAQLQuery(q *CAQLQuery,
	nodes ...*SnowthNode,
) (*DF4Response, error) {
	return sc.GetCAQLQueryContext(context.Background(), q, nodes...)
}

// GetCAQLQueryContext is the context aware version of GetCAQLQuery.
func (sc *SnowthClient) GetCAQLQueryContext(ctx context.Context, q *CAQLQuery,
	nodes ...*SnowthNode,
) (*DF4Response, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if q == nil {
		q = &CAQLQuery{}
	}

	u := "/extension/lua/public/caql_v1"
	q.Format = "DF4"

	qBuf, err := encodeJSON(q)
	if err != nil {
		return nil, err
	}

	bBuf, err := io.ReadAll(qBuf)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body buffer: %w", err)
	}

	// CAQL extension does not like the JSON in the request body to end with \n.
	if strings.HasSuffix(string(bBuf), "\n") {
		bBuf = bBuf[:len(bBuf)-1]
	}

	r := &DF4Response{}

	body, _, err := sc.DoRequestContext(ctx, node, "POST", u,
		bytes.NewBuffer(bBuf), nil)
	if err != nil {
		if body != nil {
			cErr := &CAQLError{}
			if err := decodeJSON(body, &cErr); err == nil {
				return nil, cErr
			}
		}

		return nil, err
	}

	rb, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("unable to read IRONdb response body: %w", err)
	}

	rb = replaceInf(rb)

	if err := decodeJSON(bytes.NewBuffer(rb), &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	r.Query = q.Query

	return r, err
}
