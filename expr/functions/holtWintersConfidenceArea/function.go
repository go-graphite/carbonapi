//go:build !cairo
// +build !cairo

package holtWintersConfidenceArea

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

var UnsupportedError = fmt.Errorf("must build w/ cairo support")

type holtWintersConfidenceArea struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &holtWintersConfidenceArea{}
	functions := []string{"holtWintersConfidenceArea"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *holtWintersConfidenceArea) Do(_ context.Context, _ parser.Expr, _, _ int64, _ map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return nil, UnsupportedError
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *holtWintersConfidenceArea) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"holtWintersConfidenceArea": {
			Description: "Performs a Holt-Winters forecast using the series as input data and plots\n the area between the upper and lower bands of the predicted forecast deviations.",
			Function:    "holtWintersConfidenceArea(seriesList, delta=3, bootstrapInterval='7d')",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "holtWintersConfidenceArea",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(3),
					Name:    "delta",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion("7d"),
					Name:    "bootstrapInterval",
					Suggestions: types.NewSuggestions(
						"7d",
						"30d",
					),
					Type: types.Interval,
				},
			},
		},
	}
}
