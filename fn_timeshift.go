package main

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// timeShift(seriesList, timeShift, resetEnd=True)
func timeShift(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	// FIXME(dgryski): support resetEnd=true

	offs, err := getIntervalArg(e, 1, -1)
	if err != nil {
		return nil
	}

	arg, err := getSeriesArg(e.args[0], from+offs, until+offs, values)
	if err != nil {
		return nil
	}

	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("timeShift(%s,'%d')", a.GetName(), offs))
		r.StartTime = proto.Int32(a.GetStartTime() - offs)
		r.StopTime = proto.Int32(a.GetStopTime() - offs)
		results = append(results, &r)
	}
	return results
}
