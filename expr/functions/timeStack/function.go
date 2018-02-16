package timeStack

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &timeStack{}
	functions := []string{"timeStack"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type timeStack struct {
	interfaces.FunctionBase
}

// timeStack(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd)
func (f *timeStack) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	unit, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}

	start, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	end, err := e.GetIntArg(3)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for i := int32(start); i < int32(end); i++ {
		offs := i * unit
		arg, err := helper.GetSeriesArg(e.Args()[0], from+offs, until+offs, values)
		if err != nil {
			return nil, err
		}

		for _, a := range arg {
			r := *a
			r.Name = fmt.Sprintf("timeShift(%s,%d)", a.Name, offs)
			r.StartTime = a.StartTime - offs
			r.StopTime = a.StopTime - offs
			results = append(results, &r)
		}
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeStack) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"timeStack": {
			Description: "Takes one metric or a wildcard seriesList, followed by a quoted string with the\nlength of time (See ``from / until`` in the render\\_api_ for examples of time formats).\nAlso takes a start multiplier and end multiplier for the length of time\n\ncreate a seriesList which is composed the original metric series stacked with time shifts\nstarting time shifts from the start multiplier through the end multiplier\n\nUseful for looking at history, or feeding into averageSeries or stddevSeries.\n\nExample:\n\n.. code-block:: none\n\n  &target=timeStack(Sales.widgets.largeBlue,\"1d\",0,7)    # create a series for today and each of the previous 7 days",
			Function:    "timeStack(seriesList, timeShiftUnit='1d', timeShiftStart=0, timeShiftEnd=7)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "timeStack",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: "1d",
					Name:    "timeShiftUnit",
					Suggestions: []string{
						"1h",
						"6h",
						"12h",
						"1d",
						"2d",
						"7d",
						"14d",
						"30d",
					},
					Type: types.Interval,
				},
				{
					Default: "0",
					Name:    "timeShiftStart",
					Type:    types.Integer,
				},
				{
					Default: "7",
					Name:    "timeShiftEnd",
					Type:    types.Integer,
				},
			},
		},
	}
}
