package seriesByTag

import (
	"context"

	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type seriesByTag struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &seriesByTag{}
	for _, n := range []string{"seriesByTag"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *seriesByTag) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var results []*types.MetricData
	key := parser.MetricRequest{Metric: e.ToString(), From: from, Until: until}
	data, ok := values[key]
	if !ok {
		return results, nil
	}
	return data, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *seriesByTag) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"seriesByTag": {
			Description: "Returns a SeriesList of series matching all the specified tag expressions.\n\nExample:\n\n.. code-block:: none\n\n  &target=seriesByTag(\"tag1=value1\",\"tag2!=value2\")\n\nReturns a seriesList of all series that have tag1 set to value1, AND do not have tag2 set to value2.\n\nTags specifiers are strings, and may have the following formats:\n\n.. code-block:: none\n\n  tag=spec    tag value exactly matches spec\n  tag!=spec   tag value does not exactly match spec\n  tag=~value  tag value matches the regular expression spec\n  tag!=~spec  tag value does not match the regular expression spec\n\nAny tag spec that matches an empty value is considered to match series that don't have that tag.\n\nAt least one tag spec must require a non-empty value.\n\nRegular expression conditions are treated as being anchored at the start of the value.\n\nSee :ref:`querying tagged series <querying-tagged-series>` for more detail.",
			Function:    "seriesByTag(*tagExpressions)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "seriesByTag",
			Params: []types.FunctionParam{
				{
					Name:     "tagExpressions",
					Required: true,
					Type:     types.String,
					Multiple: true,
				},
			},
		},
	}
}
