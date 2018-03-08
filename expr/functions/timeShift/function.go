package timeShift

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type timeShift struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &timeShift{}
	functions := []string{"timeShift"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// timeShift(seriesList, timeShift, resetEnd=True)
func (f *timeShift) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// FIXME(dgryski): support resetEnd=true
	// FIXME(civil): support alignDst
	offs, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}

	arg, err := helper.GetSeriesArg(e.Args()[0], from+offs, until+offs, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("timeShift(%s,'%d')", a.Name, offs)
		r.StartTime = a.StartTime - offs
		r.StopTime = a.StopTime - offs
		results = append(results, &r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeShift) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"timeShift": {
			Description: "Takes one metric or a wildcard seriesList, followed by a quoted string with the\nlength of time (See ``from / until`` in the render\\_api_ for examples of time formats).\n\nDraws the selected metrics shifted in time. If no sign is given, a minus sign ( - ) is\nimplied which will shift the metric back in time. If a plus sign ( + ) is given, the\nmetric will be shifted forward in time.\n\nWill reset the end date range automatically to the end of the base stat unless\nresetEnd is False. Example case is when you timeshift to last week and have the graph\ndate range set to include a time in the future, will limit this timeshift to pretend\nending at the current time. If resetEnd is False, will instead draw full range including\nfuture time.\n\nBecause time is shifted by a fixed number of seconds, comparing a time period with DST to\na time period without DST, and vice-versa, will result in an apparent misalignment. For\nexample, 8am might be overlaid with 7am. To compensate for this, use the alignDST option.\n\nUseful for comparing a metric against itself at a past periods or correcting data\nstored at an offset.\n\nExample:\n\n.. code-block:: none\n\n  &target=timeShift(Sales.widgets.largeBlue,\"7d\")\n  &target=timeShift(Sales.widgets.largeBlue,\"-7d\")\n  &target=timeShift(Sales.widgets.largeBlue,\"+1h\")",
			Function:    "timeShift(seriesList, timeShift, resetEnd=True, alignDST=False)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "timeShift",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "timeShift",
					Required: true,
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
					Default: "true",
					Name:    "resetEnd",
					Type:    types.Boolean,
				},
				/*
					{
						Default: "false",
						Name:    "alignDst",
						Type:    types.Boolean,
					},
				*/
			},
		},
	}
}
