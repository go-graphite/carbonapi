package main

import (
	"fmt"
	"math"

	"github.com/JaderDias/movingmedian"
	"github.com/gogo/protobuf/proto"
)

// movingAverage(seriesList, windowSize)
func movingAverage(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	var n int
	var err error

	var scaleByStep bool

	switch e.args[1].etype {
	case etConst:
		n, err = getIntArg(e, 1)
	case etString:
		var n32 int32
		n32, err = getIntervalArg(e, 1, 1)
		n = int(n32)
		scaleByStep = true
	default:
		err = ErrBadType
	}
	if err != nil {
		return nil
	}

	windowSize := n

	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	if scaleByStep {
		windowSize /= int(arg[0].GetStepTime())
	}

	var result []*metricData

	for _, a := range arg {
		w := &Windowed{data: make([]float64, windowSize)}

		r := *a
		r.Name = proto.String(fmt.Sprintf("movingAverage(%s,%d)", a.GetName(), windowSize))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		r.StartTime = proto.Int32(from)
		r.StopTime = proto.Int32(until)

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				// make sure missing values are ignored
				v = math.NaN()
			}
			r.Values[i] = w.Mean()
			w.Push(v)
			if i < windowSize || math.IsNaN(r.Values[i]) {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		result = append(result, &r)
	}
	return result
}

// movingMedian(seriesList, windowSize)
func movingMedian(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	var n int
	var err error

	var scaleByStep bool

	switch e.args[1].etype {
	case etConst:
		n, err = getIntArg(e, 1)
	case etString:
		var n32 int32
		n32, err = getIntervalArg(e, 1, 1)
		n = int(n32)
		scaleByStep = true
	default:
		err = ErrBadType
	}
	if err != nil {
		return nil
	}

	windowSize := n

	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	if scaleByStep {
		windowSize /= int(arg[0].GetStepTime())
	}

	var result []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("movingMedian(%s,%d)", a.GetName(), windowSize))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		r.StartTime = proto.Int32(from)
		r.StopTime = proto.Int32(until)

		data := movingmedian.NewMovingMedian(windowSize)

		for i, v := range a.Values {
			r.Values[i] = math.NaN()
			if a.IsAbsent[i] {
				data.Push(math.NaN())
			} else {
				data.Push(v)
			}
			if i >= (windowSize - 1) {
				r.Values[i] = data.Median()
			}
			if math.IsNaN(r.Values[i]) {
				r.IsAbsent[i] = true
			}
		}
		result = append(result, &r)
	}
	return result
}
