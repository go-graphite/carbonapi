package logit

import (
	"context"
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type logit struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &logit{}
	functions := []string{"logit"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// logarithm(seriesList, base=10)
// Alias: log
func (f *logit) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range arg {
		r := a.CopyLink()
		r.Name = fmt.Sprintf("logit(%s)", a.Name)
		r.Values = make([]float64, len(a.Values))
		r.Tags["logit"] = "logit"

		for i, v := range a.Values {
			if math.IsNaN(v) || v == 1 {
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = math.Log(v / (1 - v))
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *logit) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"logit": {
			Description: "Takes one metric or a wildcard seriesList and applies the logit function log(x / (1 - x)) to each datapoint.\n\nExample:\n\n.. code-block:: none\n\n  &target=logit(Server.instance01.threads.busy)\n&target=logit(Server.instance*.threads.busy)",
			Function:    "logit(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "logit",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
