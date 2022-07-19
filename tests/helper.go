package tests

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/ansel1/merry"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type FuncEvaluator struct {
	eval func(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

func (evaluator *FuncEvaluator) Eval(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.IsName() {
		return values[parser.MetricRequest{Metric: e.Target(), From: from, Until: until}], nil
	} else if e.IsConst() {
		p := types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:   e.Target(),
				Values: []float64{e.FloatValue()},
			},
			Tags: map[string]string{"name": e.Target()},
		}
		return []*types.MetricData{&p}, nil
	}
	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.Args()) == 0 {
		return nil, parser.ErrMissingArgument
	}

	if evaluator.eval != nil {
		return evaluator.eval(context.Background(), e, from, until, values)
	}

	return nil, helper.ErrUnknownFunction(e.Target())
}

func DummyEvaluator() interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: nil,
	}

	return e
}

func EvaluatorFromFunc(function interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: function.Do,
	}

	return e
}

func EvaluatorFromFuncWithMetadata(metadata map[string]interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
		eval: func(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
			if f, ok := metadata[e.Target()]; ok {
				return f.Do(context.Background(), e, from, until, values)
			}
			return nil, fmt.Errorf("unknown function: %v", e.Target())
		},
	}
	return e
}

func DeepClone(original map[parser.MetricRequest][]*types.MetricData) map[parser.MetricRequest][]*types.MetricData {
	clone := map[parser.MetricRequest][]*types.MetricData{}
	for key, originalMetrics := range original {
		copiedMetrics := make([]*types.MetricData, 0, len(originalMetrics))
		for _, originalMetric := range originalMetrics {
			copiedMetric := types.MetricData{
				FetchResponse: pb.FetchResponse{
					Name:                    originalMetric.Name,
					StartTime:               originalMetric.StartTime,
					StopTime:                originalMetric.StopTime,
					StepTime:                originalMetric.StepTime,
					Values:                  make([]float64, len(originalMetric.Values)),
					PathExpression:          originalMetric.PathExpression,
					ConsolidationFunc:       originalMetric.ConsolidationFunc,
					XFilesFactor:            originalMetric.XFilesFactor,
					HighPrecisionTimestamps: originalMetric.HighPrecisionTimestamps,
					AppliedFunctions:        make([]string, len(originalMetric.AppliedFunctions)),
					RequestStartTime:        originalMetric.RequestStartTime,
					RequestStopTime:         originalMetric.RequestStopTime,
				},
				GraphOptions:      originalMetric.GraphOptions,
				ValuesPerPoint:    originalMetric.ValuesPerPoint,
				Tags:              make(map[string]string),
				AggregateFunction: originalMetric.AggregateFunction,
			}

			copy(copiedMetric.Values, originalMetric.Values)
			copy(copiedMetric.AppliedFunctions, originalMetric.AppliedFunctions)
			copiedMetrics = append(copiedMetrics, &copiedMetric)
			for k, v := range originalMetric.Tags {
				copiedMetric.Tags[k] = v
			}
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
		case reflect.Map:
			if v1.Len() != v2.Len() {
				return false
			}
			if v1.Len() == 0 {
				return true
			}

			keys1 := v1.MapKeys()
			for _, k := range keys1 {
				val1 := v1.MapIndex(k)
				val2 := v2.MapIndex(k)
				if !deepCompareFields(val1, val2) {
					return false
				}
			}
			return true
		case reflect.Func:
			return v1.Pointer() == v2.Pointer()
		default:
			fmt.Printf("unsupported v1.Kind=%v t1.Name=%v, t1.Value=%v\n\n", v1.Kind(), v1.Type().Name(), v1.String())
			return false
		}
	}
	return true
}

func MetricDataIsEqual(d1, d2 *types.MetricData, compareTags bool) bool {
	v1 := reflect.ValueOf(*d1)
	v2 := reflect.ValueOf(*d2)

	for i := 0; i < v1.NumField(); i++ {
		if v1.Type().Field(i).Name == "Tags" && !compareTags {
			continue
		}
		r := deepCompareFields(v1.Field(i), v2.Field(i))
		if !r {
			return r
		}
	}
	return true
}

func DeepEqual(t *testing.T, target string, original, modified map[parser.MetricRequest][]*types.MetricData, compareTags bool) {
	for key := range original {
		if len(original[key]) == len(modified[key]) {
			for i := range original[key] {
				if !MetricDataIsEqual(original[key][i], modified[key][i], compareTags) {
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

func NearlyEqual(a, b []float64) bool {
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
	if len(a.Values) != len(b.Values) {
		return false
	}
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
	Target string
	M      map[parser.MetricRequest][]*types.MetricData
	W      []float64
	Name   string
	Step   int64
	Start  int64
	Stop   int64
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
		exp, _, _ := parser.ParseExpr(tt.Target)
		g, err := evaluator.Eval(context.Background(), exp, 0, 1, tt.M)
		if err != nil {
			t.Errorf("failed to eval %v: %+v", tt.Name, err)
			return
		}
		DeepEqual(t, g[0].Name, originalMetrics, tt.M, false)
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
		if _, ok := g[0].Tags["name"]; !ok {
			t.Errorf("metric with name %v doesn't contain 'name' tag", g[0].Name)
		}
	})
}

type MultiReturnEvalTestItem struct {
	Target  string
	M       map[parser.MetricRequest][]*types.MetricData
	Name    string
	Results map[string][]*types.MetricData
}

func TestMultiReturnEvalExpr(t *testing.T, tt *MultiReturnEvalTestItem) {
	evaluator := metadata.GetEvaluator()

	originalMetrics := DeepClone(tt.M)
	exp, _, err := parser.ParseExpr(tt.Target)
	if err != nil {
		t.Errorf("failed to parse %v: %+v", tt.Target, err)
		return
	}
	g, err := evaluator.Eval(context.Background(), exp, 0, 1, tt.M)
	if err != nil {
		t.Errorf("failed to eval %v: %+v", tt.Name, err)
		return
	}
	DeepEqual(t, tt.Name, originalMetrics, tt.M, true)
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
		t.Errorf("unexpected results len: got %d, want %d for %s", len(g), len(tt.Results), tt.Target)
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

func TestMultiReturnEvalExprModifiedOrigin(t *testing.T, tt *MultiReturnEvalTestItem) error {
	evaluator := metadata.GetEvaluator()

	exp, _, err := parser.ParseExpr(tt.Target)
	if err != nil {
		t.Errorf("failed to parse %v: %+v", tt.Target, err)
		return err
	}
	g, err := evaluator.Eval(context.Background(), exp, 0, 1, tt.M)
	if err != nil {
		t.Errorf("failed to eval %v: %+v", tt.Name, err)
		return err
	}
	if len(g) == 0 {
		t.Errorf("returned no data %v", tt.Name)
		return nil
	}
	if g[0] == nil {
		t.Errorf("returned no value %v", tt.Name)
		return nil
	}
	if g[0].StepTime == 0 {
		t.Errorf("missing Step for %+v", g)
	}
	if len(g) != len(tt.Results) {
		t.Errorf("unexpected results len: got %d, want %d for %s", len(g), len(tt.Results), tt.Target)
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

	return nil
}

type RewriteTestResult struct {
	Rewritten bool
	Targets   []string
	Err       error
}

type RewriteTestItem struct {
	//E    parser.Expr
	Target string
	M      map[parser.MetricRequest][]*types.MetricData
	Want   RewriteTestResult
}

func rewriteExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	if e.IsFunc() {
		metadata.FunctionMD.RLock()
		f, ok := metadata.FunctionMD.RewriteFunctions[e.Target()]
		metadata.FunctionMD.RUnlock()
		if ok {
			return f.Do(context.Background(), e, from, until, values)
		}
	}
	return false, nil, nil
}

func TestRewriteExpr(t *testing.T, tt *RewriteTestItem) {
	originalMetrics := DeepClone(tt.M)
	testName := tt.Target
	exp, _, err := parser.ParseExpr(tt.Target)
	if err != nil {
		t.Errorf("failed to parse %s: %+v", tt.Target, err)
		return
	}

	rewritten, targets, err := rewriteExpr(exp, 0, 1, tt.M)
	if err != tt.Want.Err {
		t.Errorf("unexpected error while calling rewrite for '%s': got '%+v', expected '%+v'", testName, err, tt.Want.Err)
		return
	}
	if rewritten != tt.Want.Rewritten {
		t.Errorf("unexpected result for rewritten for '%s': got '%v', expected '%v'", testName, rewritten, tt.Want.Rewritten)
		return
	}

	if len(targets) != len(tt.Want.Targets) {
		t.Errorf("%s returned a different number of metrics, actual %v, Want %v", testName, len(targets), len(tt.Want.Targets))
		return
	}
	DeepEqual(t, testName, originalMetrics, tt.M, false)

	for i, want := range tt.Want.Targets {
		if want != targets[i] {
			t.Errorf("unexpected result for rewrite for '%s': got='%s', expected='%s'", testName, targets[i], want)
		}
	}
}

type EvalTestItem struct {
	//E    parser.Expr
	Target string
	M      map[parser.MetricRequest][]*types.MetricData
	Want   []*types.MetricData
}

type EvalTestItemWithError struct {
	Target string
	M      map[parser.MetricRequest][]*types.MetricData
	Want   []*types.MetricData
	Error  error
}

type EvalTestItemWithRange struct {
	Target string
	M      map[parser.MetricRequest][]*types.MetricData
	Want   []*types.MetricData
	From   int64
	Until  int64
}

func (r *EvalTestItemWithRange) TestItem() *EvalTestItem {
	return &EvalTestItem{
		Target: r.Target,
		M:      r.M,
		Want:   r.Want,
	}
}

func TestEvalExprModifiedOrigin(t *testing.T, tt *EvalTestItem, from, until int64, strictOrder bool) error {
	evaluator := metadata.GetEvaluator()
	testName := tt.Target
	exp, _, err := parser.ParseExpr(tt.Target)
	if err != nil {
		t.Errorf("failed to parse %s: %+v", tt.Target, err)
		return nil
	}
	g, err := evaluator.Eval(context.Background(), exp, from, until, tt.M)
	if err != nil {
		return err
	}
	if len(g) != len(tt.Want) {
		t.Errorf("%s returned a different number of metrics, actual %v, Want %v", testName, len(g), len(tt.Want))
		return nil
	}

	for i, want := range tt.Want {
		actual := g[i]
		if _, ok := actual.Tags["name"]; !ok {
			t.Errorf("metric %+v with name %v doesn't contain 'name' tag", actual, actual.Name)
		}
		if actual == nil {
			t.Errorf("returned no value %v", tt.Target)
			return nil
		}
		if actual.StepTime == 0 {
			t.Errorf("missing Step for %+v", g)
		}
		if actual.Name != want.Name {
			t.Errorf("bad Name for %s metric %d: got %s, Want %s", testName, i, actual.Name, want.Name)
		}
		if !NearlyEqualMetrics(actual, want) {
			t.Errorf("different values for %s metric %s: got %v, Want %v", testName, actual.Name, actual.Values, want.Values)
			return nil
		}
		if actual.StepTime != want.StepTime {
			t.Errorf("different StepTime for %s metric %s: got %v, Want %v", testName, actual.Name, actual.StepTime, want.StepTime)
		}
		if actual.StartTime != want.StartTime {
			t.Errorf("different StartTime for %s metric %s: got %v, Want %v", testName, actual.Name, actual.StartTime, want.StartTime)
		}
		if actual.StopTime != want.StopTime {
			t.Errorf("different StopTime for %s metric %s: got %v, Want %v", testName, actual.Name, actual.StopTime, want.StopTime)
		}
	}
	return nil
}

func TestEvalExpr(t *testing.T, tt *EvalTestItem) {
	originalMetrics := DeepClone(tt.M)
	err := TestEvalExprModifiedOrigin(t, tt, 0, 1, false)
	if err != nil {
		t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
		return
	}
	DeepEqual(t, tt.Target, originalMetrics, tt.M, true)
}

func TestEvalExprWithRange(t *testing.T, tt *EvalTestItemWithRange) {
	originalMetrics := DeepClone(tt.M)
	tt2 := tt.TestItem()
	err := TestEvalExprModifiedOrigin(t, tt2, tt.From, tt.Until, false)
	if err != nil {
		t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
		return
	}
	DeepEqual(t, tt.Target, originalMetrics, tt.M, true)
}

func TestEvalExprWithRangeModifiedOrigin(t *testing.T, tt *EvalTestItemWithRange) {
	tt2 := tt.TestItem()
	err := TestEvalExprModifiedOrigin(t, tt2, tt.From, tt.Until, false)
	if err != nil {
		t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
		return
	}
}

func TestEvalExprWithError(t *testing.T, tt *EvalTestItemWithError) {
	originalMetrics := DeepClone(tt.M)
	tt2 := &EvalTestItem{
		Target: tt.Target,
		M:      tt.M,
		Want:   tt.Want,
	}
	err := TestEvalExprModifiedOrigin(t, tt2, 0, 1, false)
	if !merry.Is(err, tt.Error) {
		t.Errorf("unexpected error while evaluating %s: got `%+v`, expected `%+v`", tt.Target, err, tt.Error)
		return
	}
	DeepEqual(t, tt.Target, originalMetrics, tt.M, true)
}

func TestEvalExprOrdered(t *testing.T, tt *EvalTestItem) {
	originalMetrics := DeepClone(tt.M)
	err := TestEvalExprModifiedOrigin(t, tt, 0, 1, true)
	if err != nil {
		t.Errorf("unexpected error while evaluating %s: got `%+v`", tt.Target, err)
		return
	}
	DeepEqual(t, tt.Target, originalMetrics, tt.M, true)
}
