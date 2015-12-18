package main

import (
	"fmt"
	"math"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	"github.com/gogo/protobuf/proto"
)

// summarize(seriesList, intervalString, func='sum', alignToFrom=False)
func summarize(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}

	bucketSize, err := getIntervalArg(e, 1, 1)
	if err != nil {
		return nil
	}

	summarizeFunction, err := getStringArgDefault(e, 2, "sum")
	if err != nil {
		return nil
	}

	alignToFrom, err := getBoolArgDefault(e, 3, false)
	if err != nil {
		return nil
	}

	start := args[0].GetStartTime()
	stop := args[0].GetStopTime()
	if !alignToFrom {
		start, stop = alignToBucketSize(start, stop, bucketSize)
	}

	buckets := getBuckets(start, stop, bucketSize)
	results := make([]*metricData, 0, len(args))
	for _, arg := range args {

		var name string
		switch len(e.args) {
		case 2:
			name = fmt.Sprintf("summarize(%s,'%s')", arg.GetName(), e.args[1].valStr)
		case 3:
			name = fmt.Sprintf("summarize(%s,'%s','%s')", arg.GetName(), e.args[1].valStr, e.args[2].valStr)
		case 4:
			name = fmt.Sprintf("summarize(%s,'%s','%s',%s)", arg.GetName(), e.args[1].valStr, e.args[2].valStr, e.args[3].target)
		}

		r := metricData{FetchResponse: pb.FetchResponse{
			Name:      proto.String(name),
			Values:    make([]float64, buckets, buckets),
			IsAbsent:  make([]bool, buckets, buckets),
			StepTime:  proto.Int32(bucketSize),
			StartTime: proto.Int32(start),
			StopTime:  proto.Int32(stop),
		}}

		t := arg.GetStartTime() // unadjusted
		bucketEnd := start + bucketSize
		values := make([]float64, 0, bucketSize/arg.GetStepTime())
		ridx := 0
		bucketItems := 0
		for i, v := range arg.Values {
			bucketItems++
			if !arg.IsAbsent[i] {
				values = append(values, v)
			}

			t += arg.GetStepTime()

			if t >= stop {
				break
			}

			if t >= bucketEnd {
				rv := summarizeValues(summarizeFunction, values)

				if math.IsNaN(rv) {
					r.IsAbsent[ridx] = true
				}

				r.Values[ridx] = rv
				ridx++
				bucketEnd += bucketSize
				bucketItems = 0
				values = values[:0]
			}
		}

		// last partial bucket
		if bucketItems > 0 {
			rv := summarizeValues(summarizeFunction, values)
			if math.IsNaN(rv) {
				r.Values[ridx] = 0
				r.IsAbsent[ridx] = true
			} else {
				r.Values[ridx] = rv
				r.IsAbsent[ridx] = false
			}
		}

		results = append(results, &r)
	}
	return results
}
