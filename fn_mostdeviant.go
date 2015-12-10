package main

import (
	"container/heap"
	"math"
)

// mostDeviant(n, seriesList)
func mostDeviant(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	n, err := getIntArg(e, 0)
	if err != nil {
		return nil
	}

	args, err := getSeriesArg(e.args[1], from, until, values)
	if err != nil {
		return nil
	}

	var mh metricHeap

	for index, arg := range args {
		variance := varianceValue(arg.Values, arg.IsAbsent)
		if math.IsNaN(variance) {
			continue
		}

		if len(mh) < n {
			heap.Push(&mh, metricHeapElement{idx: index, val: variance})
			continue
		}

		if variance > mh[0].val {
			mh[0].idx = index
			mh[0].val = variance
			heap.Fix(&mh, 0)
		}
	}

	results := make([]*metricData, n)

	for len(mh) > 0 {
		v := heap.Pop(&mh).(metricHeapElement)
		results[len(mh)] = args[v.idx]
	}

	return results
}
