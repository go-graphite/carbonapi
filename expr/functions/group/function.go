package group

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type group struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &group{}
	functions := []string{"group"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// group(*seriesLists)
func (f *group) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(ctx, f.GetEvaluator(), e, from, until, values)
	if err != nil {
		return nil, err
	}

	return args, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *group) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"group": {
			Description: "Takes an arbitrary number of seriesLists and adds them to a single seriesList. This is used\nto pass multiple seriesLists to a function which only takes one",
			Function:    "group(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "group",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Type:     types.SeriesList,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
		},
	}
}
