package scale

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type scale struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &scale{}
	functions := []string{"scale", "scaleAfterTimestamp"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// scale(seriesList, factor)
func (f *scale) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	scale, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	timestamp, err := e.GetIntArgDefault(2, 0)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(arg))
	for _, a := range arg {
		r := *a
		if timestamp == 0 {
			r.Name = fmt.Sprintf("scale(%s,%g)", a.Name, scale)
		} else {
			r.Name = fmt.Sprintf("scale(%s,%g,%d)", a.Name, scale, timestamp)
		}
		r.Values = make([]float64, len(a.Values))

		currentTimestamp := a.StartTime
		for i, v := range a.Values {
			r.Values[i] = v
			if currentTimestamp >= int64(timestamp) {
				r.Values[i] *= scale
			}

			currentTimestamp += a.StepTime
		}
		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *scale) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"scale": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and multiplies the datapoint\n" +
				"by the constant provided at each point.\n" +
				"carbonapi extends this function by optional 3-rd parameter that accepts unix-timestamp. If provided, only values with timestamp newer than it will be scaled\n\n" +
				"Example:\n\n.. code-block:: none\n\n  &target=scale(Server.instance01.threads.busy,10)\n  &target=scale(Server.instance*.threads.busy,10)\n\n" +
				"Alias: scaleAfterTimestamp",
			Function: "scale(seriesList, factor)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "scale",
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
				{
					Name:     "timestamp",
					Required: false,
					Type:     types.Integer,
					Default:  types.NewSuggestion(0),
				},
			},
		},
		"scaleAfterTimestamp": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant, and multiplies the datapoint\n" +
				"by the constant provided at each point.\n" +
				"carbonapi extends this function by optional 3-rd parameter that accepts unix-timestamp. If provided, only values with timestamp newer than it will be scaled\n\n" +
				"Example:\n\n.. code-block:: none\n\n  &target=scale(Server.instance01.threads.busy,10)\n  &target=scale(Server.instance*.threads.busy,10)",
			Function: "scale(seriesList, factor)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "scale",
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
				{
					Name:     "timestamp",
					Required: false,
					Type:     types.Integer,
					Default:  types.NewSuggestion(0),
				},
			},
		},
	}
}
