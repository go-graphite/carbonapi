package summarize

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
	f := &function{}
	functions := []string{"summarize"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// summarize(seriesList, intervalString, func='sum', alignToFrom=False)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, nil
	}

	bucketSize, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}

	summarizeFunction, err := e.GetStringNamedOrPosArgDefault("func", 2, "sum")
	if err != nil {
		return nil, err
	}
	_, funcOk := e.NamedArgs()["func"]
	if !funcOk {
		funcOk = len(e.Args()) > 2
	}

	alignToFrom, err := e.GetBoolNamedOrPosArgDefault("alignToFrom", 3, false)
	if err != nil {
		return nil, err
	}
	_, alignOk := e.NamedArgs()["alignToFrom"]
	if !alignOk {
		alignOk = len(e.Args()) > 3
	}

	start := args[0].StartTime
	stop := args[0].StopTime
	if !alignToFrom {
		start, stop = helper.AlignToBucketSize(start, stop, bucketSize)
	}

	buckets := helper.GetBuckets(start, stop, bucketSize)
	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {

		name := fmt.Sprintf("summarize(%s,'%s'", arg.Name, e.Args()[1].StringValue())
		if funcOk || alignOk {
			// we include the "func" argument in the presence of
			// "alignToFrom", even if the former was omitted
			// this is so that a call like "summarize(foo, '5min', alignToFrom=true)"
			// doesn't produce a metric name that has a boolean value
			// where a function name should be
			// so we show "summarize(foo,'5min','sum',true)" instead of "summarize(foo,'5min',true)"
			//
			// this does not match graphite's behaviour but seems more correct
			name += fmt.Sprintf(",'%s'", summarizeFunction)
		}
		if alignOk {
			name += fmt.Sprintf(",%v", alignToFrom)
		}
		name += ")"

		if arg.StepTime > bucketSize {
			// We don't have enough data to do math
			results = append(results, &types.MetricData{FetchResponse: pb.FetchResponse{
				Name:      name,
				Values:    arg.Values,
				IsAbsent:  arg.IsAbsent,
				StepTime:  arg.StepTime,
				StartTime: arg.StartTime,
				StopTime:  arg.StopTime,
			}})
			continue
		}

		r := types.MetricData{FetchResponse: pb.FetchResponse{
			Name:      name,
			Values:    make([]float64, buckets, buckets),
			IsAbsent:  make([]bool, buckets, buckets),
			StepTime:  bucketSize,
			StartTime: start,
			StopTime:  stop,
		}}

		t := arg.StartTime // unadjusted
		bucketEnd := start + bucketSize
		values := make([]float64, 0, bucketSize/arg.StepTime)
		ridx := 0
		bucketItems := 0
		for i, v := range arg.Values {
			bucketItems++
			if !arg.IsAbsent[i] {
				values = append(values, v)
			}

			t += arg.StepTime

			if t >= stop {
				break
			}

			if t >= bucketEnd {
				rv := helper.SummarizeValues(summarizeFunction, values)

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
			rv := helper.SummarizeValues(summarizeFunction, values)
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
	return results, nil
}
