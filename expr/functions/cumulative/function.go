package cumulative

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type cumulative struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &cumulative{}
	functions := []string{"cumulative"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// cumulative(seriesList)
func (f *cumulative) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.AggregateFunction = types.AggSum
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *cumulative) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"cumulative": {
			Description: "Takes one metric or a wildcard seriesList.\n\nWhen a graph is drawn where width of the graph size in pixels is smaller than\nthe number of datapoints to be graphed, Graphite consolidates the values to\nto prevent line overlap. The cumulative() function changes the consolidation\nfunction from the default of 'average' to 'sum'. This is especially useful in\nsales graphs, where fractional values make no sense and a 'sum' of consolidated\nvalues is appropriate.\n\nAlias for :func:`consolidateBy(series, 'sum') <graphite.render.functions.consolidateBy>`\n\n.. code-block:: none\n\n  &target=cumulative(Sales.widgets.largeBlue)",
			Function:    "cumulative(seriesList)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "cumulative",
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
