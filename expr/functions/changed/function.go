package changed

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type changed struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &changed{}
	functions := []string{"changed"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// changed(SeriesList)
func (f *changed) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	result := make([]*types.MetricData, len(args))
	for i, a := range args {
		r := a.CopyTag(e.Target()+"("+a.Name+")", a.Tags)
		r.Values = make([]float64, len(a.Values))

		prev := math.NaN()
		for i, v := range a.Values {
			if math.IsNaN(prev) {
				prev = v
				r.Values[i] = 0
			} else if !math.IsNaN(v) && prev != v {
				r.Values[i] = 1
				prev = v
			} else {
				r.Values[i] = 0
			}
		}
		result[i] = r
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *changed) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"changed": {
			Description: "Takes one metric or a wildcard seriesList.\nOutput 1 when the value changed, 0 when null or the same\n\nExample:\n\n.. code-block:: none\n\n  &target=changed(Server01.connections.handled)",
			Function:    "changed(seriesList)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "changed",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
