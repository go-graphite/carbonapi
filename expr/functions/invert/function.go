package invert

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("invert", &invert{})
}

type invert struct {
	interfaces.FunctionBase
}

// invert(seriesList)
func (f *invert) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		for i, v := range a.Values {
			if a.IsAbsent[i] || v == 0 {
				r.Values[i] = 0
				r.IsAbsent[i] = true
				continue
			}
			r.Values[i] = 1 / v
		}
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *invert) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"invert": {
			Description: "Takes one metric or a wildcard seriesList, and inverts each datapoint (i.e. 1/x).\n\nExample:\n\n.. code-block:: none\n\n  &target=invert(Server.instance01.threads.busy)",
			Function:    "invert(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "invert",
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