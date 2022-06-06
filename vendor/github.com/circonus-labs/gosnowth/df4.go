package gosnowth

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
)

// DF4Response values represent time series data in the DF4 format.
type DF4Response struct {
	Ver  string          `json:"version,omitempty"`
	Head DF4Head         `json:"head"`
	Meta []DF4Meta       `json:"meta,omitempty"`
	Data [][]interface{} `json:"data,omitempty"`
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
	Count  int64 `json:"count"`
	Start  int64 `json:"start"`
	Period int64 `json:"period"`
	// TODO: Replace the Explain value with an actual typed schema when one
	// becomes available.
	Explain json.RawMessage `json:"explain,omitempty"`
}

// Copy returns a deep copy of the base DF4 response.
func (dr *DF4Response) Copy() *DF4Response {
	b := &DF4Response{
		Data: make([][]interface{}, len(dr.Data)),
		Meta: make([]DF4Meta, len(dr.Meta)),
		Ver:  dr.Ver,
		Head: DF4Head{
			Count:   dr.Head.Count,
			Start:   dr.Head.Start,
			Period:  dr.Head.Period,
			Explain: dr.Head.Explain,
		},
	}

	copy(b.Meta, dr.Meta)
	for i, v := range dr.Data {
		b.Data[i] = make([]interface{}, len(v))
		copy(b.Data[i], v)
	}

	return b
}

// replaceInf is used to remove infinity and NaN values from DF4 JSON strings
// prior to attempting to parse them into DF4Response values.
func replaceInf(b []byte) []byte {
	v := make([]byte, len(b))
	copy(v, b)

	v = bytes.ReplaceAll(v, []byte("+inf,"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+","))
	v = bytes.ReplaceAll(v, []byte("+inf]"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"]"))
	v = bytes.ReplaceAll(v, []byte("+inf\n"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"\n"))
	v = bytes.ReplaceAll(v, []byte("-inf,"), []byte(
		strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64)+","))
	v = bytes.ReplaceAll(v, []byte("-inf]"), []byte(
		strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64)+"]"))
	v = bytes.ReplaceAll(v, []byte("-inf\n"), []byte(
		strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64)+"\n"))
	v = bytes.ReplaceAll(v, []byte("inf,"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+","))
	v = bytes.ReplaceAll(v, []byte("inf]"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"]"))
	v = bytes.ReplaceAll(v, []byte("inf\n"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"\n"))

	v = bytes.ReplaceAll(v, []byte("NaN,"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+","))
	v = bytes.ReplaceAll(v, []byte("NaN]"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"]"))
	v = bytes.ReplaceAll(v, []byte("NaN\n"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"\n"))
	v = bytes.ReplaceAll(v, []byte("nan,"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+","))
	v = bytes.ReplaceAll(v, []byte("nan]"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"]"))
	v = bytes.ReplaceAll(v, []byte("nan\n"), []byte(
		strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64)+"\n"))

	return v
}
