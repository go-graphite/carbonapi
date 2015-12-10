package main

import (
	"container/heap"
	"math"
)

func lowest(fn summarizeFunc) func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
		return lowestFunc(e, from, until, values, fn)
	}
}

func highest(fn summarizeFunc) func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	return func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
		return highestFunc(e, from, until, values, fn)
	}
}

// lowestAverage(seriesList, n) , lowestCurrent(seriesList, n)
func lowestFunc(e *expr, from, until int32, values map[metricRequest][]*metricData, compute summarizeFunc) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	n, err := getIntArg(e, 1)
	if err != nil {
		return nil
	}
	var results []*metricData

	// we have fewer arguments than we want result series
	if len(arg) < n {
		return arg
	}

	var mh metricHeap

	for i, a := range arg {
		m := compute(a.Values, a.IsAbsent)
		heap.Push(&mh, metricHeapElement{idx: i, val: m})
	}

	results = make([]*metricData, n)

	// results should be ordered ascending
	for i := 0; i < n; i++ {
		v := heap.Pop(&mh).(metricHeapElement)
		results[i] = arg[v.idx]
	}

	return results
}

// highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)
func highestFunc(e *expr, from, until int32, values map[metricRequest][]*metricData, compute summarizeFunc) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	n, err := getIntArg(e, 1)
	if err != nil {
		return nil
	}
	var results []*metricData

	// we have fewer arguments than we want result series
	if len(arg) < n {
		return arg
	}

	var mh metricHeap

	for i, a := range arg {
		m := compute(a.Values, a.IsAbsent)
		if math.IsNaN(m) {
			continue
		}

		if len(mh) < n {
			heap.Push(&mh, metricHeapElement{idx: i, val: m})
			continue
		}
		// m is bigger than smallest max found so far
		if mh[0].val < m {
			mh[0].val = m
			mh[0].idx = i
			heap.Fix(&mh, 0)
		}
	}

	results = make([]*metricData, n)

	// results should be ordered ascending
	for len(mh) > 0 {
		v := heap.Pop(&mh).(metricHeapElement)
		results[len(mh)] = arg[v.idx]
	}

	return results
}
