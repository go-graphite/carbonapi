package hitcount

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type hitcount struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &hitcount{}
	functions := []string{"hitcount"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// hitcount(seriesList, intervalString, alignToInterval=False)
func (f *hitcount) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	bucketSizeInt32, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}
	interval := int64(bucketSizeInt32)

	alignToInterval, err := e.GetBoolNamedOrPosArgDefault("alignToInterval", 2, false)
	if err != nil {
		return nil, err
	}

	if alignToInterval {
		// from needs to be adjusted before grabbing the series arg as it has been adjusted in the metric request
		from = helper.AlignStartToInterval(from, until, interval)
	}

	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return []*types.MetricData{}, nil
	}

	start := args[0].StartTime
	stop := args[0].StopTime

	// Note: the start time for the fetch request is adjusted in expr.Metrics() so that the fetched
	// data is already aligned by interval if this parameter is set to true
	if alignToInterval {
		intervalCount := (stop - start) / interval
		stop = start + (intervalCount * interval) + interval
	}

	results := make([]*types.MetricData, 0, len(args))
	for _, arg := range args {
		var nameBuf strings.Builder
		bucketSizeStr := e.Arg(1).StringValue()
		nameBuf.Grow(len(arg.Name) + 13 + len(bucketSizeStr))
		nameBuf.WriteString("hitcount(")
		nameBuf.WriteString(arg.Name)
		nameBuf.WriteString(",'")
		nameBuf.WriteString(bucketSizeStr)
		nameBuf.WriteString("'")
		if alignToInterval {
			nameBuf.WriteString(",")
			nameBuf.WriteString(strconv.FormatBool(alignToInterval))
		}
		nameBuf.WriteString(")")

		bucketCount := helper.GetBuckets(start, stop, interval)

		r := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:      nameBuf.String(),
				StepTime:  interval,
				StartTime: start,
				StopTime:  stop,
			},
			Tags: helper.CopyTags(arg),
		}
		r.Tags["hitcount"] = strconv.FormatInt(int64(bucketSizeInt32), 10)

		step := arg.StepTime
		buckets := make([][]float64, bucketCount)
		newStart := stop - bucketCount*interval
		r.StartTime = newStart

		for i, v := range arg.Values {
			if math.IsNaN(v) {
				continue
			}

			start_time := arg.StartTime + int64(i)*step
			startBucket, startMod := helper.Divmod(start_time-newStart, interval)
			end_time := start_time + step
			endBucket, endMod := helper.Divmod(end_time-newStart, interval)

			if endBucket >= bucketCount {
				endBucket = bucketCount - 1
				endMod = interval
			}

			if startBucket == endBucket {
				// All hits go into a single bucket
				if startBucket >= 0 {
					buckets[startBucket] = append(buckets[startBucket], v*float64(endMod-startMod))
				}
			} else {
				// Spread the hits amongst 2 or more buckets
				if startBucket >= 0 {
					buckets[startBucket] = append(buckets[startBucket], v*float64(interval-startMod))
				}
				hitsPerBucket := v * float64(interval)
				for j := startBucket + 1; j < endBucket; j++ {
					buckets[j] = append(buckets[j], hitsPerBucket)
				}
				if endMod > 0 {
					buckets[endBucket] = append(buckets[endBucket], v*float64(endMod))
				}
			}
		}
		r.Values = make([]float64, len(buckets))
		for i, bucket := range buckets {
			if len(bucket) != 0 {
				var sum float64
				for _, v := range bucket {
					sum += v
				}
				r.Values[i] = sum
			} else {
				r.Values[i] = math.NaN()
			}
		}

		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *hitcount) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"hitcount": {
			Description: "Estimate hit counts from a list of time series.\n\nThis function assumes the values in each time series represent\nhits per second.  It calculates hits per some larger interval\nsuch as per day or per hour.  This function is like summarize(),\nexcept that it compensates automatically for different time scales\n(so that a similar graph results from using either fine-grained\nor coarse-grained records) and handles rarely-occurring events\ngracefully.",
			Function:    "hitcount(seriesList, intervalString, alignToInterval=False)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "hitcount",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "intervalString",
					Required: true,
					Suggestions: types.NewSuggestions(
						"10min",
						"1h",
						"1d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion(false),
					Name:    "alignToInterval",
					Type:    types.Boolean,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
