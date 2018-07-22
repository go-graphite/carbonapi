package tests

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type FuncEvaluator struct {
	eval func(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

func (evaluator *FuncEvaluator) EvalExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.IsName() {
		return values[parser.MetricRequest{Metric: e.Target(), From: from, Until: until}], nil
	} else if e.IsConst() {
		p := types.MetricData{FetchResponse: pb.FetchResponse{Name: e.Target(), Values: []float64{e.FloatValue()}}}
		return []*types.MetricData{&p}, nil
	}
	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.Args()) == 0 {
		return nil, parser.ErrMissingArgument
	}

	return evaluator.eval(e, from, until, values)
}

func EvaluatorFromFunc(function interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: function.Do,
	}

	return e
}

func EvaluatorFromFuncWithMetadata(metadata map[string]interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: func(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
			if f, ok := metadata[e.Target()]; ok {
				return f.Do(e, from, until, values)
			}
			return nil, fmt.Errorf("unknown function: %v", e.Target())
		},
	}
	return e
}

// FuncRewriter is a struct that can evaluate rewrite func
// It's useful for tests for rewrite functions. See how to initialize it below
type FuncRewriter struct {
	eval func(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error)
}

// RewriteExpr rewrites expressions (see expr/expr.go RewriteExpr for more details
func (evaluator *FuncRewriter) RewriteExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	return evaluator.eval(e, from, until, values)
}

// RewriterFromFunc creates a struct that interface.Rewriter compatible and only executes specified function
func RewriterFromFunc(function interfaces.RewriteFunction) interfaces.Rewriter {
	e := &FuncRewriter{
		eval: function.Do,
	}

	return e
}

func DeepClone(original map[parser.MetricRequest][]*types.MetricData) map[parser.MetricRequest][]*types.MetricData {
	clone := map[parser.MetricRequest][]*types.MetricData{}
	for key, originalMetrics := range original {
		copiedMetrics := []*types.MetricData{}
		for _, originalMetric := range originalMetrics {
			copiedMetric := types.MetricData{
				FetchResponse: pb.FetchResponse{
					Name:      originalMetric.Name,
					StartTime: originalMetric.StartTime,
					StopTime:  originalMetric.StopTime,
					StepTime:  originalMetric.StepTime,
					Values:    make([]float64, len(originalMetric.Values)),
				},
			}

			copy(copiedMetric.Values, originalMetric.Values)
			copiedMetrics = append(copiedMetrics, &copiedMetric)
		}

		clone[key] = copiedMetrics
	}

	return clone
}

func compareFloat64(v1, v2 float64) bool {
	if math.IsNaN(v1) && math.IsNaN(v2) {
		return true
	}
	if math.IsInf(v1, 1) && math.IsInf(v2, 1) {
		return true
	}

	if math.IsInf(v1, 0) && math.IsInf(v2, 0) {
		return true
	}

	d := math.Abs(v1 - v2)
	return d < eps
}

func deepCompareFields(v1, v2 reflect.Value) bool {
	if !v1.CanInterface() {
		return true
	}
	t1 := v1.Type()
	if t1.Comparable() {
		if t1.Name() == "float64" {
			return compareFloat64(v1.Interface().(float64), v2.Interface().(float64))
		}
		if t1.Name() == "float32" {
			v1f64 := float64(v1.Interface().(float32))
			v2f64 := float64(v2.Interface().(float32))
			return compareFloat64(v1f64, v2f64)
		}
		return reflect.DeepEqual(v1.Interface(), v2.Interface())
	} else {
		switch v1.Kind() {
		case reflect.Struct:
			if v1.NumField() == 0 {
				// We don't know how to compare that
				return false
			}
			for i := 0; i < v1.NumField(); i++ {
				r := deepCompareFields(v1.Field(i), v2.Field(i))
				if !r {
					return r
				}
			}
		case reflect.Slice, reflect.Array:
			if v1.Len() != v2.Len() {
				return false
			}
			if v1.Len() == 0 {
				return true
			}
			if v1.Index(0).Kind() != v2.Index(0).Kind() {
				return false
			}
			for i := 0; i < v1.Len(); i++ {
				e1 := v1.Index(i)
				e2 := v2.Index(i)
				if !deepCompareFields(e1, e2) {
					return false
				}
			}
		case reflect.Func:
			return v1.Pointer() == v2.Pointer()
		default:
			fmt.Printf("unsupported v1.Kind=%v t1.Name=%v, t1.Value=%v\n\n", v1.Kind(), v1.Type().Name(), v1.String())
			return false
		}
	}
	return true
}

func MetricDataIsEqual(d1, d2 *types.MetricData) bool {
	v1 := reflect.ValueOf(*d1)
	v2 := reflect.ValueOf(*d2)

	for i := 0; i < v1.NumField(); i++ {
		r := deepCompareFields(v1.Field(i), v2.Field(i))
		if !r {
			return r
		}
	}
	return true
}

func DeepEqual(t *testing.T, target string, original, modified map[parser.MetricRequest][]*types.MetricData) {
	for key := range original {
		if len(original[key]) == len(modified[key]) {
			for i := range original[key] {
				if !MetricDataIsEqual(original[key][i], modified[key][i]) {
					t.Errorf(
						"%s: source data was modified key %v index %v original:\n%v\n modified:\n%v",
						target,
						key,
						i,
						original[key][i],
						modified[key][i],
					)
				}
			}
		} else {
			t.Errorf(
				"%s: source data was modified key %v original length %d, new length %d",
				target,
				key,
				len(original[key]),
				len(modified[key]),
			)
		}
	}
}

const eps = 0.0000000001

func NearlyEqual(a []float64, b []float64) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		// "same"
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		}
		if math.IsNaN(a[i]) || math.IsNaN(b[i]) {
			// unexpected NaN
			return false
		}
		// "close enough"
		if math.Abs(v-b[i]) > eps {
			return false
		}
	}

	return true
}

func NearlyEqualMetrics(a, b *types.MetricData) bool {
	for i := range a.Values {
		if (math.IsNaN(a.Values[i]) && !math.IsNaN(b.Values[i])) || (!math.IsNaN(a.Values[i]) && math.IsNaN(b.Values[i])) {
			return false
		}
		// "close enough"
		if math.Abs(a.Values[i]-b.Values[i]) > eps {
			return false
		}
	}

	return true
}

type SummarizeEvalTestItem struct {
	E     parser.Expr
	M     map[parser.MetricRequest][]*types.MetricData
	W     []float64
	Name  string
	Step  int64
	Start int64
	Stop  int64
}

func InitTestSummarize() (int64, int64, int64) {
	t0, err := time.Parse(time.UnixDate, "Wed Sep 10 10:32:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirtyTwo := t0.Unix()

	t0, err = time.Parse(time.UnixDate, "Wed Sep 10 10:59:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenFiftyNine := t0.Unix()

	t0, err = time.Parse(time.UnixDate, "Wed Sep 10 10:30:00 CEST 2014")
	if err != nil {
		panic(err)
	}

	tenThirty := t0.Unix()

	return tenThirtyTwo, tenFiftyNine, tenThirty
}

func TestSummarizeEvalExpr(t *testing.T, tt *SummarizeEvalTestItem) {
	evaluator := metadata.GetEvaluator()

	t.Run(tt.Name, func(t *testing.T) {
		originalMetrics := DeepClone(tt.M)
		g, err := evaluator.EvalExpr(tt.E, 0, 1, tt.M)
		if err != nil {
			t.Errorf("failed to eval %v: %+v", tt.Name, err)
			return
		}
		DeepEqual(t, g[0].Name, originalMetrics, tt.M)
		if g[0].StepTime != tt.Step {
			t.Errorf("bad Step for %s:\ngot  %d\nwant %d", g[0].Name, g[0].StepTime, tt.Step)
		}
		if g[0].StartTime != tt.Start {
			t.Errorf("bad Start for %s: got %s want %s", g[0].Name, time.Unix(g[0].StartTime, 0).Format(time.StampNano), time.Unix(tt.Start, 0).Format(time.StampNano))
		}
		if g[0].StopTime != tt.Stop {
			t.Errorf("bad Stop for %s: got %s want %s", g[0].Name, time.Unix(g[0].StopTime, 0).Format(time.StampNano), time.Unix(tt.Stop, 0).Format(time.StampNano))
		}

		if !NearlyEqual(g[0].Values, tt.W) {
			t.Errorf("failed: %s:\ngot  %+v,\nwant %+v", g[0].Name, g[0].Values, tt.W)
		}
		if g[0].Name != tt.Name {
			t.Errorf("bad Name for %+v: got %v, want %v", g, g[0].Name, tt.Name)
		}
	})
}

type MultiReturnEvalTestItem struct {
	E       parser.Expr
	M       map[parser.MetricRequest][]*types.MetricData
	Name    string
	Results map[string][]*types.MetricData
}

func TestMultiReturnEvalExpr(t *testing.T, tt *MultiReturnEvalTestItem) {
	evaluator := metadata.GetEvaluator()

	originalMetrics := DeepClone(tt.M)
	g, err := evaluator.EvalExpr(tt.E, 0, 1, tt.M)
	if err != nil {
		t.Errorf("failed to eval %v: %+v", tt.Name, err)
		return
	}
	DeepEqual(t, tt.Name, originalMetrics, tt.M)
	if len(g) == 0 {
		t.Errorf("returned no data %v", tt.Name)
		return
	}
	if g[0] == nil {
		t.Errorf("returned no value %v", tt.Name)
		return
	}
	if g[0].StepTime == 0 {
		t.Errorf("missing Step for %+v", g)
	}
	if len(g) != len(tt.Results) {
		t.Errorf("unexpected results len: got %d, want %d", len(g), len(tt.Results))
	}
	for _, gg := range g {
		r, ok := tt.Results[gg.Name]
		if !ok {
			t.Errorf("missing result Name: %v", gg.Name)
			continue
		}
		if r[0].Name != gg.Name {
			t.Errorf("result Name mismatch, got\n%#v,\nwant\n%#v", gg.Name, r[0].Name)
		}
		if !reflect.DeepEqual(r[0].Values, gg.Values) ||
			r[0].StartTime != gg.StartTime ||
			r[0].StopTime != gg.StopTime ||
			r[0].StepTime != gg.StepTime {
			t.Errorf("result mismatch, got\n%#v,\nwant\n%#v", gg, r)
		}
	}
}

type EvalTestItem struct {
	E    parser.Expr
	M    map[parser.MetricRequest][]*types.MetricData
	Want []*types.MetricData
}

func TestEvalExpr(t *testing.T, tt *EvalTestItem) {
	evaluator := metadata.GetEvaluator()
	originalMetrics := DeepClone(tt.M)
	testName := tt.E.Target() + "(" + tt.E.RawArgs() + ")"
	g, err := evaluator.EvalExpr(tt.E, 0, 1, tt.M)
	if err != nil {
		t.Errorf("failed to eval %s: %+v", testName, err)
		return
	}
	if len(g) != len(tt.Want) {
		t.Errorf("%s returned a different number of metrics, actual %v, Want %v", testName, len(g), len(tt.Want))
		return

	}
	DeepEqual(t, testName, originalMetrics, tt.M)

	for i, want := range tt.Want {
		actual := g[i]
		if actual == nil {
			t.Errorf("returned no value %v", tt.E.RawArgs())
			return
		}
		if actual.StepTime == 0 {
			t.Errorf("missing Step for %+v", g)
		}
		if actual.Name != want.Name {
			t.Errorf("bad Name for %s metric %d: got %s, Want %s", testName, i, actual.Name, want.Name)
		}
		if !NearlyEqualMetrics(actual, want) {
			t.Errorf("different values for %s metric %s: got %v, Want %v", testName, actual.Name, actual.Values, want.Values)
			return
		}
	}
}
