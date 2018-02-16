package aliasByMetric

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"strings"
)

func init() {
	metadata.RegisterFunction("aliasByMetric", &aliasByMetric{})
}

type aliasByMetric struct {
	interfaces.FunctionBase
}

func (f *aliasByMetric) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		metric := helper.ExtractMetric(a.Name)
		part := strings.Split(metric, ".")
		r.Name = part[len(part)-1]
		r.Values = a.Values
		r.IsAbsent = a.IsAbsent
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aliasByMetric) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"aliasByMetric": {
			Description: "Takes a seriesList and applies an alias derived from the base metric name.\n\n.. code-block:: none\n\n  &target=aliasByMetric(carbon.agents.graphite.creates)",
			Function:    "aliasByMetric(seriesList)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByMetric",
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