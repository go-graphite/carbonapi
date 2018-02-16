package alias

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("alias", &alias{})
}

type alias struct {
	interfaces.FunctionBase
}

func (f *alias) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	alias, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = alias
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *alias) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"alias": {
			Description: "Takes one metric or a wildcard seriesList and a string in quotes.\nPrints the string instead of the metric name in the legend.\n\n.. code-block:: none\n\n  &target=alias(Sales.widgets.largeBlue,\"Large Blue Widgets\")",
			Function:    "alias(seriesList, newName)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "alias",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "newName",
					Required: true,
					Type:     types.String,
				},
			},
		},
	}
}