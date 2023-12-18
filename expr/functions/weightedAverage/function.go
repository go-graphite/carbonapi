package weightedAverage

import (
	"context"
	"sort"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type weightedAverage struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &weightedAverage{}
	functions := []string{"weightedAverage"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// weightedAverage(seriesListAvg, seriesListWeight, *nodes)
func (f *weightedAverage) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	aggKeyPairs := make(map[string]map[string]*types.MetricData)
	var productList []*types.MetricData

	avgs, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	weights, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}

	// TODO: should fail if len(avgs) != len(weights)
	if len(avgs)+len(weights) == 0 {
		return []*types.MetricData{}, nil
	}

	alignedMetrics := helper.AlignSeries(append(avgs, weights...))
	avgs = alignedMetrics[0:len(avgs)]
	weights = alignedMetrics[len(avgs):]
	xFilesFactor := float64(alignedMetrics[0].XFilesFactor)

	nodes, err := e.GetNodeOrTagArgs(2, false)
	if err != nil {
		return nil, err
	}

	avgNames := make([]string, 0, len(avgs))
	weightNames := make([]string, 0, len(weights))

	for _, metric := range avgs {
		key := helper.AggKey(metric, nodes)
		if val, ok := aggKeyPairs[key]; !ok {
			// Normally, key shouldn't exist
			aggKeyPairs[key] = map[string]*types.MetricData{"avg": metric}
		} else {
			// According to graphite-web, this is overriden, so only the latest `key` is used
			val["avg"] = metric
		}
		avgNames = append(avgNames, metric.Name)
	}
	sort.Strings(avgNames)

	for _, metric := range weights {
		key := helper.AggKey(metric, nodes)
		if val, ok := aggKeyPairs[key]; !ok {
			// Normally, key shouldn't exist
			aggKeyPairs[key] = map[string]*types.MetricData{"weight": metric}
		} else {
			// According to graphite-web, this is overriden, so only the latest `key` is used
			val["weight"] = metric
		}
		weightNames = append(weightNames, metric.Name)
	}
	sort.Strings(weightNames)

	for _, pair := range aggKeyPairs {
		if _, ok := pair["avg"]; !ok {
			continue
		}
		if _, ok := pair["weight"]; !ok {
			continue
		}
		product, err := helper.AggregateSeries(e, []*types.MetricData{pair["avg"], pair["weight"]}, consolidations.ConsolidationToFunc["multiply"], xFilesFactor, false)
		if err != nil {
			return nil, err
		}
		productList = append(productList, product...)
	}
	if len(productList) == 0 {
		return []*types.MetricData{}, nil
	}

	sumProducts, err := helper.AggregateSeries(e, productList, consolidations.AggSum, xFilesFactor, false)
	if err != nil {
		return nil, err
	}
	sumWeights, err := helper.AggregateSeries(e, weights, consolidations.AggSum, xFilesFactor, false)
	if err != nil {
		return nil, err
	}
	weightedAverageSeries, err := helper.AggregateSeries(e, append(sumProducts, sumWeights...), func(v []float64) float64 { return v[0] / v[1] }, xFilesFactor, false)
	if err != nil {
		return nil, err
	}

	return weightedAverageSeries, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *weightedAverage) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"weightedAverage": {
			Function: "weightedAverage(seriesListAvg, seriesListWeight, *nodes)",
			Module:   "graphite.render.functions",
			Params: []types.FunctionParam{
				{
					Type:     types.SeriesList,
					Name:     "seriesListAvg",
					Required: true,
				},
				{
					Type:     types.SeriesList,
					Name:     "seriesListWeight",
					Required: true,
				},
				{
					Type:     types.NodeOrTag,
					Name:     "nodes",
					Multiple: true,
				},
			},
			Group:       "Combine",
			Description: "Takes a series of average values and a series of weights and\nproduces a weighted average for all values.\nThe corresponding values should share one or more zero-indexed nodes and/or tags.\n\nExample:\n\n.. code-block:: none\n\n  &target=weightedAverage(*.transactions.mean,*.transactions.count,0)\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.",
			Name:        "weightedAverage",
		},
	}
}
