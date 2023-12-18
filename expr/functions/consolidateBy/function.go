package consolidateBy

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type consolidateBy struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &consolidateBy{}
	functions := []string{"consolidateBy"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// consolidateBy(seriesList, aggregationMethod)
func (f *consolidateBy) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	name, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLink()
		r.Name = "consolidateBy(" + a.Name + ",\"" + name + "\")"
		r.ConsolidationFunc = name
		r.Tags["consolidateBy"] = name
		results[i] = r
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *consolidateBy) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"consolidateBy": {
			Description: "Takes one metric or a wildcard seriesList and a consolidation function name.\n\nValid function names are 'sum', 'average', 'min', 'max', 'first' & 'last'.\n\nWhen a graph is drawn where width of the graph size in pixels is smaller than\nthe number of datapoints to be graphed, Graphite consolidates the values to\nto prevent line overlap. The consolidateBy() function changes the consolidation\nfunction from the default of 'average' to one of 'sum', 'max', 'min', 'first', or 'last'.\nThis is especially useful in sales graphs, where fractional values make no sense and a 'sum'\nof consolidated values is appropriate.\n\n.. code-block:: none\n\n  &target=consolidateBy(Sales.widgets.largeBlue, 'sum')\n  &target=consolidateBy(Servers.web01.sda1.free_space, 'max')",
			Function:    "consolidateBy(seriesList, consolidationFunc)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "consolidateBy",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "consolidationFunc",
					Options:  types.StringsToSuggestionList(consolidations.AvailableConsolidationFuncs()),
					Required: true,
					Type:     types.String,
				},
			},
		},
	}
}
