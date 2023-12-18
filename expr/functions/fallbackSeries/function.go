package fallbackSeries

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type fallbackSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &fallbackSeries{}
	functions := []string{"fallbackSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// fallbackSeries( seriesList, fallback )
func (f *fallbackSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/*
		Takes a wildcard seriesList, and a second fallback metric.
		If the wildcard does not match any series, draws the fallback metric.
	*/
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingTimeseries
	}

	seriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	fallback, errFallback := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if errFallback != nil && err != nil {
		return nil, errFallback
	}

	if len(seriesList) > 0 {
		return seriesList, nil
	}
	return fallback, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *fallbackSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"fallbackSeries": {
			Description: "Takes a wildcard seriesList, and a second fallback metric.\nIf the wildcard does not match any series, draws the fallback metric.\n\nExample:\n\n.. code-block:: none\n\n  &target=fallbackSeries(server*.requests_per_second, constantLine(0))\n\nDraws a 0 line when server metric does not exist.",
			Function:    "fallbackSeries(seriesList, fallback)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "fallbackSeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "fallback",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
