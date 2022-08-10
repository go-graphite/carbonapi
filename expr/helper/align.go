package helper

import (
	"math"
	"time"

	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
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

// GetCommonStep returns LCM(steps), changed (bool) for slice of metrics.
// If all metrics have the same step, changed == false.
func GetCommonStep(args []*types.MetricData) (commonStep int64, changed bool) {
	steps := make([]int64, 0, 1)
	stepsIndex := make(map[int64]struct{})
	for _, arg := range args {
		if _, ok := stepsIndex[arg.StepTime]; !ok {
			stepsIndex[arg.StepTime] = struct{}{}
			steps = append(steps, arg.StepTime)
		}
	}
	if len(steps) == 1 {
		return steps[0], false
	}
	commonStep = LCM(steps...)
	return commonStep, true
}

// GetStepRange returns min(steps), changed (bool) for slice of metrics.
// If all metrics have the same step, changed == false.
func GetStepRange(args []*types.MetricData) (minStep, maxStep int64, needScale bool) {
	minStep = args[0].StepTime
	maxStep = args[0].StepTime
	for _, arg := range args {
		if minStep > arg.StepTime {
			minStep = arg.StepTime
		}
		if maxStep < arg.StepTime {
			maxStep = arg.StepTime
		}
	}
	needScale = minStep != maxStep

	return
}

// ScaleToCommonStep returns the metrics, aligned LCM of all metrics steps.
// If commonStep == 0, then it will be calculated automatically
// It respects xFilesFactor and fills gaps in the begin and end with NaNs if needed.
func ScaleToCommonStep(args []*types.MetricData, commonStep int64) []*types.MetricData {
	if commonStep < 0 || len(args) == 0 {
		// This doesn't make sence
		return args
	}

	// If it's invoked with commonStep other than 0, changes are applied by default
	if commonStep == 0 {
		commonStep, _ = GetCommonStep(args)
	}

	minStart := args[0].StartTime
	for _, arg := range args {
		if minStart > arg.StartTime {
			minStart = arg.StartTime
		}
	}
	minStart -= (minStart % commonStep) // align StartTime against step

	maxVals := 0

	for _, arg := range args {
		if arg.StepTime == commonStep {
			if minStart < arg.StartTime {
				valCnt := (arg.StartTime - minStart) / arg.StepTime
				newVals := genNaNs(int(valCnt))
				arg.Values = append(newVals, arg.Values...)
			}
			arg.StartTime = minStart

			if len(arg.Values) > maxVals {
				maxVals = len(arg.Values)
			}
		} else {
			// arg = arg.Copy(true)
			// args[a] = arg
			stepFactor := commonStep / arg.StepTime
			// newStart := minStart - (arg.StartTime % commonStep)
			if arg.StartTime > minStart {
				// Fill with NaNs from newStart to arg.StartTime
				valCnt := (arg.StartTime - minStart) / arg.StepTime
				nans := genNaNs(int(valCnt))
				arg.Values = append(nans, arg.Values...)
				arg.StartTime = minStart
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

			if len(arg.Values) > maxVals {
				maxVals = len(arg.Values)
			}
		}
	}

	for _, arg := range args {
		if maxVals > len(arg.Values) {
			valCnt := maxVals - len(arg.Values)
			newVals := genNaNs(valCnt)
			arg.Values = append(arg.Values, newVals...)
		}
		arg.RecalcStopTime()
	}

	return args
}

// GetInterval returns minStartTime, maxStartTime for slice of metrics.
func GetInterval(args []*types.MetricData) (minStartTime, maxStopTime int64) {
	minStartTime = args[0].StartTime
	maxStopTime = args[0].StopTime
	for _, arg := range args {
		if minStartTime > arg.StartTime {
			minStartTime = arg.StartTime
		}

		arg.FixStopTime()
		if maxStopTime < arg.StopTime {
			maxStopTime = arg.StopTime
		}
	}

	return
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

// ScaleValuesToCommonStep returns map[parser.MetricRequest][]*types.MetricData. If any element of []*types.MetricData is changed, it doesn't change original
// metric, but creates the new one to avoid cache spoiling.
func ScaleValuesToCommonStep(values map[parser.MetricRequest][]*types.MetricData) map[parser.MetricRequest][]*types.MetricData {
	// Calculate global commonStep
	var args []*types.MetricData
	for _, metrics := range values {
		args = append(args, metrics...)
	}

	commonStep, changed := GetCommonStep(args)
	if !changed {
		return values
	}

	for m, metrics := range values {
		values[m] = ScaleToCommonStep(metrics, commonStep)
	}

	return values
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
	minStart, maxStop := GetInterval(args)

	if ExtrapolatePoints {
		minStepTime, _, needScale := GetStepRange(args)
		if needScale {
			for _, arg := range args {
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
	}

	for _, arg := range args {
		if minStart < arg.StartTime {
			valCnt := (arg.StartTime - minStart) / arg.StepTime
			newVals := genNaNs(int(valCnt))
			arg.Values = append(newVals, arg.Values...)
		}

		arg.StartTime = minStart

		if maxStop > arg.StopTime {
			valCnt := (maxStop - arg.StopTime) / arg.StepTime
			newVals := genNaNs(int(valCnt))
			arg.Values = append(arg.Values, newVals...)
			arg.StopTime = maxStop
		}

		arg.RecalcStopTime()
	}

	return args
}

// ScaleSeries aligns and scale different series together. By default it only prepends and appends NaNs in case of different length, but if ExtrapolatePoints is enabled, it can extrapolate
func ScaleSeries(args []*types.MetricData) []*types.MetricData {
	minStart, maxStop := GetInterval(args)
	var commonStep int64
	var needScale bool

	if ExtrapolatePoints {
		commonStep, _, needScale = GetStepRange(args)
		if needScale {
			for _, arg := range args {
				if arg.StepTime > commonStep {
					valsCnt := int(math.Ceil(float64(arg.StopTime-arg.StartTime) / float64(commonStep)))
					newVals := make([]float64, valsCnt)
					ts := arg.StartTime
					nextTs := arg.StartTime + arg.StepTime
					i := 0
					j := 0
					pointsPerInterval := float64(ts-nextTs) / float64(commonStep)
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
						ts += commonStep
						i++
					}
					arg.Values = newVals
					arg.StepTime = commonStep
				}
			}
			needScale = false
		}
	} else {
		commonStep, needScale = GetCommonStep(args)
	}

	if needScale {
		ScaleToCommonStep(args, commonStep)
	} else {
		maxVals := 0

		for _, arg := range args {
			if minStart < arg.StartTime {
				valCnt := (arg.StartTime - minStart) / arg.StepTime
				newVals := genNaNs(int(valCnt))
				arg.Values = append(newVals, arg.Values...)
				arg.StartTime = minStart
			}

			if maxStop > arg.StopTime {
				valCnt := (maxStop - arg.StopTime) / arg.StepTime
				newVals := genNaNs(int(valCnt))
				arg.Values = append(arg.Values, newVals...)
				arg.StopTime = maxStop
			}

			if maxVals < len(arg.Values) {
				maxVals = len(arg.Values)
			}
		}

		for _, arg := range args {
			if maxVals > len(arg.Values) {
				valCnt := maxVals - len(arg.Values)
				newVals := genNaNs(valCnt)
				arg.Values = append(arg.Values, newVals...)
			}
			arg.RecalcStopTime()
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
