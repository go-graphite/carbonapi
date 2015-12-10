package main

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/dgryski/go-onlinestats"
	"github.com/gogo/protobuf/proto"
)

// pearson(series, series, windowSize)
func pearson(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg1, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	arg2, err := getSeriesArg(e.args[1], from, until, values)
	if err != nil {
		return nil
	}

	if len(arg1) != 1 || len(arg2) != 1 {
		// must be single series
		return nil
	}

	a1 := arg1[0]
	a2 := arg2[0]

	windowSize, err := getIntArg(e, 2)
	if err != nil {
		return nil
	}

	w1 := &Windowed{data: make([]float64, windowSize)}
	w2 := &Windowed{data: make([]float64, windowSize)}

	r := *a1
	r.Name = proto.String(fmt.Sprintf("pearson(%s,%s,%d)", a1.GetName(), a2.GetName(), windowSize))
	r.Values = make([]float64, len(a1.Values))
	r.IsAbsent = make([]bool, len(a1.Values))
	r.StartTime = proto.Int32(from)
	r.StopTime = proto.Int32(until)

	for i, v1 := range a1.Values {
		v2 := a2.Values[i]
		if a1.IsAbsent[i] || a2.IsAbsent[i] {
			// ignore if either is missing
			v1 = math.NaN()
			v2 = math.NaN()
		}
		w1.Push(v1)
		w2.Push(v2)
		if i >= windowSize-1 {
			r.Values[i] = onlinestats.Pearson(w1.data, w2.data)
		} else {
			r.Values[i] = 0
			r.IsAbsent[i] = true
		}
	}

	return []*metricData{&r}
}

// pearsonClosest(series, seriesList, n, direction=abs)
func pearsonClosest(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	ref, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	if len(ref) != 1 {
		// TODO(nnuss) error("First argument must be single reference series")
		return nil
	}

	compare, err := getSeriesArg(e.args[1], from, until, values)
	if err != nil {
		return nil
	}

	n, err := getIntArg(e, 2)
	if err != nil {
		return nil
	}

	direction, err := getStringArgDefault(e, 3, "abs")
	if err != nil && len(e.args) > 3 {
		return nil
	}
	if direction != "pos" && direction != "neg" && direction != "abs" {
		// TODO(nnuss) error("pearsonClosest( _ , _ , direction=abs ) : direction must be one of { 'pos', 'neg', 'abs' }")
		return nil
	}

	// NOTE: if direction == "abs" && len(compare) <= n : we'll still do the work to rank them

	for i, v := range ref[0].IsAbsent {
		if v == true {
			ref[0].Values[i] = math.NaN()
		}
	}

	var mh metricHeap

	for index, a := range compare {
		if len(ref[0].Values) != len(a.Values) {
			// Pearson will panic if arrays are not equal length; skip
			continue
		}
		for i, v := range a.IsAbsent {
			if v == true {
				a.Values[i] = math.NaN()
			}
		}
		value := onlinestats.Pearson(ref[0].Values, a.Values)
		// Standardize the value so sort ASC will have strongest correlation first
		switch {
		case math.IsNaN(value):
			// special case of at least one series containing all zeros which leads to div-by-zero in Pearson
			continue
		case direction == "abs":
			value = math.Abs(value) * -1
		case direction == "pos" && value >= 0:
			value = value * -1
		case direction == "neg" && value <= 0:
		default:
			continue
		}
		heap.Push(&mh, metricHeapElement{idx: index, val: value})
	}

	if n > len(mh) {
		n = len(mh)
	}
	results := make([]*metricData, n)
	for i := range results {
		v := heap.Pop(&mh).(metricHeapElement)
		results[i] = compare[v.idx]
	}

	return results
}
