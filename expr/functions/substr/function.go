package substr

import (
	"context"
	"errors"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/helper/metric"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type substr struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &substr{}
	functions := []string{"substr"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// aliasSub(seriesList, start, stop)
func (f *substr) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// BUG: affected by the same positional arg issue as 'threshold'.
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	startField, err := e.GetIntNamedOrPosArgDefault("start", 1, 0)
	if err != nil {
		return nil, err
	}

	stopField, err := e.GetIntNamedOrPosArgDefault("stop", 2, 0)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		metric := metric.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")
		realStartField := startField
		if startField != 0 {
			if startField < 0 {
				realStartField = len(nodes) + startField
				if realStartField < 0 {
					realStartField = 0
				}
			}
			if realStartField > len(nodes)-1 {
				return nil, errors.New("start out of range")
			}
			nodes = nodes[realStartField:]
		}
		if stopField != 0 && stopField < len(nodes)+realStartField {
			realStopField := stopField
			if stopField < 0 {
				realStopField = len(nodes) + stopField
			} else {
				realStopField = realStopField - realStartField
			}
			if realStopField < 0 {
				return nil, errors.New("stop out of range")
			}
			nodes = nodes[:realStopField]
		}

		r := *a
		r.Name = strings.Join(nodes, ".")
		results = append(results, &r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *substr) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"substr": {
			Description: "Takes one metric or a wildcard seriesList followed by 1 or 2 integers.  Assume that the\nmetric name is a list or array, with each element separated by dots.  Prints\nn - length elements of the array (if only one integer n is passed) or n - m\nelements of the array (if two integers n and m are passed).  The list starts\nwith element 0 and ends with element (length - 1).\n\nExample:\n\n.. code-block:: none\n\n  &target=substr(carbon.agents.hostname.avgUpdateTime,2,4)\n\nThe label would be printed as \"hostname.avgUpdateTime\".",
			Function:    "substr(seriesList, start=0, stop=0)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "substr",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "start",
					Type:    types.Node,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "stop",
					Type:    types.Node,
				},
			},
		},
	}
}
