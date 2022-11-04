package randomWalk

import (
	"context"
	"math/rand"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type randomWalk struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &randomWalk{}
	functions := []string{"randomWalk", "randomWalkFunction"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// randomWalk(name, step=60)
func (f *randomWalk) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		name = "randomWalk"
	}
	stepInt, err := e.GetIntNamedOrPosArgDefault("step", 1, 60)
	if err != nil {
		return nil, err
	}
	step := int64(stepInt)

	size := (until - from) / step
	until = from + step*size // Re-compute 'until' in case 'size' is a not a divisor of the range

	r := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:              name,
			Values:            make([]float64, size),
			StepTime:          step,
			StartTime:         from,
			StopTime:          until,
			ConsolidationFunc: "average",
		},
		Tags: map[string]string{"name": name},
	}

	for i := 1; i < len(r.Values)-1; i++ {
		r.Values[i+1] = r.Values[i] + (rand.Float64() - 0.5)
	}
	return []*types.MetricData{&r}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *randomWalk) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"randomWalk": {
			Description: "Short Alias: randomWalk()\n\nReturns a random walk starting at 0. This is great for testing when there is\nno real data in whisper.\n\nExample:\n\n.. code-block:: none\n\n  &target=randomWalk(\"The.time.series\")\n\nThis would create a series named \"The.time.series\" that contains points where\nx(t) == x(t-1)+random()-0.5, and x(0) == 0.\nAccepts optional second argument as 'step' parameter (default step is 60 sec)",
			Function:    "randomWalk(name, step=60)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "randomWalk",
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
		"randomWalkFunction": {
			Description: "Short Alias: randomWalk()\n\nReturns a random walk starting at 0. This is great for testing when there is\nno real data in whisper.\n\nExample:\n\n.. code-block:: none\n\n  &target=randomWalk(\"The.time.series\")\n\nThis would create a series named \"The.time.series\" that contains points where\nx(t) == x(t-1)+random()-0.5, and x(0) == 0.\nAccepts optional second argument as 'step' parameter (default step is 60 sec)",
			Function:    "randomWalkFunction(name, step=60)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "randomWalkFunction",
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
