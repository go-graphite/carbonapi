package sortBy

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"sort"
)

type sortBy struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sortBy{}
	functions := []string{"sortByMaxima", "sortByMinima", "sortByTotal"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sortByMaxima(seriesList), sortByMinima(seriesList), sortByTotal(seriesList)
func (f *sortBy) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	original, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	arg := make([]*types.MetricData, len(original))
	copy(arg, original)
	vals := make([]float64, len(arg))

	for i, a := range arg {
		switch e.Target() {
		case "sortByTotal":
			vals[i] = helper.SummarizeValues("sum", a.Values)
		case "sortByMaxima":
			vals[i] = helper.SummarizeValues("max", a.Values)
		case "sortByMinima":
			vals[i] = 1 / helper.SummarizeValues("min", a.Values)
		}
	}

	sort.Sort(helper.ByVals{Vals: vals, Series: arg})

	return arg, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sortBy) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"sortByMaxima": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics in descending order by the maximum value across the time period\nspecified.  Useful with the &areaMode=all parameter, to keep the\nlowest value lines visible.\n\nExample:\n\n.. code-block:: none\n\n  &target=sortByMaxima(server*.instance*.memory.free)",
			Function:    "sortByMaxima(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByMaxima",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sortByMinima": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics by the lowest value across the time period\nspecified, including only series that have a maximum value greater than 0.\n\nExample:\n\n.. code-block:: none\n\n  &target=sortByMinima(server*.instance*.memory.free)",
			Function:    "sortByMinima(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByMinima",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
		"sortByTotal": {
			Description: "Takes one metric or a wildcard seriesList.\n\nSorts the list of metrics in descending order by the sum of values across the time period\nspecified.",
			Function:    "sortByTotal(seriesList)",
			Group:       "Sorting",
			Module:      "graphite.render.functions",
			Name:        "sortByTotal",
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
