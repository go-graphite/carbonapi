package timeSlice

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type timeSlice struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &timeSlice{}
	functions := []string{"timeSlice"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *timeSlice) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	start32, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}
	start := int64(start32)

	end, err := e.GetIntervalNamedOrPosArgDefault("endSliceAt", 2, -1, 0)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	start += now
	end += now

	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("timeSlice(%s, %d, %d)", a.Name, start, end)
		r.Values = make([]float64, len(a.Values))

		current := from
		for i, v := range a.Values {
			if current < start || current > end {
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = v
			}
			current += a.StepTime
		}
		results = append(results, &r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeSlice) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"timeSlice": {
			Name:        "timeSlice",
			Function:    "timeSlice(seriesList, startSliceAt, endSliceAt='now')",
			Description: "Takes one metric or a wildcard metric, followed by a quoted string with the\ntime to start the line and another quoted string with the time to end the line.\nThe start and end times are inclusive. See ``from / until`` in the render\\_api_\nfor examples of time formats.\n\nUseful for filtering out a part of a series of data from a wider range of\ndata.\n\nExample:\n\n.. code-block:: none\n\n  &target=timeSlice(network.core.port1,\"00:00 20140101\",\"11:59 20140630\")\n  &target=timeSlice(network.core.port1,\"12:00 20140630\",\"now\")",
			Module:      "graphite.render.functions",
			Group:       "Transform",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "startSliceAt",
					Type:     types.Date,
					Required: true,
				},
				{
					Name:    "endSliceAt",
					Type:    types.Interval,
					Default: types.NewSuggestion("now"),
				},
			},
		},
	}
}
