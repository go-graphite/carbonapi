package identity

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type identity struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &identity{}
	functions := []string{"identity"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// identity(name)
func (f *identity) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		return nil, err
	}

	step := int64(60)

	newValues := make([]float64, (until-from-1+step)/step)
	value := from
	for i := 0; i < len(newValues); i++ {
		newValues[i] = float64(value)
		value += step
	}

	p := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      "identity(" + name + ")",
			StartTime: from,
			StopTime:  until,
			StepTime:  step,
			Values:    newValues,
		},
		Tags: map[string]string{"name": name},
	}

	return []*types.MetricData{&p}, nil

}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *identity) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"identity": {
			Description: "Identity function: Returns datapoints where the value equals the timestamp of the datapoint.\n Useful when you have another series where the value is a timestamp, and you want to compare it to the time of the datapoint, to render an age\n\nExample:\n\n.. code-block:: none\n\n  &target=identity(\"The.time.series\")\n This would create a series named “The.time.series” that contains points where x(t) == t.)",
			Function:    "identity(name)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "identity",
			Params: []types.FunctionParam{
				{
					Name:     "name",
					Required: true,
					Type:     types.String,
				},
			},
		},
	}
}
