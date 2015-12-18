package main

import (
	"fmt"
	"math"

	"github.com/dgryski/go-onlinestats"
	"github.com/gogo/protobuf/proto"
)

// ksTest2(series, series, points|"interval")
func ksTest(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	arg1, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	arg2, err := getSeriesArg(e.args[1], from, until, values)
	if err != nil {
		return nil
	}

	if len(arg1) != 1 || len(arg2) != 1 {
		// no wildcards allowed
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
	r.Name = proto.String(fmt.Sprintf("kolmogorovSmirnovTest2(%s,%s,%d)", a1.GetName(), a2.GetName(), windowSize))
	r.Values = make([]float64, len(a1.Values))
	r.IsAbsent = make([]bool, len(a1.Values))
	r.StartTime = proto.Int32(from)
	r.StopTime = proto.Int32(until)

	d1 := make([]float64, windowSize)
	d2 := make([]float64, windowSize)

	for i, v1 := range a1.Values {
		v2 := a2.Values[i]
		if a1.IsAbsent[i] || a2.IsAbsent[i] {
			// make sure missing values are ignored
			v1 = math.NaN()
			v2 = math.NaN()
		}
		w1.Push(v1)
		w2.Push(v2)

		if i >= windowSize {
			// need a copy here because KS is destructive
			copy(d1, w1.data)
			copy(d2, w2.data)
			r.Values[i] = onlinestats.KS(d1, d2)
		} else {
			r.Values[i] = 0
			r.IsAbsent[i] = true
		}
	}
	return []*metricData{&r}
}
