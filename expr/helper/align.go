package helper

import (
	"math"
	"time"

	"github.com/go-graphite/carbonapi/expr/types"
)

// GCD returns greatest common divisor calculated via Euclidean algorithm
func GCD(a, b int64) int64 {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

// LCM returns the least common multiple of 2 or more integers via GDB
func LCM(args ...int64) int64 {
	if len(args) <= 1 {
		if len(args) == 0 {
			return 0
		}
		return args[0]
	}
	lcm := args[0] / GCD(args[0], args[1]) * args[1]

	for i := 2; i < len(args); i++ {
		lcm = LCM(lcm, args[i])
	}
	return lcm
}

// GetCommonStep returns LCM(steps) for slice of metrics.
// minStart and maxStop will be set to closest lower or equal multiple of LCM(steps).
func GetCommonStep(args []*types.MetricData) int64 {
	steps := make([]int64, 0, len(args))
	for _, arg := range args {
		steps = append(steps, arg.StepTime)
	}
	commonStep := LCM(steps...)
	return commonStep
}

// ScaleToCommonStep returns the metrics, aligned LCM of all metrics steps.
// It respects xFilesFactor and fills gaps in the begin and end with NaNs if needed.
func ScaleToCommonStep(args []*types.MetricData) []*types.MetricData {
	commonStep := GetCommonStep(args)
	for _, arg := range args {
		if arg.StepTime == commonStep {
			continue
		}
		stepFactor := commonStep / arg.StepTime
		newStart := arg.StartTime - (arg.StartTime % commonStep)
		if (arg.StartTime % commonStep) != 0 {
			// Fill with NaNs from newStart to arg.StartTime
			valCnt := (arg.StartTime - newStart) / arg.StepTime
			nans := genNaNs(int(valCnt))
			arg.Values = append(nans, arg.Values...)
			arg.StartTime = newStart
		}

		newValsLen := 1 + int64(len(arg.Values)-1)/stepFactor
		newStop := arg.StartTime + newValsLen*commonStep
		newVals := make([]float64, 0, newValsLen)

		if len(arg.Values) != int(stepFactor*newValsLen) {
			// Fill the last step with NaNs from newStart to (newStart + commonStep - arg.StepTime)
			valCnt := int(stepFactor*newValsLen) - len(arg.Values)
			nans := genNaNs(valCnt)
			arg.Values = append(arg.Values, nans...)
		}
		arg.StopTime = newStop
		for i := 0; i < len(arg.Values); i += int(stepFactor) {
			aggregatedBatch := aggregateBatch(arg.Values[i:i+int(stepFactor)], arg)
			newVals = append(newVals, aggregatedBatch)
		}
		arg.StepTime = commonStep
		arg.Values = newVals
	}
	return args
}

func aggregateBatch(vals []float64, arg *types.MetricData) float64 {
	if arg.XFilesFactor != 0 {
		notNans := 0
		for _, i := range vals {
			if !math.IsNaN(i) {
				notNans++
			}
		}
		if float32(notNans)/float32(len(vals)) < arg.XFilesFactor {
			return math.NaN()
		}
	}
	return arg.GetAggregateFunction()(vals)
}

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
				newVals := genNaNs(int(valCnt))
				arg.Values = append(newVals, arg.Values...)
				arg.StartTime = minStart
			}

			if arg.StopTime > maxStop {
				maxStop = arg.StopTime
			}
			if maxStop > arg.StopTime {
				valCnt := (maxStop - arg.StopTime) / arg.StepTime
				newVals := genNaNs(int(valCnt))
				arg.Values = append(arg.Values, newVals...)
				arg.StopTime = maxStop
			}
		}
	}
	return args
}

func genNaNs(length int) []float64 {
	nans := make([]float64, length)
	for i := range nans {
		nans[i] = math.NaN()
	}
	return nans
}
