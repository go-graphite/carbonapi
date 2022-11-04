package sinFunction

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type sinFunction struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &sinFunction{}
	functions := []string{"sinFunction", "sin"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// sinFunction(name, amplitude, step)
func (f *sinFunction) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		return nil, err
	}

	var amplitude = 1.0
	var stepInt = 60
	if e.ArgsLen() >= 2 {
		amplitude, err = e.GetFloatArgDefault(1, 1.0)
		if err != nil {
			return nil, err
		}
	}
	if e.ArgsLen() == 3 {
		stepInt, err = e.GetIntArgDefault(2, 60)
		if err != nil {
			return nil, err
		}
	}
	step := int64(stepInt)

	newValues := make([]float64, (until-from-1+step)/step)
	value := from
	for i := 0; i < len(newValues); i++ {
		newValues[i] = math.Sin(float64(value)) * amplitude
		value += step
	}

	r := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      name,
			Values:    newValues,
			StepTime:  step,
			StartTime: from,
			StopTime:  until,
		},
		Tags: map[string]string{"name": name},
	}

	return []*types.MetricData{&r}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *sinFunction) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"sinFunction": {
			Description: "Just returns the sine of the current time. The optional amplitude parameter changes the amplitude of the wave.\n" +
				"Example:\n\n.. code-block:: none\n\n &target=sin(\"The.time.series\", 2)\n\n" +
				"This would create a series named “The.time.series” that contains sin(x)*2. Accepts optional second argument as ‘amplitude’ parameter (default amplitude is 1)\n Accepts optional third argument as ‘step’ parameter (default step is 60 sec)\n\n" +
				"Alias: sin",
			Function: "sinFunction(name, amplitude=1, step=60)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "scale",
			Params: []types.FunctionParam{
				{
					Name:     "name",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "amplitude",
					Required: false,
					Type:     types.Integer,
					Default:  types.NewSuggestion(1),
				},
				{
					Name:     "step",
					Required: false,
					Type:     types.Integer,
					Default:  types.NewSuggestion(60),
				},
			},
		},
	}
}
