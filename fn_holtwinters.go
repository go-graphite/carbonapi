package main

import (
	"fmt"

	"github.com/datastream/holtwinters"
	pb "github.com/dgryski/carbonzipper/carbonzipperpb"

	"github.com/gogo/protobuf/proto"
)

func holtWintersForecast(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	var results []*metricData
	args, err := getSeriesArgs(e.args, from-7*86400, until, values)
	if err != nil {
		return nil
	}

	const alpha = 0.1
	const beta = 0.0035
	const gamma = 0.1

	for _, arg := range args {
		stepTime := arg.GetStepTime()
		numStepsToWalkToGetOriginalData := (int)((until - from) / stepTime)

		//originalSeries := arg.Values[len(arg.Values)-numStepsToWalkToGetOriginalData:]
		bootStrapSeries := arg.Values[:len(arg.Values)-numStepsToWalkToGetOriginalData]

		//In line with graphite, we define a season as a single day.
		//A period is the number of steps that make a season.
		period := (int)((24 * 60 * 60) / stepTime)

		predictions, err := holtwinters.Forecast(bootStrapSeries, alpha, beta, gamma, period, numStepsToWalkToGetOriginalData)
		if err != nil {
			return nil
		}

		predictionsOfInterest := predictions[len(predictions)-numStepsToWalkToGetOriginalData:]

		r := metricData{FetchResponse: pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("holtWintersForecast(%s)", arg.GetName())),
			Values:    make([]float64, len(predictionsOfInterest)),
			IsAbsent:  make([]bool, len(predictionsOfInterest)),
			StepTime:  proto.Int32(arg.GetStepTime()),
			StartTime: proto.Int32(arg.GetStartTime() + 7*86400),
			StopTime:  proto.Int32(arg.GetStopTime()),
		}}
		r.Values = predictionsOfInterest

		results = append(results, &r)
	}
	return results
}
