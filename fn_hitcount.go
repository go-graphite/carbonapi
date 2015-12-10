package main

import (
	"fmt"
	"math"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	"github.com/gogo/protobuf/proto"
)

// hitcount(seriesList, intervalString, alignToInterval=False)
func hitcount(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	bucketSize, err := getIntervalArg(e, 1, 1)
	if err != nil {
		return nil
	}

	alignToInterval, err := getBoolArgDefault(e, 2, false)
	if err != nil {
		return nil
	}

	start := args[0].GetStartTime()
	stop := args[0].GetStopTime()
	if alignToInterval {
		start = alignStartToInterval(start, stop, bucketSize)
	}

	buckets := getBuckets(start, stop, bucketSize)
	results := make([]*metricData, 0, len(args))
	for _, arg := range args {

		var name string
		switch len(e.args) {
		case 2:
			name = fmt.Sprintf("hitcount(%s,'%s')", arg.GetName(), e.args[1].valStr)
		case 3:
			name = fmt.Sprintf("hitcount(%s,'%s',%s)", arg.GetName(), e.args[1].valStr, e.args[2].target)
		}

		r := metricData{FetchResponse: pb.FetchResponse{
			Name:      proto.String(name),
			Values:    make([]float64, buckets, buckets+1),
			IsAbsent:  make([]bool, buckets, buckets+1),
			StepTime:  proto.Int32(bucketSize),
			StartTime: proto.Int32(start),
			StopTime:  proto.Int32(stop),
		}}

		bucketEnd := start + bucketSize
		t := arg.GetStartTime()
		ridx := 0
		var count float64
		bucketItems := 0
		for i, v := range arg.Values {
			bucketItems++
			if !arg.IsAbsent[i] {
				if math.IsNaN(count) {
					count = 0
				}

				count += v * float64(arg.GetStepTime())
			}

			t += arg.GetStepTime()

			if t >= stop {
				break
			}

			if t >= bucketEnd {
				if math.IsNaN(count) {
					r.Values[ridx] = 0
					r.IsAbsent[ridx] = true
				} else {
					r.Values[ridx] = count
				}

				ridx++
				bucketEnd += bucketSize
				count = math.NaN()
				bucketItems = 0
			}
		}

		// remaining values
		if bucketItems > 0 {
			if math.IsNaN(count) {
				r.Values[ridx] = 0
				r.IsAbsent[ridx] = true
			} else {
				r.Values[ridx] = count
			}
		}

		results = append(results, &r)
	}
	return results
}
