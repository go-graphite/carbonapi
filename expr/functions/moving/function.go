package moving

import (
	"fmt"
	"github.com/JaderDias/movingmedian"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
	"strconv"
)

func init() {
	functions := []string{"movingAverage", "movingMin", "movingMax", "movingSum"}
	fObj := &MovingAverage{}
	for _, f := range functions {
		metadata.RegisterFunction(f, fObj)
	}
	metadata.RegisterFunction("movingMedian", &MovingMedian{})
}

type MovingAverage struct {
	interfaces.FunctionBase
}

// movingXyz(seriesList, windowSize)
func (f *MovingAverage) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var n int
	var err error

	var scaleByStep bool

	var argstr string

	switch e.Args()[1].Type() {
	case parser.EtConst:
		n, err = e.GetIntArg(1)
		argstr = strconv.Itoa(n)
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		argstr = fmt.Sprintf("%q", e.Args()[1].StringValue())
		n = int(n32)
		scaleByStep = true
	default:
		err = parser.ErrBadType
	}
	if err != nil {
		return nil, err
	}

	windowSize := n

	start := from
	if scaleByStep {
		start -= int32(n)
	}

	arg, err := helper.GetSeriesArg(e.Args()[0], start, until, values)
	if err != nil {
		return nil, err
	}

	var offset int

	if scaleByStep {
		windowSize /= int(arg[0].StepTime)
		offset = windowSize
	}

	var result []*types.MetricData

	for _, a := range arg {
		w := &types.Windowed{Data: make([]float64, windowSize)}

		r := *a
		r.Name = fmt.Sprintf("%s(%s,%s)", e.Target(), a.Name, argstr)
		r.Values = make([]float64, len(a.Values)-offset)
		r.IsAbsent = make([]bool, len(a.Values)-offset)
		r.StartTime = from
		r.StopTime = until

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				// make sure missing values are ignored
				v = math.NaN()
			}

			if ridx := i - offset; ridx >= 0 {
				switch e.Target() {
				case "movingAverage":
					r.Values[ridx] = w.Mean()
				case "movingSum":
					r.Values[ridx] = w.Sum()
					//TODO(cldellow): consider a linear time min/max-heap for these,
					// e.g. http://stackoverflow.com/questions/8905525/computing-a-moving-maximum/8905575#8905575
				case "movingMin":
					r.Values[ridx] = w.Min()
				case "movingMax":
					r.Values[ridx] = w.Max()
				}
				if i < windowSize || math.IsNaN(r.Values[ridx]) {
					r.Values[ridx] = 0
					r.IsAbsent[ridx] = true
				}
			}
			w.Push(v)
		}
		result = append(result, &r)
	}
	return result, nil
}

type MovingMedian struct {
	interfaces.FunctionBase
}

// movingMedian(seriesList, windowSize)
func (f *MovingMedian) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var n int
	var err error

	var scaleByStep bool

	var argstr string

	switch e.Args()[1].Type() {
	case parser.EtConst:
		n, err = e.GetIntArg(1)
		argstr = strconv.Itoa(n)
	case parser.EtString:
		var n32 int32
		n32, err = e.GetIntervalArg(1, 1)
		n = int(n32)
		argstr = fmt.Sprintf("%q", e.Args()[1].StringValue())
		scaleByStep = true
	default:
		err = parser.ErrBadType
	}
	if err != nil {
		return nil, err
	}

	windowSize := n

	start := from
	if scaleByStep {
		start -= int32(n)
	}

	arg, err := helper.GetSeriesArg(e.Args()[0], start, until, values)
	if err != nil {
		return nil, err
	}

	var offset int

	if scaleByStep {
		windowSize /= int(arg[0].StepTime)
		offset = windowSize
	}

	var result []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("movingMedian(%s,%s)", a.Name, argstr)
		r.Values = make([]float64, len(a.Values)-offset)
		r.IsAbsent = make([]bool, len(a.Values)-offset)
		r.StartTime = from
		r.StopTime = until

		data := movingmedian.NewMovingMedian(windowSize)

		for i, v := range a.Values {
			if a.IsAbsent[i] {
				data.Push(math.NaN())
			} else {
				data.Push(v)
			}
			if ridx := i - offset; ridx >= 0 {
				r.Values[ridx] = math.NaN()
				if i >= (windowSize - 1) {
					r.Values[ridx] = data.Median()
				}
				if math.IsNaN(r.Values[ridx]) {
					r.IsAbsent[ridx] = true
				}
			}
		}
		result = append(result, &r)
	}
	return result, nil
}
