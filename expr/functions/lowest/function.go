package lowest

import (
	"container/heap"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("lowestAverage", &Function{})
	metadata.RegisterFunction("lowestCurrent", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// lowestAverage(seriesList, n) , lowestCurrent(seriesList, n)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

	var compute func([]float64, []bool) float64

	switch e.Target() {
	case "lowestAverage":
		compute = helper.AvgValue
	case "lowestCurrent":
		compute = helper.CurrentValue
	}

	for i, a := range arg {
		m := compute(a.Values, a.IsAbsent)
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
