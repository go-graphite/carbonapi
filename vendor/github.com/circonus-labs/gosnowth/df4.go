package gosnowth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// DF4Response values represent time series data in the DF4 format.
type DF4Response struct {
	Ver   string    `json:"version,omitempty"`
	Head  DF4Head   `json:"head"`
	Meta  []DF4Meta `json:"meta"`
	Data  []DF4Data `json:"data"`
	Query string    `json:"-"`
}

// DF4Meta values contain information and metadata about the metrics in a DF4
// time series data response.
type DF4Meta struct {
	Kind  string   `json:"kind"`
	Label string   `json:"label"`
	Tags  []string `json:"tags,omitempty"`
}

// DF4Head values contain information about the time range of the data elements
// in a DF4 time series data response.
type DF4Head struct {
	Count   int64           `json:"count"`
	Start   int64           `json:"start"`
	Period  int64           `json:"period"`
	Error   []string        `json:"error,omitempty"`
	Warning []string        `json:"warning,omitempty"`
	Explain json.RawMessage `json:"explain,omitempty"`
}

// MarshalJSON encodes a DF4Head value into a JSON format byte slice.
func (h *DF4Head) MarshalJSON() ([]byte, error) {
	v := struct {
		Count   int64           `json:"count"`
		Start   int64           `json:"start"`
		Period  int64           `json:"period"`
		Error   json.RawMessage `json:"error,omitempty"`
		Warning json.RawMessage `json:"warning,omitempty"`
		Explain json.RawMessage `json:"explain,omitempty"`
	}{
		Count:   h.Count,
		Start:   h.Start,
		Period:  h.Period,
		Explain: h.Explain,
	}

	if len(h.Error) == 1 {
		b, err := json.Marshal(h.Error[0])
		if err != nil {
			return nil, fmt.Errorf(
				"unable to marshal df4 head error value into JSON data: %w",
				err)
		}

		v.Error = b
	} else if len(h.Error) > 1 {
		b, err := json.Marshal(h.Error)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to marshal df4 head error value into JSON data: %w",
				err)
		}

		v.Error = b
	}

	if len(h.Warning) == 1 {
		b, err := json.Marshal(h.Warning[0])
		if err != nil {
			return nil, fmt.Errorf(
				"unable to marshal df4 head warning value into JSON data: %w",
				err)
		}

		v.Warning = b
	} else if len(h.Warning) > 1 {
		b, err := json.Marshal(h.Warning)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to marshal df4 head warning value into JSON data: %w",
				err)
		}

		v.Warning = b
	}

	return json.Marshal(v)
}

// UnmarshalJSON decodes a DF4Head value from a JSON format byte slice.
func (h *DF4Head) UnmarshalJSON(b []byte) error { //nolint:gocyclo
	m := map[string]interface{}{}

	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf(
			"unable to unmarshal df4 head value from JSON data: %w", err)
	}

	for k, v := range m {
		switch k {
		case "count":
			switch vt := v.(type) {
			case float64:
				h.Count = int64(vt)
			case string:
				i, err := strconv.ParseInt(vt, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"unable to parse %s value from JSON data: %w",
						k, err)
				}

				h.Count = i
			default:
				return fmt.Errorf("unable to parse %s value from JSON data", k)
			}
		case "start":
			switch vt := v.(type) {
			case float64:
				h.Start = int64(vt)
			case string:
				i, err := strconv.ParseInt(vt, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"unable to parse %s value from JSON data: %w",
						k, err)
				}

				h.Start = i
			default:
				return fmt.Errorf("unable to parse %s value from JSON data", k)
			}
		case "period":
			switch vt := v.(type) {
			case float64:
				h.Period = int64(vt)
			case string:
				i, err := strconv.ParseInt(vt, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"unable to parse %s value from JSON data: %w",
						k, err)
				}

				h.Period = i
			default:
				return fmt.Errorf("unable to parse %s value from JSON data", k)
			}
		case "error":
			switch vt := v.(type) {
			case string:
				h.Error = []string{vt}
			case []string:
				h.Error = vt
			case []interface{}:
				if len(vt) > 0 {
					ss := make([]string, len(vt))

					for i, vs := range vt {
						s, ok := vs.(string)
						if !ok {
							return fmt.Errorf(
								"unable to parse %s value from JSON data", k)
						}

						ss[i] = s
					}

					h.Error = ss
				}
			default:
				return fmt.Errorf("unable to parse %s value from JSON data", k)
			}
		case "warning":
			switch vt := v.(type) {
			case string:
				h.Warning = []string{vt}
			case []string:
				h.Warning = vt
			case []interface{}:
				if len(vt) > 0 {
					ss := make([]string, len(vt))

					for i, vs := range vt {
						s, ok := vs.(string)
						if !ok {
							return fmt.Errorf(
								"unable to parse %s value from JSON data", k)
						}

						ss[i] = s
					}

					h.Warning = ss
				}
			default:
				return fmt.Errorf("unable to parse %s value from JSON data", k)
			}
		case "explain":
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf(
					"unable to parse %s value from JSON data: %w", k, err)
			}

			h.Explain = b
		}
	}

	return nil
}

// DF4Data values contain slices of data points of DF4 format time series data.
type DF4Data []interface{}

// NullEmpty sets values within a DF4Data value equal to an empty array to nil.
func (d *DF4Data) NullEmpty() {
	if d == nil {
		return
	}

	for i, v := range *d {
		if vv, ok := v.([]interface{}); ok && len(vv) == 0 {
			(*d)[i] = nil
		}
	}
}

// Numeric retrieves the data in this value as a slice of float64 values.
func (dd *DF4Data) Numeric() []*float64 {
	if dd == nil {
		return nil
	}

	r := make([]*float64, len(*dd))

	for i, v := range *dd {
		switch tv := v.(type) {
		case float64:
			r[i] = &tv
		case int64:
			tvv := float64(tv)

			r[i] = &tvv
		case int:
			tvv := float64(tv)

			r[i] = &tvv
		case float32:
			tvv := float64(tv)

			r[i] = &tvv
		}
	}

	return r
}

// Text retrieves the data in this value as a slice of string values.
func (dd *DF4Data) Text() []*string {
	if dd == nil {
		return nil
	}

	r := make([]*string, len(*dd))

	for i, v := range *dd {
		switch vv := v.(type) {
		case string:
			r[i] = &vv
		case []interface{}:
			if len(vv) > 0 {
				if vvs, ok := vv[0].([]interface{}); ok && len(vvs) > 1 {
					if s, ok := vvs[1].(string); ok {
						r[i] = &s

						break
					}
				}
			}
		}
	}

	return r
}

// Histogram retrieves the data in this value as a slice of map[string]int64
// values.
func (dd *DF4Data) Histogram() []*map[string]int64 {
	if dd == nil {
		return nil
	}

	r := make([]*map[string]int64, len(*dd))

	for i, v := range *dd {
		if m, ok := v.(map[string]interface{}); ok {
			mv := make(map[string]int64, len(m))

			for k, iv := range m {
				switch tv := iv.(type) {
				case int64:
					mv[k] = tv
				case int:
					mv[k] = int64(tv)
				case float64:
					mv[k] = int64(tv)
				case float32:
					mv[k] = int64(tv)
				}
			}

			r[i] = &mv
		} else if m, ok := v.(map[string]int64); ok {
			r[i] = &m
		}
	}

	return r
}

// Copy returns a deep copy of the base DF4 response.
func (dr *DF4Response) Copy() *DF4Response {
	b := &DF4Response{
		Data: make([]DF4Data, len(dr.Data)),
		Meta: make([]DF4Meta, len(dr.Meta)),
		Ver:  dr.Ver,
		Head: DF4Head{
			Count:   dr.Head.Count,
			Start:   dr.Head.Start,
			Period:  dr.Head.Period,
			Error:   dr.Head.Error,
			Warning: dr.Head.Warning,
			Explain: dr.Head.Explain,
		},
	}

	copy(b.Meta, dr.Meta)

	for i, v := range dr.Data {
		b.Data[i] = make(DF4Data, len(v))
		copy(b.Data[i], v)
	}

	return b
}

// replaceInf is used to remove infinity and NaN values from DF4 JSON strings
// prior to attempting to parse them into DF4Response values.
func replaceInf(b []byte) []byte {
	v := make([]byte, len(b))
	copy(v, b)

	maxFloat := strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)
	negMaxFloat := strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64)

	v = bytes.ReplaceAll(v, []byte("+inf,"), []byte(maxFloat+","))
	v = bytes.ReplaceAll(v, []byte("+inf]"), []byte(maxFloat+"]"))
	v = bytes.ReplaceAll(v, []byte("+inf\n"), []byte(maxFloat+"\n"))
	v = bytes.ReplaceAll(v, []byte("-inf,"), []byte(negMaxFloat+","))
	v = bytes.ReplaceAll(v, []byte("-inf]"), []byte(negMaxFloat+"]"))
	v = bytes.ReplaceAll(v, []byte("-inf\n"), []byte(negMaxFloat+"\n"))
	v = bytes.ReplaceAll(v, []byte("inf,"), []byte(maxFloat+","))
	v = bytes.ReplaceAll(v, []byte("inf]"), []byte(maxFloat+"]"))
	v = bytes.ReplaceAll(v, []byte("inf\n"), []byte(maxFloat+"\n"))

	v = bytes.ReplaceAll(v, []byte("NaN,"), []byte(maxFloat+","))
	v = bytes.ReplaceAll(v, []byte("NaN]"), []byte(maxFloat+"]"))
	v = bytes.ReplaceAll(v, []byte("NaN\n"), []byte(maxFloat+"\n"))
	v = bytes.ReplaceAll(v, []byte("nan,"), []byte(maxFloat+","))
	v = bytes.ReplaceAll(v, []byte("nan]"), []byte(maxFloat+"]"))
	v = bytes.ReplaceAll(v, []byte("nan\n"), []byte(maxFloat+"\n"))

	return v
}
