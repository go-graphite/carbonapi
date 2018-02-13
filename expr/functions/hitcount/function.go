package hitcount

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
)

func init() {
	metadata.RegisterFunction("hitcount", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// hitcount(seriesList, intervalString, alignToInterval=False)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	bucketSize, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}

	alignToInterval, err := e.GetBoolNamedOrPosArgDefault("alignToInterval", 2, false)
	if err != nil {
		return nil, err
	}
	_, ok := e.NamedArgs()["alignToInterval"]
	if !ok {
		ok = len(e.Args()) > 2
	}

	start := args[0].StartTime
	stop := args[0].StopTime
	if alignToInterval {
		start = helper.AlignStartToInterval(start, stop, bucketSize)
	}

	buckets := helper.GetBuckets(start, stop, bucketSize)
	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {

		name := fmt.Sprintf("hitcount(%s,'%s'", arg.Name, e.Args()[1].StringValue())
		if ok {
			name += fmt.Sprintf(",%v", alignToInterval)
		}
		name += ")"

		r := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      name,
			Values:    make([]float64, buckets, buckets+1),
			IsAbsent:  make([]bool, buckets, buckets+1),
			StepTime:  bucketSize,
			StartTime: start,
			StopTime:  stop,
		}}

		bucketEnd := start + bucketSize
		t := arg.StartTime
		ridx := 0
		var count float64
		bucketItems := 0
		for i, v := range arg.Values {
			bucketItems++
			if !arg.IsAbsent[i] {
				if math.IsNaN(count) {
					count = 0
				}

				count += v * float64(arg.StepTime)
			}

			t += arg.StepTime

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
	return results, nil
}
