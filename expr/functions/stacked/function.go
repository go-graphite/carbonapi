package stacked

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type stacked struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &stacked{}
	functions := []string{"stacked"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *stacked) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	stackName, err := e.GetStringNamedOrPosArgDefault("stackname", 1, types.DefaultStackName)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(arg))

	for i, a := range arg {
		r := a.CopyLinkTags()
		r.Stacked = true
		r.StackName = stackName
		r.Tags["stacked"] = stackName
		results[i] = r
	}

	return results, nil
}

func (f *stacked) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"stacked": {
			Name: "stacked",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "stack",
					Type: types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList and change them so they are\nstacked. This is a way of stacking just a couple of metrics without having\nto use the stacked area mode (that stacks everything). By means of this a mixed\nstacked and non stacked graph can be made\n\nIt can also take an optional argument with a name of the stack, in case there is\nmore than one, e.g. for input and output metrics.\n\nExample:\n\n.. code-block:: none\n\n  &target=stacked(company.server.application01.ifconfig.TXPackets, 'tx')",
			Function:    "stacked(seriesLists, stackName='__DEFAULT__')",
			Group:       "Graph",
		},
	}
}
