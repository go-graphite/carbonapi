package legendValue

import (
	"context"
	"fmt"

	"github.com/grafana/carbonapi/expr/consolidations"
	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type legendValue struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &legendValue{}
	functions := []string{"legendValue"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// legendValue(seriesList, newName)
func (f *legendValue) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var system string
	var methods []string
	for i := 1; i < len(e.Args()); i++ {
		method, err := e.GetStringArg(i)
		if err != nil {
			return nil, err
		}
		if method == "si" || method == "binary" {
			system = method
		} else {
			methods = append(methods, method)
		}
	}
	var results []*types.MetricData
	for _, a := range arg {
		r := a.CopyLink()
		for _, method := range methods {
			summary := consolidations.SummarizeValues(method, a.Values, a.XFilesFactor)
			if system == "" {
				r.Name = fmt.Sprintf("%s (%s: %f)", r.Name, method, summary)
			} else {
				v, prefix := helper.FormatUnits(summary, system)
				var formatted string
				if prefix != "" {
					prefix += " "
				}

				if v < 0.1 {
					formatted = fmt.Sprintf("%.9g%s", v, prefix)
				}
				if v < 1.0 {
					formatted = fmt.Sprintf("%.2f%s", v, prefix)
				}
				if formatted != "" {
					r.Name = fmt.Sprintf("%s (%s: %s)", r.Name, method, formatted)
				} else {
					r.Name = fmt.Sprintf("%s (%s: %f)", r.Name, method, summary)
				}
			}
		}

		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *legendValue) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"legendValue": {
			Description: "Takes one metric or a wildcard seriesList and a string in quotes.\nAppends a value to the metric name in the legend.  Currently one or several of: `last`, `avg`,\n`total`, `min`, `max`.\nThe last argument can be `si` (default) or `binary`, in that case values will be formatted in the\ncorresponding system.\n\n.. code-block:: none\n\n  &target=legendValue(Sales.widgets.largeBlue, 'avg', 'max', 'si')",
			Function:    "legendValue(seriesList, *valueTypes)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "legendValue",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "valuesTypes",
					Options:  types.StringsToSuggestionList(consolidations.AvailableSummarizers),
					Type:     types.String,
				},
			},
		},
	}
}
