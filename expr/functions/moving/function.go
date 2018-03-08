package moving

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
	"strconv"
)

type moving struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &moving{}
	functions := []string{"movingAverage", "movingMin", "movingMax", "movingSum"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// movingXyz(seriesList, windowSize)
func (f *moving) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *moving) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"movingAverage": {
			Description: "Graphs the moving average of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\naverage of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingAverage(Server.instance01.threads.busy,10)\n  &target=movingAverage(Server.instance*.threads.idle,'5min')",
			Function:    "movingAverage(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingAverage",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: []string{
						"5",
						"7",
						"10",
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					},
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
		},
		"movingMin": {
			Description: "Graphs the moving minimum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nminimum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMin(Server.instance01.requests,10)\n  &target=movingMin(Server.instance*.errors,'5min')",
			Function:    "movingMin(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMin",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: []string{
						"5",
						"7",
						"10",
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					},
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
		},
		"movingMax": {
			Description: "Graphs the moving maximum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nmaximum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingMax(Server.instance01.requests,10)\n  &target=movingMax(Server.instance*.errors,'5min')",
			Function:    "movingMax(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingMax",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: []string{
						"5",
						"7",
						"10",
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					},
					Type: types.IntOrInterval,
				},
				{
					Name: "xFilesFactor",
					Type: types.Float,
				},
			},
		},
		"movingSum": {
			Description: "Graphs the moving sum of a metric (or metrics) over a fixed number of\npast points, or a time interval.\n\nTakes one metric or a wildcard seriesList followed by a number N of datapoints\nor a quoted string with a length of time like '1hour' or '5min' (See ``from /\nuntil`` in the render\\_api_ for examples of time formats), and an xFilesFactor value to specify\nhow many points in the window must be non-null for the output to be considered valid. Graphs the\nsum of the preceeding datapoints for each point on the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=movingSum(Server.instance01.requests,10)\n  &target=movingSum(Server.instance*.errors,'5min')",
			Function:    "movingSum(seriesList, windowSize, xFilesFactor=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "movingSum",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Suggestions: []string{
						"5",
						"7",
						"10",
						"1min",
						"5min",
						"10min",
						"30min",
						"1hour",
					},
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
