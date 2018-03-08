package constantLine

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type constantLine struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &constantLine{}
	functions := []string{"constantLine"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *constantLine) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	value, err := e.GetFloatArg(0)

	if err != nil {
		return nil, err
	}
	p := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("%g", value),
			StartTime: from,
			StopTime:  until,
			StepTime:  until - from,
			Values:    []float64{value, value},
			IsAbsent:  []bool{false, false},
		},
	}

	return []*types.MetricData{&p}, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *constantLine) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"constantLine": {
			Description: "Takes a float F.\n\nDraws a horizontal line at value F across the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=constantLine(123.456)",
			Function:    "constantLine(value)",
			Group:       "Special",
			Module:      "graphite.render.functions",
			Name:        "constantLine",
			Params: []types.FunctionParam{
				{
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}
