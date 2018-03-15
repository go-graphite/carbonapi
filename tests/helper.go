package tests

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
	"reflect"
)

type FuncEvaluator struct {
	eval func(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

func (evaluator *FuncEvaluator) EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return evaluator.eval(e, from, until, values)
}

func EvaluatorFromFunc(function interfaces.Function) interfaces.Evaluator {
	e := &FuncEvaluator{
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
					IsAbsent:  make([]bool, len(originalMetric.IsAbsent)),
				},
			}

			copy(copiedMetric.Values, originalMetric.Values)
			copy(copiedMetric.IsAbsent, originalMetric.IsAbsent)
			copiedMetrics = append(copiedMetrics, &copiedMetric)
		}

		clone[key] = copiedMetrics
	}

	return clone
}

func DeepEqual(t *testing.T, target string, original, modified map[parser.MetricRequest][]*types.MetricData) {
	for key := range original {
		if len(original[key]) == len(modified[key]) {
			for i := range original[key] {
				if !reflect.DeepEqual(original[key][i], modified[key][i]) {
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

func NearlyEqual(a []float64, absent []bool, b []float64) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		// "same"
		if absent[i] && math.IsNaN(b[i]) {
			continue
		}
		if absent[i] || math.IsNaN(b[i]) {
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

	if len(a.IsAbsent) != len(b.IsAbsent) {
		return false
	}

	for i := range a.IsAbsent {
		if a.IsAbsent[i] != b.IsAbsent[i] {
			return false
		}
		// "close enough"
		if math.Abs(a.Values[i]-b.Values[i]) > eps {
			return false
		}
	}

	return true
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
			t.Errorf("missing step for %+v", g)
		}
		if actual.Name != want.Name {
			t.Errorf("bad name for %s metric %d: got %s, Want %s", testName, i, actual.Name, want.Name)
		}
		if !NearlyEqualMetrics(actual, want) {
			t.Errorf("different values for %s metric %s: got %v, Want %v", testName, actual.Name, actual.Values, want.Values)
			return
		}
	}
}
