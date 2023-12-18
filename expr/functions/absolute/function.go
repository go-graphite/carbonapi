package absolute

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type absolute struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &absolute{}
	for _, n := range []string{"absolute"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *absolute) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(ctx, f.GetEvaluator(), e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		for i, v := range a.Values {
			if math.IsNaN(a.Values[i]) {
				r.Values[i] = math.NaN()
				continue
			}
			r.Values[i] = math.Abs(v)
		}
		r.Tags["absolute"] = "1"
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *absolute) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"absolute": {
			Description: "Takes one metric or a wildcard seriesList and applies the mathematical abs function to each\ndatapoint transforming it to its absolute value.\n\nExample:\n\n.. code-block:: none\n\n  &target=absolute(Server.instance01.threads.busy)\n  &target=absolute(Server.instance*.threads.busy)",
			Function:    "absolute(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "absolute",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			ValuesChange: true,
		},
	}
}
