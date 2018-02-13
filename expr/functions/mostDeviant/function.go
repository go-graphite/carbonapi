package mostDeviant

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
	metadata.RegisterFunction("mostDeviant", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// mostDeviant(seriesList, n) -or- mostDeviant(n, seriesList)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var nArg int
	if !e.Args()[0].IsConst() {
		// mostDeviant(seriesList, n)
		nArg = 1
	}
	seriesArg := nArg ^ 1 // XOR to make seriesArg the opposite argument. ( 0^1 -> 1 ; 1^1 -> 0 )

	n, err := e.GetIntArg(nArg)
	if err != nil {
		return nil, err
	}

	args, err := helper.GetSeriesArg(e.Args()[seriesArg], from, until, values)
	if err != nil {
		return nil, err
	}

	var mh types.MetricHeap

	for index, arg := range args {
		variance := helper.VarianceValue(arg.Values, arg.IsAbsent)
		if math.IsNaN(variance) {
			continue
		}

		if len(mh) < n {
			heap.Push(&mh, types.MetricHeapElement{Idx: index, Val: variance})
			continue
		}

		if variance > mh[0].Val {
			mh[0].Idx = index
			mh[0].Val = variance
			heap.Fix(&mh, 0)
		}
	}

	results := make([]*types.MetricData, len(mh))

	for len(mh) > 0 {
		v := heap.Pop(&mh).(types.MetricHeapElement)
		results[len(mh)] = args[v.Idx]
	}

	return results, err
}
