package threshold

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type threshold struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &threshold{}
	functions := []string{"threshold"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *threshold) Do(_ context.Context, _ interfaces.Evaluator, e parser.Expr, from, until int64, _ map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// XXX does not match graphite's signature
	// BUG(nnuss): the signature *does* match but there is an edge case because of named argument handling if you use it *just* wrong:
	//			   threshold(value, "gold", label="Aurum")
	//			   will result in:
	//			   value = value
	//			   label = "Aurum" (by named argument)
	//			   color = "" (by default as len(positionalArgs) == 2 and there is no named 'color' arg)

	value, err := e.GetFloatArg(0)
	if err != nil {
		return nil, err
	}

	defaultLabel := e.Arg(0).StringValue()

	name, err := e.GetStringNamedOrPosArgDefault("label", 1, defaultLabel)
	if err != nil {
		return nil, err
	}

	color, err := e.GetStringNamedOrPosArgDefault("color", 2, "")
	if err != nil {
		return nil, err
	}

	newValues := []float64{value, value}
	stepTime := until - from
	stopTime := from + stepTime*int64(len(newValues))
	p := &types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:              name,
			StartTime:         from,
			StopTime:          stopTime,
			StepTime:          stepTime,
			Values:            newValues,
			ConsolidationFunc: "average",
		},
		Tags:         map[string]string{"name": name},
		GraphOptions: types.GraphOptions{Color: color},
	}

	return []*types.MetricData{p}, nil
}

func (f *threshold) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"threshold": {
			Name: "threshold",
			Params: []types.FunctionParam{
				{
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
				{
					Name: "label",
					Type: types.String,
				},
				{
					Name: "color",
					Type: types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes a float F, followed by a label (in double quotes) and a color.\n(See ``bgcolor`` in the render\\_api_ for valid color names & formats.)\n\nDraws a horizontal line at value F across the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=threshold(123.456, \"omgwtfbbq\", \"red\")",
			Function:    "threshold(value, label=None, color=None)",
			Group:       "Graph",
		},
	}
}
