package compare

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/go-graphite/carbonapi/expr/types"
)

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

func MaxInt(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func TestMetricData(t *testing.T, got, want []*types.MetricData) {
	for i := 0; i < MaxInt(len(want), len(got)); i++ {
		if i >= len(got) {
			t.Errorf("\n-[%d] = %v", i, want[i])
		} else if i >= len(want) {
			t.Errorf("\n+[%d] = %v", i, got[i])
		} else {
			actual := got[i]
			if _, ok := actual.Tags["name"]; !ok {
				t.Errorf("metric %+v with name %v doesn't contain 'name' tag", actual, actual.Name)
			}
			if actual == nil {
				t.Errorf("returned no value")
				return
			}
			if actual.StepTime == 0 {
				t.Errorf("missing Step for %+v", actual)
			}
			if actual.Name != want[i].Name {
				t.Errorf("bad Name metric[%d]: got %s, want %s", i, actual.Name, want[i].Name)
			}
			if !NearlyEqualMetrics(actual, want[i]) {
				t.Errorf("different values metric[%d] %s: got %v, want %v", i, actual.Name, actual.Values, want[i].Values)
			}
			if actual.StepTime != want[i].StepTime {
				t.Errorf("different StepTime metric[%d] %s: got %v, want %v", i, actual.Name, actual.StepTime, want[i].StepTime)
			}
			if actual.StartTime != want[i].StartTime {
				t.Errorf("different StartTime metric[%d] %s: got %v, want %v", i, actual.Name, actual.StartTime, want[i].StartTime)
			}
			if actual.StopTime != want[i].StopTime {
				t.Errorf("different StopTime metric[%d] %s: got %v, want %v", i, actual.Name, actual.StopTime, want[i].StopTime)
			}
		}
	}
}
