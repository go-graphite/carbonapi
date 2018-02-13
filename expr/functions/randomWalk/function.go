package randomWalk

import (
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math/rand"
)

func init() {
	f := &Function{}
	functions := []string{"randomWalk", "randomWalkFunction"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// squareRoot(seriesList)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		name = "randomWalk"
	}

	size := until - from

	r := types.MetricData{FetchResponse: pb.FetchResponse{
		Name:      name,
		Values:    make([]float64, size),
		IsAbsent:  make([]bool, size),
		StepTime:  1,
		StartTime: from,
		StopTime:  until,
	}}

	for i := 1; i < len(r.Values)-1; i++ {
		r.Values[i+1] = r.Values[i] + (rand.Float64() - 0.5)
	}
	return []*types.MetricData{&r}, nil
}
