package helper

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/expr/helper/metric"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

var evaluator interfaces.Evaluator

// Backref is a pre-compiled expression for backref
var Backref = regexp.MustCompile(`\\(\d+)`)

// ErrUnknownFunction is an error message about unknown function
type ErrUnknownFunction string

func (e ErrUnknownFunction) Error() string {
	return fmt.Sprintf("unknown function in evalExpr: %q", string(e))
}

// SetEvaluator sets evaluator for all helper functions
func SetEvaluator(e interfaces.Evaluator) {
	evaluator = e
}

// GetSeriesArg returns argument from series.
func GetSeriesArg(ctx context.Context, arg parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if !arg.IsName() && !arg.IsFunc() {
		return nil, parser.ErrMissingTimeseries
	}

	a, err := evaluator.Eval(ctx, arg, from, until, values)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// RemoveEmptySeriesFromName removes empty series from list of names.
func RemoveEmptySeriesFromName(args []*types.MetricData) string {
	var argNames []string
	for _, arg := range args {
		argNames = append(argNames, arg.Name)
	}

	return strings.Join(argNames, ",")
}

// GetSeriesArgs returns arguments of series
func GetSeriesArgs(ctx context.Context, e []parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	var args []*types.MetricData

	for _, arg := range e {
		a, err := GetSeriesArg(ctx, arg, from, until, values)
		if err != nil && !merry.Is(err, parser.ErrSeriesDoesNotExist) {
			return nil, err
		}
		args = append(args, a...)
	}

	if len(args) == 0 {
		return nil, parser.ErrSeriesDoesNotExist
	}

	return args, nil
}

// GetSeriesArgsAndRemoveNonExisting will fetch all required arguments, but will also filter out non existing Series
// This is needed to be graphite-web compatible in cases when you pass non-existing Series to, for example, sumSeries
func GetSeriesArgsAndRemoveNonExisting(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := GetSeriesArgs(ctx, e.Args(), from, until, values)
	if err != nil {
		return nil, err
	}

	// We need to rewrite name if there are some missing metrics
	if len(args) < len(e.Args()) {
		e.SetRawArgs(RemoveEmptySeriesFromName(args))
	}

	return args, nil
}

// AggKey returns joined by dot nodes of tags names
func AggKey(arg *types.MetricData, nodesOrTags []parser.NodeOrTag) string {
	matched := make([]string, 0, len(nodesOrTags))
	metricTags := arg.Tags
	nodes := strings.Split(metricTags["name"], ".")
	for _, nt := range nodesOrTags {
		if nt.IsTag {
			tagStr := nt.Value.(string)
			matched = append(matched, metricTags[tagStr])
		} else {
			f := nt.Value.(int)
			if f < 0 {
				f += len(nodes)
			}
			if f >= len(nodes) || f < 0 {
				continue
			}
			matched = append(matched, nodes[f])
		}
	}
	if len(matched) > 0 {
		return strings.Join(matched, ".")
	}
	return ""
}

// AggKey returns joined by dot nodes of tags names
func AggKeyInt(arg *types.MetricData, ints []int) string {
	matched := make([]string, 0, len(ints))
	nodes := strings.Split(arg.Tags["name"], ".")
	for _, f := range ints {
		if f < 0 {
			f += len(nodes)
		}
		if f >= len(nodes) || f < 0 {
			continue
		}
		matched = append(matched, nodes[f])
	}
	if len(matched) > 0 {
		return strings.Join(matched, ".")
	}
	return ""
}

type seriesFunc1 func(*types.MetricData) *types.MetricData

// ForEachSeriesDo do action for each serie in list.
func ForEachSeriesDo1(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData, function seriesFunc1) ([]*types.MetricData, error) {
	arg, err := GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, parser.ErrMissingTimeseries
	}
	var results []*types.MetricData

	for _, a := range arg {
		results = append(results, function(a))
	}
	return results, nil
}

type seriesFunc func(*types.MetricData, *types.MetricData) *types.MetricData

// ForEachSeriesDo do action for each serie in list.
func ForEachSeriesDo(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData, function seriesFunc) ([]*types.MetricData, error) {
	arg, err := GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, parser.ErrMissingTimeseries
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = e.Target() + "(" + a.Name + ")"
		r.Values = make([]float64, len(a.Values))
		results = append(results, function(a, &r))
	}
	return results, nil
}

// AggregateFunc type that defined aggregate function
type AggregateFunc func([]float64) float64

// AggregateSeries aggregates series
func AggregateSeries(e parser.Expr, args []*types.MetricData, function AggregateFunc) ([]*types.MetricData, error) {
	if len(args) == 0 {
		return args, nil
	}

	args = ScaleSeries(args)

	length := len(args[0].Values)
	r := *args[0]
	r.Name = e.Target() + "(" + e.RawArgs() + ")"
	r.Tags = map[string]string{"name": metric.ExtractMetric(e.RawArgs())}
	r.Values = make([]float64, length)

	values := make([]float64, len(args))
	for i := range args[0].Values {
		for n, arg := range args {
			values[n] = arg.Values[i]
		}

		r.Values[i] = math.NaN()
		if len(values) > 0 {
			r.Values[i] = function(values)
		}
	}

	return []*types.MetricData{&r}, nil
}

// Contains check if slice 'a' contains value 'i'
func Contains(a []int, i int) bool {
	for _, aa := range a {
		if aa == i {
			return true
		}
	}
	return false
}
