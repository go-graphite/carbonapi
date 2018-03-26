package highest

import (
	"container/heap"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type highest struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &highest{}
	functions := []string{"highestAverage", "highestCurrent", "highestMax"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)
func (f *highest) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {

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
	case "highestMax":
		compute = helper.MaxValue
	case "highestAverage":
		compute = helper.AvgValue
	case "highestCurrent":
		compute = helper.CurrentValue
	}

	for i, a := range arg {
		m := compute(a.Values)
		if math.IsNaN(m) {
			continue
		}

		if len(mh) < n {
			heap.Push(&mh, types.MetricHeapElement{Idx: i, Val: m})
			continue
		}
		// m is bigger than smallest max found so far
		if mh[0].Val < m {
			mh[0].Val = m
			mh[0].Idx = i
			heap.Fix(&mh, 0)
		}
	}

	results = make([]*types.MetricData, len(mh))

	// results should be ordered ascending
	for len(mh) > 0 {
		v := heap.Pop(&mh).(types.MetricHeapElement)
		results[len(mh)] = arg[v.Idx]
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *highest) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"highestAverage": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the top N metrics with the highest\naverage value for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=highestAverage(server*.instance*.threads.busy,5)\n\nDraws the top 5 servers with the highest average value.\n\nThis is an alias for :py:func:`highest <highest>` with aggregation ``average``.",
			Function:    "highestAverage(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "highestAverage",
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
		"highestCurrent": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the N metrics with the highest value\nat the end of the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=highestCurrent(server*.instance*.threads.busy,5)\n\nDraws the 5 servers with the highest busy threads.\n\nThis is an alias for :py:func:`highest <highest>` with aggregation ``current``.",
			Function:    "highestCurrent(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "highestCurrent",
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
		"highestMax": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\n\nOut of all metrics passed, draws only the N metrics with the highest maximum\nvalue in the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=highestMax(server*.instance*.threads.busy,5)\n\nDraws the top 5 servers who have had the most busy threads during the time\nperiod specified.\n\nThis is an alias for :py:func:`highest <highest>` with aggregation ``max``.",
			Function:    "highestMax(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "highestMax",
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
