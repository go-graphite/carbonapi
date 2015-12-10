package main

import (
	"fmt"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"

	"github.com/gogo/protobuf/proto"
)

func constantLine(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	value, err := getFloatArg(e, 0)

	if err != nil {
		return nil
	}
	p := metricData{
		FetchResponse: pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("%g", value)),
			StartTime: proto.Int32(from),
			StopTime:  proto.Int32(until),
			StepTime:  proto.Int32(until - from),
			Values:    []float64{value, value},
			IsAbsent:  []bool{false, false},
		},
	}

	return []*metricData{&p}
}
