package helper

import (
	"math"
	"time"

	"github.com/go-graphite/carbonapi/expr/types"
)

// GetBuckets returns amount buckets for timeSeries (defined with startTime, stopTime and step (bucket) size.
func GetBuckets(start, stop, bucketSize int64) int64 {
	return int64(math.Ceil(float64(stop-start) / float64(bucketSize)))
}

// AlignStartToInterval aligns start of serie to interval
func AlignStartToInterval(start, stop, bucketSize int64) int64 {
	for _, v := range []int64{86400, 3600, 60} {
		if bucketSize >= v {
			start -= start % v
			break
		}
	}

	return start
}

// AlignToBucketSize aligns start and stop of serie to specified bucket (step) size
func AlignToBucketSize(start, stop, bucketSize int64) (int64, int64) {
	start = time.Unix(start, 0).Truncate(time.Duration(bucketSize) * time.Second).Unix()
	newStop := time.Unix(stop, 0).Truncate(time.Duration(bucketSize) * time.Second).Unix()

	// check if a partial bucket is needed
	if stop != newStop {
		newStop += bucketSize
	}

	return start, newStop
}

// AlignSeries aligns different series together. By default it only prepends and appends NaNs in case of different length, but if ExtrapolatePoints is enabled, it can extrapolate
func AlignSeries(args []*types.MetricData) []*types.MetricData {
	minStart := args[0].StartTime
	maxStop := args[0].StopTime
	maxVals := 0
	minStepTime := args[0].StepTime
	for j := 0; j < 2; j++ {
		if ExtrapolatePoints {
			for _, arg := range args {
				if arg.StepTime < minStepTime {
					minStepTime = arg.StepTime
				}

				if arg.StepTime > minStepTime {
					valsCnt := int(math.Ceil(float64(arg.StopTime-arg.StartTime) / float64(minStepTime)))
					newVals := make([]float64, valsCnt)
					ts := arg.StartTime
					nextTs := arg.StartTime + arg.StepTime
					i := 0
					j := 0
					pointsPerInterval := float64(ts-nextTs) / float64(minStepTime)
					v := arg.Values[0]
					dv := (arg.Values[0] - arg.Values[1]) / pointsPerInterval
					for ts < arg.StopTime {
						newVals[i] = v
						v += dv
						if ts > nextTs {
							j++
							nextTs += arg.StepTime
							v = arg.Values[j]
							dv = (arg.Values[j-1] - v) / pointsPerInterval
						}
						ts += minStepTime
						i++
					}
					arg.Values = newVals
					arg.StepTime = minStepTime
				}
			}
		}

		for _, arg := range args {
			if len(arg.Values) > maxVals {
				maxVals = len(arg.Values)
			}
			if arg.StartTime < minStart {
				minStart = arg.StartTime
			}
			if minStart < arg.StartTime {
				valCnt := (arg.StartTime - minStart) / arg.StepTime
				newVals := make([]float64, valCnt)
				newVals = append(newVals, arg.Values...)
				arg.Values = newVals
				arg.StartTime = minStart
			}

			if arg.StopTime > maxStop {
				maxStop = arg.StopTime
			}
			if maxStop > arg.StopTime {
				valCnt := (maxStop - arg.StopTime) / arg.StepTime
				newVals := make([]float64, valCnt)
				arg.Values = append(arg.Values, newVals...)
				arg.StopTime = maxStop
			}
		}
	}
	return args
}
