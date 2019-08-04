package aggregateLine

import (
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type aggregateLine struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	f := &aggregateLine{}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aggregateLine"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// aggregateLine(*seriesLists)
func (f *aggregateLine) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	callback := "avg"
	keepStep := false
	switch len(e.Args()) {
	case 2:
		callback, err = e.GetStringArg(1)
		if err != nil {
			return nil, err
		}
	case 3:
		callback, err = e.GetStringArg(1)
		if err != nil {
			return nil, err
		}

		keepStep, err = e.GetBoolArgDefault(2, false)
		if err != nil {
			return nil, err
		}
	}

	aggFunc, ok := consolidations.ConsolidationToFunc[callback]
	if !ok {
		return nil, fmt.Errorf("unsupported consolidation function %s", callback)
	}

	var results []*types.MetricData
	for _, a := range args {
		val := aggFunc(a.Values)
		var name string
		if !math.IsNaN(val) {
			name = fmt.Sprintf("aggregateLine(%s, %g)", a.Name, val)
		} else {
			name = fmt.Sprintf("aggregateLine(%s, None)", a.Name)
		}

		r := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				StartTime:         a.FetchResponse.StartTime,
				StopTime:          a.FetchResponse.StopTime,
				PathExpression:    a.FetchResponse.PathExpression,
				ConsolidationFunc: a.FetchResponse.ConsolidationFunc,
				RequestStartTime:  a.FetchResponse.RequestStartTime,
				RequestStopTime:   a.FetchResponse.RequestStopTime,
				XFilesFactor:      a.FetchResponse.XFilesFactor,
			},
		}
		if keepStep {
			r.FetchResponse.Values = make([]float64, len(a.Values))
			for i := range r.Values {
				r.Values[i] = val
			}
			r.FetchResponse.StepTime = a.FetchResponse.StepTime
		} else {
			r.FetchResponse.StepTime = a.FetchResponse.StopTime - a.FetchResponse.StartTime
			r.FetchResponse.Values = []float64{val, val}
		}

		results = append(results, &r)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aggregateLine) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aggregateLine": {
			Name:        "aggregateLine",
			Function:    "aggregateLine(seriesList, func='average', keepStep=False)",
			Description: "Takes a metric or wildcard seriesList and draws a horizontal line\nbased on the function applied to each series.\n\nIf the optional keepStep parameter is set to True, the result will\nhave the same time period and step as the source series.\n\nNote: By default, the graphite renderer consolidates data points by\naveraging data points over time. If you are using the 'min' or 'max'\nfunction for aggregateLine, this can cause an unusual gap in the\nline drawn by this function and the data itself. To fix this, you\nshould use the consolidateBy() function with the same function\nargument you are using for aggregateLine. This will ensure that the\nproper data points are retained and the graph should line up\ncorrectly.\n\nExample:\n\n.. code-block:: none\n\n  &target=aggregateLine(server01.connections.total, 'avg')\n  &target=aggregateLine(server*.connections.total, 'avg')",
			Module:      "graphite.render.functions",
			Group:       "Calculate",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Type:     types.SeriesList,
					Required: true,
				},
				{
					Name:     "func",
					Type:     types.AggFunc,
					Required: false,
					Options:  consolidations.AvailableConsolidationFuncs(),
					Default: &types.Suggestion{
						Value: "average",
						Type:  types.SString,
					},
				},
				{
					Name: "keepStep",
					Type: types.Boolean,
					Default: &types.Suggestion{
						Value: false,
						Type:  types.SBool,
					},
				},
			},
		},
	}
}
