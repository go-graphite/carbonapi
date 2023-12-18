package legendValue

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
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
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var system string
	var methods []string
	for i := 1; i < e.ArgsLen(); i++ {
		method, err := e.GetStringArg(i)
		if err != nil {
			return nil, err
		}
		if method == "si" || method == "binary" {
			system = method
		} else {
			if err := consolidations.CheckValidConsolidationFunc(method); err != nil {
				return nil, err
			}
			methods = append(methods, method)
		}
	}

	results := make([]*types.MetricData, len(arg))
	for i, a := range arg {
		r := a.CopyLink()
		var nameBuf strings.Builder
		nameBuf.Grow(len(r.Name) + len(methods)*5)
		nameBuf.WriteString(r.Name)
		for _, method := range methods {
			summary := consolidations.SummarizeValues(method, a.Values, a.XFilesFactor)
			nameBuf.WriteString(" (")
			nameBuf.WriteString(method)
			nameBuf.WriteString(": ")
			if system == "" {
				nameBuf.WriteString(strconv.FormatFloat(summary, 'g', -1, 64))
			} else {
				v, prefix := helper.FormatUnits(summary, system)
				if prefix != "" {
					prefix += " "
				}

				if math.Abs(v) < 0.1 {
					nameBuf.WriteString(strconv.FormatFloat(v, 'g', 9, 64))
				} else {
					nameBuf.WriteString(strconv.FormatFloat(v, 'f', 2, 64))
				}

				nameBuf.WriteString(prefix)
			}
			nameBuf.WriteString(")")
		}
		r.Name = nameBuf.String()

		results[i] = r
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
			NameChange: true, // name changed
		},
	}
}
