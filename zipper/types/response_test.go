package types

import (
	"math"
	"testing"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

func TestMergeFetchResponsesWithNanAppending(t *testing.T) {
	m1 := protov3.FetchResponse{
		Values: []float64{math.NaN(), 0, 0, 0, math.NaN()},
	}

	m2 := protov3.FetchResponse{
		Values: []float64{0},
	}

	exp := protov3.FetchResponse{
		Values: []float64{0, 0, 0, 0, math.NaN()},
	}

	err := MergeFetchResponses(&m1, &m2)
	if err != nil {
		t.Error(err)
		return
	}

	if !cmpFloat64Arrays(m1.Values, exp.Values, 0.00001) {
		t.Errorf("Error merging responses\nExp: %v\nGot: %v", exp, m1)
	}
}

func cmpFloat64Arrays(a, b []float64, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		} else if math.IsInf(a[i], 1) && math.IsInf(b[i], 1) {
			continue
		} else if math.IsInf(a[i], -1) && math.IsInf(b[i], -1) {
			continue
		}

		if math.Abs(a[i]-b[i]) >= epsilon {
			return false
		}
	}

	return true
}
