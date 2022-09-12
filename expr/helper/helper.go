package helper

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

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
		if err != nil {
			return nil, err
		}
		args = append(args, a...)
	}

	if len(args) == 0 {
		return nil, nil
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
	if len(args) < e.ArgsLen() {
		e.SetRawArgs(RemoveEmptySeriesFromName(args))
	}

	return args, nil
}

// AggKey returns joined by dot nodes of tags names
func AggKey(arg *types.MetricData, nodesOrTags []parser.NodeOrTag) string {
	matched := make([]string, 0, len(nodesOrTags))
	metricTags := arg.Tags
	name := types.ExtractNameTag(arg.Name)
	nodes := strings.Split(name, ".")
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

type seriesFunc1 func(*types.MetricData) *types.MetricData

// ForEachSeriesDo do action for each serie in list.
func ForEachSeriesDo1(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData, function seriesFunc1) ([]*types.MetricData, error) {
	arg, err := GetSeriesArg(ctx, e.Arg(0), from, until, values)
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
	arg, err := GetSeriesArg(ctx, e.Arg(0), from, until, values)
	if err != nil {
		return nil, parser.ErrMissingTimeseries
	}
	var results []*types.MetricData

	for _, a := range arg {
		r := a.CopyName(e.Target() + "(" + a.Name + ")")
		r.Values = make([]float64, len(a.Values))
		results = append(results, function(a, r))
	}
	return results, nil
}

// AggregateFunc type that defined aggregate function
type AggregateFunc func([]float64) float64

// AggregateSeries aggregates series
func AggregateSeries(e parser.Expr, args []*types.MetricData, function AggregateFunc, xFilesFactor float64, extractTagsFromArgs bool) ([]*types.MetricData, error) {
	if len(args) == 0 {
		// GraphiteWeb does this, no matter the function
		// https://github.com/graphite-project/graphite-web/blob/b52987ac97f49dcfb401a21d4b92860cfcbcf074/webapp/graphite/render/functions.py#L228
		return args, nil
	}

	var applyXFilesFactor = xFilesFactor >= 0

	args = ScaleSeries(args)
	length := len(args[0].Values)
	r := args[0].CopyNameArg(e.Target()+"("+e.RawArgs()+")", e.Target(), GetCommonTags(args), extractTagsFromArgs)
	r.Values = make([]float64, length)

	if _, ok := r.Tags["name"]; !ok {
		r.Tags["name"] = r.Name
	}

	values := make([]float64, len(args))
	for i := range args[0].Values {
		for n, arg := range args {
			values[n] = arg.Values[i]
		}

		r.Values[i] = math.NaN()
		if len(values) > 0 {
			if applyXFilesFactor && XFilesFactorValues(values, xFilesFactor) {
				r.Values[i] = function(values)
			} else {
				r.Values[i] = function(values)
			}
		}
	}

	return []*types.MetricData{r}, nil
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

// CopyTags makes a deep copy of the tags
func CopyTags(series *types.MetricData) map[string]string {
	out := make(map[string]string, len(series.Tags))
	for k, v := range series.Tags {
		out[k] = v
	}
	return out
}

func GetCommonTags(series []*types.MetricData) map[string]string {
	if len(series) == 0 {
		return make(map[string]string)
	}
	commonTags := CopyTags(series[0])
	for _, serie := range series {
		for k, v := range commonTags {
			if serie.Tags[k] != v {
				delete(commonTags, k)
			}
		}
	}

	return commonTags
}

func SafeRound(x float64, precision int) float64 {
	if math.IsNaN(x) {
		return x
	}
	roundTo := math.Pow10(precision)
	return math.RoundToEven(x*roundTo) / roundTo
}

func XFilesFactorValues(values []float64, xFilesFactor float64) bool {
	if math.IsNaN(xFilesFactor) || xFilesFactor == 0 {
		return true
	}
	nonNull := 0
	for _, val := range values {
		if !math.IsNaN(val) {
			nonNull++
		}
	}
	return XFilesFactor(nonNull, len(values), xFilesFactor)
}

func XFilesFactor(nonNull int, total int, xFilesFactor float64) bool {
	if nonNull < 0 || total <= 0 {
		return false
	}
	return float64(nonNull)/float64(total) >= xFilesFactor
}

type unitPrefix struct {
	prefix string
	size   uint64
}

const floatEpsilon = 0.00000000001

const (
	unitSystemBinary = "binary"
	unitSystemSI     = "si"
)

var UnitSystems = map[string][]unitPrefix{
	unitSystemBinary: {
		{"Pi", 1125899906842624}, // 1024^5
		{"Ti", 1099511627776},    // 1024^4
		{"Gi", 1073741824},       // 1024^3
		{"Mi", 1048576},          // 1024^2
		{"Ki", 1024},
	},
	unitSystemSI: {
		{"P", 1000000000000000}, // 1000^5
		{"T", 1000000000000},    // 1000^4
		{"G", 1000000000},       // 1000^3
		{"M", 1000000},          // 1000^2
		{"K", 1000},
	},
}

// formatUnits formats the given value according to the given unit prefix system
func FormatUnits(v float64, system string) (float64, string) {
	unitsystem := UnitSystems[system]
	for _, p := range unitsystem {
		fsize := float64(p.size)
		if math.Abs(v) >= fsize {
			v2 := v / fsize
			if (v2-math.Floor(v2)) < floatEpsilon && v > 1 {
				v2 = math.Floor(v2)
			}
			return v2, p.prefix
		}
	}
	if (v-math.Floor(v)) < floatEpsilon && v > 1 {
		v = math.Floor(v)
	}
	return v, ""
}
