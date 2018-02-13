package pearson

import (
	"container/heap"
	"errors"
	"fmt"
	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

func init() {
	metadata.RegisterFunction("pearson", &Pearson{})
	metadata.RegisterFunction("pearsonClosest", &PearsonClosest{})
}

type Pearson struct {
	interfaces.FunctionBase
}

// pearson(series, series, windowSize)
func (f *Pearson) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg1, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	arg2, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	if len(arg1) != 1 || len(arg2) != 1 {
		return nil, types.ErrWildcardNotAllowed
	}

	a1 := arg1[0]
	a2 := arg2[0]

	windowSize, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	w1 := &types.Windowed{Data: make([]float64, windowSize)}
	w2 := &types.Windowed{Data: make([]float64, windowSize)}

	r := *a1
	r.Name = fmt.Sprintf("pearson(%s,%s,%d)", a1.Name, a2.Name, windowSize)
	r.Values = make([]float64, len(a1.Values))
	r.IsAbsent = make([]bool, len(a1.Values))
	r.StartTime = from
	r.StopTime = until

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
			r.Values[i] = onlinestats.Pearson(w1.Data, w2.Data)
		} else {
			r.Values[i] = 0
			r.IsAbsent[i] = true
		}
	}

	return []*types.MetricData{&r}, nil
}

type PearsonClosest struct {
	interfaces.FunctionBase
}

// pearsonClosest(series, seriesList, n, direction=abs)
func (f *PearsonClosest) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if len(e.Args()) > 3 {
		return nil, types.ErrTooManyArguments
	}

	ref, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	if len(ref) != 1 {
		// TODO(nnuss) error("First argument must be single reference series")
		return nil, types.ErrWildcardNotAllowed
	}

	compare, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if err != nil {
		return nil, err
	}

	n, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	direction, err := e.GetStringNamedOrPosArgDefault("direction", 3, "abs")
	if err != nil {
		return nil, err
	}
	if direction != "pos" && direction != "neg" && direction != "abs" {
		return nil, errors.New("direction must be one of: pos, neg, abs")
	}

	// NOTE: if direction == "abs" && len(compare) <= n : we'll still do the work to rank them

	refValues := make([]float64, len(ref[0].Values))
	copy(refValues, ref[0].Values)
	for i, v := range ref[0].IsAbsent {
		if v {
			refValues[i] = math.NaN()
		}
	}

	var mh types.MetricHeap

	for index, a := range compare {
		compareValues := make([]float64, len(a.Values))
		copy(compareValues, a.Values)
		if len(refValues) != len(compareValues) {
			// Pearson will panic if arrays are not equal length; skip
			continue
		}
		for i, v := range a.IsAbsent {
			if v {
				compareValues[i] = math.NaN()
			}
		}
		value := onlinestats.Pearson(refValues, compareValues)
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
		heap.Push(&mh, types.MetricHeapElement{Idx: index, Val: value})
	}

	if n > len(mh) {
		n = len(mh)
	}
	results := make([]*types.MetricData, n)
	for i := range results {
		v := heap.Pop(&mh).(types.MetricHeapElement)
		results[i] = compare[v.Idx]
	}

	return results, nil
}
