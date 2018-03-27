package timeFunction

import (
	"errors"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type timeFunction struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &timeFunction{}
	functions := []string{"timeFunction", "time"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *timeFunction) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		return nil, err
	}

	stepInt, err := e.GetIntArgDefault(1, 60)
	if err != nil {
		return nil, err
	}
	if stepInt <= 0 {
		return nil, errors.New("step can't be less than 0")
	}
	step := int64(stepInt)

	// emulate the behavior of this Python code:
	//   while when < requestContext["endTime"]:
	//     newValues.append(time.mktime(when.timetuple()))
	//     when += delta

	newValues := make([]float64, (until-from-1+step)/step)
	value := from
	for i := 0; i < len(newValues); i++ {
		newValues[i] = float64(value)
		value += step
	}

	p := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:              name,
			StartTime:         from,
			StopTime:          until,
			StepTime:          step,
			Values:            newValues,
			ConsolidationFunc: "max",
		},
	}

	return []*types.MetricData{&p}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *timeFunction) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"timeFunction": {
			Description: "Short Alias: time()\n\nJust returns the timestamp for each X value. T\n\nExample:\n\n.. code-block:: none\n\n  &target=time(\"The.time.series\")\n\nThis would create a series named \"The.time.series\" that contains in Y the same\nvalue (in seconds) as X.\nAccepts optional second argument as 'step' parameter (default step is 60 sec)",
			Function:    "timeFunction(name, step=60)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "timeFunction",
			Params: []types.FunctionParam{
				{
					Name:     "name",
					Required: true,
					Type:     types.String,
				},
				{
					Default: types.NewSuggestion(60),
					Name:    "step",
					Type:    types.Integer,
				},
			},
		},
		"time": {
			Description: "Short Alias: time()\n\nJust returns the timestamp for each X value. T\n\nExample:\n\n.. code-block:: none\n\n  &target=time(\"The.time.series\")\n\nThis would create a series named \"The.time.series\" that contains in Y the same\nvalue (in seconds) as X.\nAccepts optional second argument as 'step' parameter (default step is 60 sec)",
			Function:    "time(name, step=60)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "time",
			Params: []types.FunctionParam{
				{
					Name:     "name",
					Required: true,
					Type:     types.String,
				},
				{
					Default: types.NewSuggestion(60),
					Name:    "step",
					Type:    types.Integer,
				},
			},
		},
	}
}
