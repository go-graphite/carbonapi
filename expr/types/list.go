package types

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// FunctionType is a special type to handle parameter type in function description
type FunctionType int

const (
	// AggFunc is a constant for AggregationFunction type
	AggFunc FunctionType = iota
	// Boolean is a constant for Boolean type
	Boolean
	// Date is a constant for Date type
	Date
	// Float is a constant for Float type
	Float
	// IntOrInterval is a constant for Interval-Or-Integer type
	IntOrInterval
	// Integer is a constant for Integer type
	Integer
	// Interval is a constant for Interval type
	Interval
	// Node is a constant for Node type
	Node
	// NodeOrTag is a constant for Node-Or-Tag type
	NodeOrTag
	// SeriesList is a constant for SeriesList type
	SeriesList
	// SeriesLists is a constant for SeriesLists type
	SeriesLists
	// String is a constant for String type
	String
	// Tag is a constant for Tag type
	Tag
	// Any is a constant for `Any` type
	Any
	// AggOrSeriesFunc is a constant for AggregationFunction or SeriesList
	AggOrSeriesFunc
	// IntOrInf is a constant for Integer that can be infinity
	IntOrInf
)

var strToFunctionType = map[string]FunctionType{
	"aggFunc":         AggFunc,
	"boolean":         Boolean,
	"date":            Date,
	"float":           Float,
	"intOrInterval":   IntOrInterval,
	"integer":         Integer,
	"interval":        Interval,
	"node":            Node,
	"nodeOrTag":       NodeOrTag,
	"seriesList":      SeriesList,
	"seriesLists":     SeriesLists,
	"string":          String,
	"tag":             Tag,
	"any":             Any,
	"aggOrSeriesFunc": AggOrSeriesFunc,
	"intOrInf":        IntOrInf,
}

// FunctionTypeToStr provides a mapping between internal type constants and graphite-friendly string that have a name of a type
var FunctionTypeToStr = map[FunctionType]string{
	AggFunc:         "aggFunc",
	Boolean:         "boolean",
	Date:            "date",
	Float:           "float",
	IntOrInterval:   "intOrInterval",
	Integer:         "integer",
	Interval:        "interval",
	Node:            "node",
	NodeOrTag:       "nodeOrTag",
	SeriesList:      "seriesList",
	SeriesLists:     "seriesLists",
	String:          "string",
	Tag:             "tag",
	Any:             "any",
	AggOrSeriesFunc: "aggOrSeriesFunc",
	IntOrInf:        "intOrInf",
}

// MarshalJSON marshals metric data to JSON
func (t FunctionType) MarshalJSON() ([]byte, error) {
	v, ok := FunctionTypeToStr[t]
	if ok {
		return json.Marshal(v)
	}

	return nil, fmt.Errorf("unknown type specified: %v", t)
}

func (t *FunctionType) UnmarshalJSON(d []byte) error {
	var err error
	s := strings.Trim(string(d), "\n\t \"")
	v, ok := strToFunctionType[s]
	if ok {
		*t = v
	} else {
		err = fmt.Errorf("failed to parse value '%v'", string(d))
	}

	return err
}

type SuggestionTypes int

const (
	SInt SuggestionTypes = iota
	SInt32
	SInt64
	SUint
	SUint32
	SUint64
	SFloat64
	SString
	SBool
	SNone
)

type Suggestion struct {
	Type  SuggestionTypes
	Value interface{}
}

func NewSuggestion(arg interface{}) *Suggestion {
	switch v := arg.(type) {
	case int:
		return &Suggestion{Type: SInt, Value: v}
	case int32:
		return &Suggestion{Type: SInt32, Value: v}
	case int64:
		return &Suggestion{Type: SInt64, Value: v}
	case uint:
		return &Suggestion{Type: SUint, Value: v}
	case uint32:
		return &Suggestion{Type: SUint32, Value: v}
	case uint64:
		return &Suggestion{Type: SUint64, Value: v}
	case float64:
		return &Suggestion{Type: SFloat64, Value: v}
	case string:
		return &Suggestion{Type: SString, Value: v}
	case bool:
		return &Suggestion{Type: SBool, Value: v}
	}

	return &Suggestion{Type: SNone}
}

func NewSuggestions(vaArgs ...interface{}) []*Suggestion {
	res := make([]*Suggestion, 0, len(vaArgs))

	for _, a := range vaArgs {
		res = append(res, NewSuggestion(a))
	}

	return res
}

// MarshalJSON marshals metric data to JSON
func (t Suggestion) MarshalJSON() ([]byte, error) {
	switch t.Type {
	case SInt:
		return json.Marshal(t.Value.(int))
	case SInt32:
		return json.Marshal(t.Value.(int32))
	case SInt64:
		return json.Marshal(t.Value.(int64))
	case SUint:
		return json.Marshal(t.Value.(uint))
	case SUint32:
		return json.Marshal(t.Value.(int64))
	case SUint64:
		return json.Marshal(t.Value.(uint64))
	case SFloat64:
		return json.Marshal(t.Value.(float64))
	case SString:
		return json.Marshal(t.Value.(string))
	case SBool:
		return json.Marshal(t.Value.(bool))
	case SNone:
		return []byte{}, nil
	}

	return nil, fmt.Errorf("unknown type %v", t.Type)
}

func (t *Suggestion) UnmarshalJSON(d []byte) error {
	if len(d) == 0 {
		t.Type = SNone
		return nil
	}

	var res interface{}
	err := json.Unmarshal(d, &res)
	if err != nil {
		switch string(d) {
		case "1e9999":
			res = math.Inf(1)
		default:
			return err
		}
	}
	switch v := res.(type) {
	case int:
		t.Type = SInt
		t.Value = v
	case int32:
		t.Type = SInt32
		t.Value = v
	case int64:
		t.Type = SInt64
		t.Value = v
	case float64:
		t.Type = SFloat64
		t.Value = v
	case string:
		t.Type = SString
		t.Value = v
	case bool:
		t.Type = SBool
		t.Value = v
	case nil:
		t.Type = SNone
		t.Value = nil
	default:
		return fmt.Errorf("unknown type for suggestion: %v, %T", string(d), v)
	}

	return nil
}

type Option struct {
	Type  SuggestionTypes
	Value interface{}
}

func StringsToSuggestionList(in []string) []Suggestion {
	res := make([]Suggestion, 0, len(in))

	for _, v := range in {
		res = append(res, Suggestion{Type: SString, Value: v})
	}
	return res
}

// FunctionParam contains list of all available parameters of function
type FunctionParam struct {
	Name        string        `json:"name"`
	Multiple    bool          `json:"multiple,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Type        FunctionType  `json:"type,omitempty"`
	Options     []Suggestion  `json:"options,omitempty"`
	Suggestions []*Suggestion `json:"suggestions,omitempty"`
	Default     *Suggestion   `json:"default,omitempty"`
}

// FunctionDescription contains full function description.
type FunctionDescription struct {
	Description string          `json:"description"`
	Function    string          `json:"function"`
	Group       string          `json:"group"`
	Module      string          `json:"module"`
	Name        string          `json:"name"`
	Params      []FunctionParam `json:"params,omitempty"`

	Proxied bool `json:"proxied"`

	Aggretated    bool `json:"aggregate,omitempty"`       //  function aggregate metrics, for tests and verify results in future
	NameChange    bool `json:"name-change,omitempty"`     // function change name, for tests and verify results in future
	NameTagChange bool `json:"name-tag-change,omitempty"` //  function change name tag, for tests and verify results in future
	ValuesChange  bool `json:"values-change,omitempty"`   //  function change values, for tests and verify results in future
}
