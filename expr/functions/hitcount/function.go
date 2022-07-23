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
	eArgs := e.Args()
	// TODO(dgryski): make sure the arrays are all the same 'size'
	args, err := helper.GetSeriesArg(ctx, eArgs[0], from, until, values)
	if err != nil {
		return nil, err
	}

	bucketSizeInt32, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}
	bucketSize := int64(bucketSizeInt32)

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
		var nameBuf strings.Builder
		bucketSizeStr := eArgs[1].StringValue()
		nameBuf.Grow(len(arg.Name) + 13 + len(bucketSizeStr))
		nameBuf.WriteString("hitcount(")
		nameBuf.WriteString(arg.Name)
		nameBuf.WriteString(",'")
		nameBuf.WriteString(bucketSizeStr)
		nameBuf.WriteString("'")
		if ok {
			nameBuf.WriteString(",")
			nameBuf.WriteString(strconv.FormatBool(alignToInterval))
		}
		nameBuf.WriteString(")")

		r := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              nameBuf.String(),
				Values:            make([]float64, buckets, buckets+1),
				StepTime:          bucketSize,
				StartTime:         start,
				StopTime:          stop,
				ConsolidationFunc: "max",
			},
			Tags: arg.Tags,
		}

		bucketEnd := start + bucketSize
		t := arg.StartTime
		ridx := 0
		var count float64
		bucketItems := 0
		for _, v := range arg.Values {
			bucketItems++
			if !math.IsNaN(v) {
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
				r.Values[ridx] = count

				ridx++
				bucketEnd += bucketSize
				count = math.NaN()
				bucketItems = 0
			}
		}

		// remaining values
		if bucketItems > 0 {
			r.Values[ridx] = count
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
