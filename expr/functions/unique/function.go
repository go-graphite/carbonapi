package unique

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type unique struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &unique{}
	functions := []string{"unique"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// unique(seriesList)
func (f *unique) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	seenNames := make(map[string]bool)

	for _, a := range arg {
		if _, ok := seenNames[a.Name]; !ok {
			seenNames[a.Name] = true
			results = append(results, a)
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *unique) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"unique": {
			Description: "Takes an arbitrary number of seriesLists and returns unique series, filtered by name.\n\nExample:\n\n.. code-block:: none\n\n  &target=unique(mostDeviant(server.*.disk_free,5),lowestCurrent(server.*.disk_free,5))\n\n  Draws servers with low disk space, and servers with highly deviant disk space, but never the same series twice.",
			Function:    "unique(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "unique",
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
