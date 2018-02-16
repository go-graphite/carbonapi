package test

import (
	"github.com/go-graphite/carbonapi/expr/types"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
)

func MakeResponse(name string, values []float64, step, start int32) *types.MetricData {

	absent := make([]bool, len(values))

	for i, v := range values {
		if math.IsNaN(v) {
			values[i] = 0
			absent[i] = true
		}
	}

	stop := start + int32(len(values))*step

	return &types.MetricData{FetchResponse: pb.FetchResponse{
		Name:      name,
		Values:    values,
		StartTime: start,
		StepTime:  step,
		StopTime:  stop,
		IsAbsent:  absent,
	}}
}
