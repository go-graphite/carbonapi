package offset

import (
	"context"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type offset struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &offset{}
	functions := []string{"add", "offset"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// offset(seriesList,factor)
func (f *offset) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	factor, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	factorStr := strconv.FormatFloat(factor, 'g', -1, 64)
	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLink()
		r.Name = e.Target() + "(" + a.Name + "," + factorStr + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags[e.Target()] = factorStr

		for i, v := range a.Values {
			r.Values[i] = v + factor
		}
		results[i] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *offset) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"add": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and adds the constant to\neach datapoint.\n\nExample:\n\n.. code-block:: none\n\n  &target=add(Server.instance01.threads.busy,10)\n  &target=add(Server.instance*.threads.busy, 10)",
			Function:    "add(seriesList, constant)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "add",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "constant",
					Required: true,
					Type:     types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"offset": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and adds the constant to\neach datapoint.\n\nExample:\n\n.. code-block:: none\n\n  &target=offset(Server.instance01.threads.busy,10)",
			Function:    "offset(seriesList, factor)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "offset",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "factor",
					Required: true,
					Type:     types.Float,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
