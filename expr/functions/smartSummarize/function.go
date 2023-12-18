package smartSummarize

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type smartSummarize struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &smartSummarize{}
	functions := []string{"smartSummarize"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// smartSummarize(seriesList, intervalString, func='sum', alignTo=None)
func (f *smartSummarize) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(dgryski): make sure the arrays are all the same 'size'
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	alignToInterval, err := e.GetStringNamedOrPosArgDefault("alignTo", 3, "")
	if err != nil {
		return nil, err
	}

	if alignToInterval != "" {
		// Note: the start time for the fetch request is adjusted in expr.Metrics() so that the fetched
		// data is already aligned by interval if this parameter specifics an interval to align to
		newStart, err := parser.StartAlignTo(from, alignToInterval)
		if err != nil {
			return nil, err
		}
		from = newStart
	}
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return []*types.MetricData{}, nil
	}

	bucketSizeInt32, err := e.GetIntervalArg(1, 1)
	if err != nil {
		return nil, err
	}
	if bucketSizeInt32 <= 0 {
		return nil, parser.ErrInvalidInterval
	}
	bucketSize := int64(bucketSizeInt32)
	bucketSizeStr := e.Arg(1).StringValue()

	summarizeFunction, err := e.GetStringNamedOrPosArgDefault("func", 2, "sum")
	if err != nil {
		return nil, err
	}
	if err := consolidations.CheckValidConsolidationFunc(summarizeFunction); err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, len(args))
	for n, arg := range args {
		var name string

		if alignToInterval != "" {
			name = "smartSummarize(" + arg.Name + ",'" + bucketSizeStr + "','" + summarizeFunction + "','" + alignToInterval + "')"
		} else {
			name = "smartSummarize(" + arg.Name + ",'" + bucketSizeStr + "','" + summarizeFunction + "')"
		}

		r := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				StepTime:          bucketSize,
				StartTime:         arg.StartTime,
				StopTime:          arg.StopTime,
				ConsolidationFunc: arg.ConsolidationFunc,
			},
			Tags: helper.CopyTags(arg),
		}
		r.Tags["smartSummarize"] = fmt.Sprintf("%d", bucketSizeInt32)
		r.Tags["smartSummarizeFunction"] = summarizeFunction

		ts := arg.StartTime
		for ts < arg.StopTime {
			bucketUpperBound := ts + bucketSize
			bucketStart := (ts - arg.StartTime + arg.StepTime - 1) / arg.StepTime             // equivalent to ceil((ts-arg.StartTime) / arg.StepTime)
			bucketEnd := (bucketUpperBound - arg.StartTime + arg.StepTime - 1) / arg.StepTime // equivalent to ceil((until-arg.StartTime) / arg.StepTime)

			if bucketEnd > int64(len(arg.Values)) {
				bucketEnd = int64(len(arg.Values))
			}

			rv := consolidations.SummarizeValues(summarizeFunction, arg.Values[bucketStart:bucketEnd], arg.XFilesFactor)
			r.Values = append(r.Values, rv)
			ts = bucketUpperBound
		}
		r.StopTime = ts
		results[n] = &r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *smartSummarize) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"smartSummarize": {
			Description: "Smarter version of summarize.\nThe alignToFrom boolean parameter has been replaced by alignTo and no longer has any effect. Alignment can be to years, months, weeks, days, hours, and minutes.\nThis function can be used with aggregation functions average, median, sum, min, max, diff, stddev, count, range, multiply & last.",
			Function:    "smartSummarize(seriesList, intervalString, func='sum', alignTo=None)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "smartSummarize",
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
					Default: types.NewSuggestion("sum"),
					Name:    "func",
					Options: types.StringsToSuggestionList(consolidations.AvailableSummarizers),
					Type:    types.AggFunc,
				},
				{
					Name: "alignTo",
					Suggestions: types.NewSuggestions(
						"1m",
						"1d",
						"1y",
					),
					Type: types.Interval,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
