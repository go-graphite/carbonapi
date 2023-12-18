package isNotNull

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type isNotNull struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &isNotNull{}
	functions := []string{"isNotNull", "isNonNull"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// isNonNull(seriesList)
// alias: isNotNull(seriesList)
func (f *isNotNull) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	e.SetTarget("isNonNull")

	return helper.ForEachSeriesDo(ctx, f.GetEvaluator(), e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = 0
			} else {
				r.Values[i] = 1
			}

		}
		r.Tags["isNonNull"] = "1"
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *isNotNull) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"isNotNull": {
			Description: "Takes a metric or wildcard seriesList and counts up the number of non-null\nvalues.  This is useful for understanding the number of metrics that have data\nat a given point in time (i.e. to count which servers are alive).\n\nExample:\n\n.. code-block:: none\n\n  &target=isNotNull(webapp.pages.*.views)\n\nReturns a seriesList where 1 is specified for non-null values, and\n0 is specified for null values.",
			Function:    "isNotNull(seriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "isNotNull",
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
		"isNonNull": {
			Description: "Takes a metric or wildcard seriesList and counts up the number of non-null\nvalues.  This is useful for understanding the number of metrics that have data\nat a given point in time (i.e. to count which servers are alive).\n\nExample:\n\n.. code-block:: none\n\n  &target=isNonNull(webapp.pages.*.views)\n\nReturns a seriesList where 1 is specified for non-null values, and\n0 is specified for null values.",
			Function:    "isNonNull(seriesList)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "isNonNull",
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
