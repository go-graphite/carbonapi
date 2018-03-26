package countSeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type countSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &countSeries{}
	functions := []string{"countSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// countSeries(seriesList)
func (f *countSeries) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(civil): Check that series have equal length
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	r := *args[0]
	r.Name = fmt.Sprintf("countSeries(%s)", e.RawArgs())
	r.Values = make([]float64, len(args[0].Values))
	count := float64(len(args))

	for i := range args[0].Values {
		r.Values[i] = count
	}

	return []*types.MetricData{&r}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *countSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"countSeries": {
			Description: "Draws a horizontal line representing the number of nodes found in the seriesList.\n\n.. code-block:: none\n\n  &target=countSeries(carbon.agents.*.*)",
			Function:    "countSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "countSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Type:     types.SeriesList,
				},
			},
		},
	}
}
