package example

// THIS PACKAGE SHOULD NOT BE IMPORTED
// USE IT AS AN EXAMPLE OF HOW TO WRITE NEW FUNCTION

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &example{}
	functions := []string{"example", "examples"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type example struct {
	interfaces.FunctionBase
}

func (f *example) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	_ = helper.Backref
	return nil, nil
}

func (f *example) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"example": {
			Description: "This is just an example of function",
			Function:    "example(seriesList, func, xFilesFactor=None)",
			Group:       "Example",
			Module:      "graphite.render.functions",
			Name:        "example",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
