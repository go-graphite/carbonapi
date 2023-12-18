package timeStack

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type timeStack struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &timeStack{}
	functions := []string{"timeStack"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// timeStack(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd)
func (f *timeStack) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	unit, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}
	unitStr := e.Arg(1).StringValue()

	start, err := e.GetIntArgDefault(2, 0)
	if err != nil {
		return nil, err
	}

	end, err := e.GetIntArgDefault(3, 7)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for i := int64(start); i < int64(end); i++ {
		offs := i * int64(unit)
		fromNew := from + offs
		untilNew := until + offs
		arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), fromNew, untilNew, values)
		if err != nil {
			return nil, err
		}

		offsStr := strconv.FormatInt(offs, 10)
		for _, a := range arg {
			r := a.CopyLink()
			r.Name = fmt.Sprintf("timeShift(%s,%s,%d)", a.Name, unitStr, offs)
			r.StartTime = a.StartTime - offs
			r.StopTime = a.StopTime - offs
			r.Tags["timeShiftUnit"] = unitStr
			r.Tags["timeShift"] = offsStr
			results = append(results, r)
		}
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeStack) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
					Default: types.NewSuggestion("1d"),
					Name:    "timeShiftUnit",
					Suggestions: types.NewSuggestions(
						"1h",
						"6h",
						"12h",
						"1d",
						"2d",
						"7d",
						"14d",
						"30d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "timeShiftStart",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion(7),
					Name:    "timeShiftEnd",
					Type:    types.Integer,
				},
			},
		},
	}
}
