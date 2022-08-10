package example

// THIS PACKAGE SHOULD NOT BE IMPORTED
// USE IT AS AN EXAMPLE OF HOW TO WRITE NEW FUNCTION

import (
	"context"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type example struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &example{}
	functions := []string{"example", "examples"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *example) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	_ = helper.Backref
	return nil, nil
}

func (f *example) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
			SeriesChange: false, // function aggregate metrics or change series items count
			NameChange:   false, // name changed
			TagsChange:   false, // name tag changed
			ValuesChange: false, // values changed
		},
	}
}
