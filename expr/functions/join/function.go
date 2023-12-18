package join

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

const (
	and = "AND"
	or  = "OR"
	xor = "XOR"
	sub = "SUB"
)

type join struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{
		{F: &join{}, Name: "join"},
	}
}

func (f *join) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"join": {
			Description: `Performs set operations on 'seriesA' and 'seriesB'. Following options are available:
 * AND - returns those metrics from 'seriesA' which are presented in 'seriesB';
 * OR  - returns all metrics from 'seriesA' and also those metrics from 'seriesB' which aren't presented in 'seriesA';
 * XOR - returns only those metrics which are presented in either 'seriesA' or 'seriesB', but not in both;
 * SUB - returns those metrics from 'seriesA' which aren't presented in 'seriesB';

Example:

.. code-block:: none

  &target=join(some.data.series.aaa, some.other.series.bbb, 'AND')`,
			Function: "join(seriesA, seriesB)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "join",
			Params: []types.FunctionParam{
				{
					Name:     "seriesA",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesB",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "type",
					Required: false,
					Type:     types.String,
					Default:  types.NewSuggestion(and),
					Options:  types.StringsToSuggestionList([]string{and, or, xor, sub}),
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}

func (f *join) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	seriesA, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	seriesB, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}
	joinType, err := e.GetStringNamedOrPosArgDefault("type", 2, and)
	if err != nil {
		return nil, err
	}
	joinType = strings.ToUpper(joinType)

	switch joinType {
	case and:
		return doAnd(seriesA, seriesB), nil
	case or:
		return doOr(seriesA, seriesB), nil
	case xor:
		return doXor(seriesA, seriesB), nil
	case sub:
		return doSub(seriesA, seriesB), nil
	default:
		return nil, fmt.Errorf("unknown join type: %s", joinType)
	}
}

func doAnd(seriesA []*types.MetricData, seriesB []*types.MetricData) (results []*types.MetricData) {
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[md.Name] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA))
	for _, md := range seriesA {
		if metricsB[md.Name] {
			results = append(results, md)
		}
	}
	return results
}

func doOr(seriesA []*types.MetricData, seriesB []*types.MetricData) (results []*types.MetricData) {
	metricsA := make(map[string]bool, len(seriesA))
	for _, md := range seriesA {
		metricsA[md.Name] = true
	}

	results = seriesA
	for _, md := range seriesB {
		if !metricsA[md.Name] {
			results = append(results, md)
		}
	}
	return results
}

func doXor(seriesA []*types.MetricData, seriesB []*types.MetricData) (results []*types.MetricData) {
	metricsA := make(map[string]bool, len(seriesA))
	for _, md := range seriesA {
		metricsA[md.Name] = true
	}
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[md.Name] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA)+len(seriesB))
	for _, md := range seriesA {
		if !metricsB[md.Name] {
			results = append(results, md)
		}
	}
	for _, md := range seriesB {
		if !metricsA[md.Name] {
			results = append(results, md)
		}
	}
	return results
}

func doSub(seriesA []*types.MetricData, seriesB []*types.MetricData) (results []*types.MetricData) {
	metricsB := make(map[string]bool, len(seriesB))
	for _, md := range seriesB {
		metricsB[md.Name] = true
	}

	results = make([]*types.MetricData, 0, len(seriesA))
	for _, md := range seriesA {
		if !metricsB[md.Name] {
			results = append(results, md)
		}
	}
	return results
}
