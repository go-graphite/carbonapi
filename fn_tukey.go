package main

import (
	"container/heap"
	"sort"
	"strings"
)

// tukeyAbove(seriesList,basis,n,interval=0) , tukeyBelow(seriesList,basis,n,interval=0)
func tukey(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {

	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	basis, err := getFloatArg(e, 1)
	if err != nil || basis <= 0 {
		return nil
	}

	n, err := getIntArg(e, 2)
	if err != nil || n < 1 {
		return nil
	}

	var interval int = 0
	if len(e.args) >= 4 {
		switch e.args[3].etype {
		case etConst:
			interval, err = getIntArg(e, 3)
		case etString:
			var i32 int32
			i32, err = getIntervalArg(e, 3, 1)
			interval = int(i32)
			interval /= int(arg[0].GetStepTime())
			// TODO(nnuss): make sure the arrays are all the same 'size'
		default:
			err = ErrBadType
		}
		if err != nil {
			return nil
		}
	}
	// TODO(nnuss): negative intervals

	// gather all the valid points
	var points []float64
	for _, a := range arg {
		for i, m := range a.Values {
			if a.IsAbsent[i] {
				continue
			}
			points = append(points, m)
		}
	}

	sort.Float64s(points)

	first := int(0.25 * float64(len(points)))
	third := int(0.75 * float64(len(points)))

	iqr := points[third] - points[first]

	max := points[third] + basis*iqr
	min := points[first] - basis*iqr

	isAbove := strings.HasSuffix(e.target, "Above")

	var mh metricHeap

	// count how many points are above the threshold
	for i, a := range arg {
		var outlier int
		for i, m := range a.Values {
			if a.IsAbsent[i] {
				continue
			}
			if isAbove {
				if m >= max {
					outlier++
				}
			} else {
				if m <= min {
					outlier++
				}
			}
		}

		// not even a single anomalous point -- ignore this metric
		if outlier == 0 {
			continue
		}

		if len(mh) < n {
			heap.Push(&mh, metricHeapElement{idx: i, val: float64(outlier)})
			continue
		}
		// current outlier count is is bigger than smallest max found so far
		foutlier := float64(outlier)
		if mh[0].val < foutlier {
			mh[0].val = foutlier
			mh[0].idx = i
			heap.Fix(&mh, 0)
		}
	}

	if len(mh) < n {
		n = len(mh)
	}
	results := make([]*metricData, n)
	// results should be ordered ascending
	for len(mh) > 0 {
		v := heap.Pop(&mh).(metricHeapElement)
		results[len(mh)] = arg[v.idx]
	}

	return results
}
