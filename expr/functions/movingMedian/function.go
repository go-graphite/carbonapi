package movingMedian

import (
	"fmt"
	"github.com/JaderDias/movingmedian"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
	"strconv"
)

type movingMedian struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &movingMedian{}
	functions := []string{"movingMedian"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// movingMedian(seriesList, windowSize)
func (f *movingMedian) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		start -= uint32(n)
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
		r.StartTime = from
		r.StopTime = until

		data := movingmedian.NewMovingMedian(windowSize)

		for i, v := range a.Values {
			data.Push(v)

			if ridx := i - offset; ridx >= 0 {
				r.Values[ridx] = math.NaN()
				if i >= (windowSize - 1) {
					r.Values[ridx] = data.Median()
				}
			}
		}
		result = append(result, &r)
	}
	return result, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *movingMedian) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"movingMedian": {
			Description: "Graphs the moving median of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nmedian of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMedian(Server.instance01.threads.busy,10)\n  &target=movingMedian(Server.instance*.threads.idle,'5min')",
			Function:    "movingMedian(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMedian",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: types.NewSuggestions(
						5,
						7,
						10,
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					),
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
		},
	}
}
