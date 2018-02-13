package highest

import (
	"container/heap"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	functions := []string{"highestAverage", "highestCurrent", "highestMax"}
	for _, f := range functions {
		metadata.RegisterFunction(f, &Function{})
	}
}

type Function struct {
	interfaces.FunctionBase
}

// highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)
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
	case "highestMax":
		compute = helper.MaxValue
	case "highestAverage":
		compute = helper.AvgValue
	case "highestCurrent":
		compute = helper.CurrentValue
	}

	for i, a := range arg {
		m := compute(a.Values, a.IsAbsent)
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
