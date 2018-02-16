package types

import (
	"encoding/json"
	"fmt"
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
)

// MarshalJSON marshals metric data to JSON
func (t FunctionType) MarshalJSON() ([]byte, error) {
	switch t {
	case AggFunc:
		return json.Marshal("aggFunc")
	case Boolean:
		return json.Marshal("boolean")
	case Date:
		return json.Marshal("date")
	case Float:
		return json.Marshal("float")
	case IntOrInterval:
		return json.Marshal("intOrInterval")
	case Integer:
		return json.Marshal("integer")
	case Interval:
		return json.Marshal("interval")
	case Node:
		return json.Marshal("node")
	case NodeOrTag:
		return json.Marshal("nodeOrTag")
	case SeriesList:
		return json.Marshal("seriesList")
	case SeriesLists:
		return json.Marshal("seriesLists")
	case String:
		return json.Marshal("string")
	case Tag:
		return json.Marshal("tag")
	}

	return nil, fmt.Errorf("unknown type specified: %v", t)
}

// FunctionParam contains list of all available parameters of function
type FunctionParam struct {
	Name        string       `json:"name"`
	Multiple    bool         `json:"multiple,omitempty"`
	Required    bool         `json:"required,omitempty"`
	Type        FunctionType `json:"type,omitempty"`
	Options     []string     `json:"options,omitempty"`
	Suggestions []string     `json:"suggestions,omitempty"`
	Default     string       `json:"default,omitempty"`
}

// FunctionDescription contains full function description.
type FunctionDescription struct {
	Description string          `json:"description"`
	Function    string          `json:"function"`
	Group       string          `json:"group"`
	Module      string          `json:"module"`
	Name        string          `json:"name"`
	Params      []FunctionParam `json:"params,omitempty"`
}
