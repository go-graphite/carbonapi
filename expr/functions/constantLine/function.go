package constantLine

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

func init() {
	f := &Function{}
	functions := []string{"constantLine"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
