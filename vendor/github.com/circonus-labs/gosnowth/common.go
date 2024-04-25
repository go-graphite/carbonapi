package gosnowth

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// IRONdbPutResponse values represent raw IRONdb PUT/POST responses.
type IRONdbPutResponse struct {
	Errors      uint64 `json:"errors"`
	Misdirected uint64 `json:"misdirected"`
	Records     uint64 `json:"records"`
	Updated     uint64 `json:"updated"`
}

// resolveURL resolves the address of a URL plus a string reference.
func resolveURL(baseURL *url.URL, ref string) string {
	refURL, _ := url.Parse(ref)

	return baseURL.ResolveReference(refURL).String()
}

// encodeJSON create a reader of JSON data representing an interface.
func encodeJSON(v interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	return buf, nil
}

// decodeJSON decodes JSON from a reader into an interface.
func decodeJSON(r io.Reader, v interface{}) error {
	if r == nil {
		return fmt.Errorf("unable to decode from nil reader")
	}

	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// encodeXML create a reader of XML data representing an interface.
func encodeXML(v interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := xml.NewEncoder(buf).Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode XML: %w", err)
	}

	return buf, nil
}

// decodeXML decodes XML from a reader into an interface.
func decodeXML(r io.Reader, v interface{}) error {
	if err := xml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	return nil
}

const million int = 1000000

// formatTimestamp returns a string containing a timestamp in the format used
// by the IRONdb API.
func formatTimestamp(t time.Time) string {
	if t.Nanosecond()/million != 0 {
		return fmt.Sprintf("%d.%03d", t.Unix(), t.Nanosecond()/million)
	}

	return fmt.Sprintf("%d", t.Unix())
}

// parseTimestamp attempts to parse an IRONdb API timestamp string into a valid
// time value.
func parseTimestamp(s string) (time.Time, error) {
	sp := strings.Split(s, ".")
	sec, nsec := int64(0), int64(0)

	var err error

	if len(sp) > 0 {
		if sec, err = strconv.ParseInt(sp[0], 10, 64); err != nil {
			return time.Time{}, fmt.Errorf("unable to parse timestamp %s: %s",
				s, err.Error())
		}
	}

	if len(sp) > 1 {
		if nsec, err = strconv.ParseInt(sp[1], 10, 64); err != nil {
			return time.Time{}, fmt.Errorf("unable to parse timestamp %s: %s",
				s, err.Error())
		}

		nsec *= int64(million)
	}

	return time.Unix(sec, nsec), nil
}

// parseDuration attempts to parse an IRONdb API duration string into a valid
// duration value.
func parseDuration(s string) (time.Duration, error) {
	if !strings.HasSuffix(s, "s") {
		s += "s"
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("unable to parse duration %s: %s",
			s, err.Error())
	}

	return d, nil
}
