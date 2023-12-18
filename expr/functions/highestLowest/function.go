package highestLowest

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
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
	functions := []string{"highestAverage", "highestCurrent", "highestMax", "highestMin", "highest", "lowestMax", "lowestMin", "lowestAverage", "lowestCurrent", "lowest"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)
func (f *highest) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	n := 1
	if e.ArgsLen() > 1 && e.Target() != "highest" && e.Target() != "lowest" {
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

	isHighest := strings.HasPrefix(e.Target(), "highest")
	switch e.Target() {
	case "highest", "lowest":
		consolidation := "average"
		switch e.ArgsLen() {
		case 2:
			n, err = e.GetIntArg(1)
			if err != nil {
				// We need to support case where only function specified
				n = 1
				consolidation, err = e.GetStringArgDefault(1, "average")
				if err != nil {
					return nil, err
				}
			}
		case 3:
			n, err = e.GetIntArg(1)

			if err != nil {
				return nil, err
			}
			consolidation, err = e.GetStringArgDefault(2, "average")

			if err != nil {
				return nil, err
			}
		}
		var ok bool
		compute, ok = consolidations.ConsolidationToFunc[consolidation]
		if !ok {
			return nil, fmt.Errorf("unsupported consolidation function %v", consolidation)
		}
	case "highestMax", "lowestMax":
		compute = consolidations.MaxValue
	case "highestAverage", "lowestAverage":
		compute = consolidations.AvgValue
	case "highestCurrent", "lowestCurrent":
		compute = consolidations.CurrentValue
	case "highestMin", "lowestMin":
		compute = consolidations.MinValue
	default:
		return nil, fmt.Errorf("unsupported function %v", e.Target())
	}

	if isHighest {
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
	} else {
		for i, a := range arg {
			m := compute(a.Values)
			heap.Push(&mh, types.MetricHeapElement{Idx: i, Val: m})
		}

		results = make([]*types.MetricData, n)

		for i := 0; i < n; i++ {
			v := heap.Pop(&mh).(types.MetricHeapElement)
			results[i] = arg[v.Idx]
		}
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *highest) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"highest": {
			Name:        "highest",
			Function:    "highest(seriesList, n=1, func='average')",
			Description: "Takes one metric or a wildcard seriesList followed by an integer N and an aggregation function.\nOut of all metrics passed, draws only the N metrics with the highest aggregated value over the\ntime period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=highest(server*.instance*.threads.busy,5,'max')\n\nDraws the 5 servers with the highest number of busy threads.",
			Module:      "graphite.render.functions",
			Group:       "Filter Series",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "n",
					Type:     types.Integer,
					Required: true,
				},
				{
					Name: "func",
					Type: types.String,
					Default: &types.Suggestion{
						Type:  types.SString,
						Value: "average",
					},
					Options: types.StringsToSuggestionList(consolidations.AvailableConsolidationFuncs()),
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
		},
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
			SeriesChange: true, // function aggregate metrics or change series items count
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
			SeriesChange: true, // function aggregate metrics or change series items count
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
		"highestMin": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\n\nOut of all metrics passed, draws only the N metrics with the highest minimum\nvalue in the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=highestMin(server*.instance*.threads.busy,5)\n\nDraws the top 5 servers who have had the highest minimum of all the values during the time\nperiod specified.\n\nThis is an alias for :py:func:`highest <highest>` with aggregation ``min``.",
			Function:    "highestMin(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "highestMin",
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
		"lowest": {
			Name:        "lowest",
			Function:    "lowest(seriesList, n=1, func='average')",
			Description: "Takes one metric or a wildcard seriesList followed by an integer N and an aggregation function.\nOut of all metrics passed, draws only the N metrics with the lowest aggregated value over the\ntime period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=lowest(server*.instance*.threads.busy,5,'min')\n\nDraws the 5 servers with the lowest number of busy threads.",
			Module:      "graphite.render.functions",
			Group:       "Filter Series",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "n",
					Type:     types.Integer,
					Required: true,
				},
				{
					Name: "func",
					Type: types.String,
					Default: &types.Suggestion{
						Type:  types.SString,
						Value: "average",
					},
					Options: types.StringsToSuggestionList(consolidations.AvailableConsolidationFuncs()),
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
		"lowestMax": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the bottom N metrics with the lowest\nmaximum value for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=lowestMax(server*.instance*.threads.busy,5)\n\nDraws the bottom 5 servers with the lowest maximum value.\n\nThis is an alias for :py:func:`lowest <lowest>` with aggregation ``max``.",
			Function:    "lowestMax(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "lowestMax",
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
		"lowestMin": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the bottom N metrics with the lowest\nminimum value for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=lowestMin(server*.instance*.threads.busy,5)\n\nDraws the bottom 5 servers with the lowest minimum value.\n\nThis is an alias for :py:func:`lowest <lowest>` with aggregation ``min``.",
			Function:    "lowestMin(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "lowestMin",
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
			SeriesChange: true, // function aggregate metrics or change series items count
		},
	}
}
