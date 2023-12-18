package aggregateSeriesLists

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type aggregateSeriesLists struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	f := &aggregateSeriesLists{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aggregateSeriesLists"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aggregateSeriesLists) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 3 {
		return nil, parser.ErrMissingArgument
	}

	seriesList1, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	seriesList2, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}

	if len(seriesList1) != len(seriesList2) {
		return nil, fmt.Errorf("seriesListFirstPos and seriesListSecondPos must have equal length")
	} else if len(seriesList1) == 0 {
		return make([]*types.MetricData, 0, 0), nil
	}

	aggFuncStr, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}
	aggFunc, ok := consolidations.ConsolidationToFunc[aggFuncStr]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", aggFuncStr)
	}

	xFilesFactor, err := e.GetFloatArgDefault(3, float64(seriesList1[0].XFilesFactor))
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(seriesList1))

	for i, series1 := range seriesList1 {
		series2 := seriesList2[i]

		r, err := helper.AggregateSeries(e, []*types.MetricData{series1.CopyLink(), series2.CopyLink()}, aggFunc, xFilesFactor, false)
		if err != nil {
			return nil, err
		}

		results = append(results, r[0])
	}

	return results, nil
}

func (f *aggregateSeriesLists) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aggregateSeriesLists": {
			Name:        "aggregateSeriesLists",
			Function:    "aggregateSeriesLists(seriesListFirstPos, seriesListSecondPos, func, xFilesFactor=None)",
			Description: "Iterates over a two lists and aggregates using specified function list1[0] to list2[0], list1[1] to list2[1] and so on. The lists will need to be the same length.",
			Module:      "graphite.render.functions",
			Group:       "Combine",
			Params: []types.FunctionParam{
				{
					Name:     "seriesListFirstPos",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "seriesListSecondPos",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "func",
					Type:     types.AggFunc,
					Required: true,
				},
				{
					Name:     "xFilesFactor",
					Type:     types.Float,
					Required: false,
				},
			},
		},
	}
}
