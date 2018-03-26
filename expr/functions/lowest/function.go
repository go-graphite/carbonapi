package lowest

import (
	"container/heap"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type lowest struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &lowest{}
	functions := []string{"lowestAverage", "lowestCurrent"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// lowestAverage(seriesList, n) , lowestCurrent(seriesList, n)
func (f *lowest) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	n := 1
	if len(e.Args()) > 1 {
		n, err = e.GetIntArg(1)
		if err != nil {
			return nil, err
		}
	}
	var results []*types.MetricData

	// we have fewer arguments than we want result series
	if len(arg) < n {
		return arg, nil
	}

	var mh types.MetricHeap

	var compute func([]float64) float64

	switch e.Target() {
	case "lowestAverage":
		compute = helper.AvgValue
	case "lowestCurrent":
		compute = helper.CurrentValue
	}

	for i, a := range arg {
		m := compute(a.Values)
		heap.Push(&mh, types.MetricHeapElement{Idx: i, Val: m})
	}

	results = make([]*types.MetricData, n)

	// results should be ordered ascending
	for i := 0; i < n; i++ {
		v := heap.Pop(&mh).(types.MetricHeapElement)
		results[i] = arg[v.Idx]
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *lowest) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"lowestAverage": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the bottom N metrics with the lowest\naverage value for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=lowestAverage(server*.instance*.threads.busy,5)\n\nDraws the bottom 5 servers with the lowest average value.\n\nThis is an alias for :py:func:`lowest <lowest>` with aggregation ``average``.",
			Function:    "lowestAverage(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "lowestAverage",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"lowestCurrent": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the N metrics with the lowest value at\nthe end of the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=lowestCurrent(server*.instance*.threads.busy,5)\n\nDraws the 5 servers with the least busy threads right now.\n\nThis is an alias for :py:func:`lowest <lowest>` with aggregation ``current``.",
			Function:    "lowestCurrent(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "lowestCurrent",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
	}
}
