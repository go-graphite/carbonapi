package zipper

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type mergeValuesData struct {
	name           string
	m1             protov3.FetchResponse
	m2             protov3.FetchResponse
	expectedResult protov3.FetchResponse
	expectedError  errors.Errors
}

var (
	errMetadataMismatchFmt = "%v mismatch, got %v, expected %v"
	errLengthMismatchFmt   = "length mismatch, got %v, expected %v"
	errContentMismatchFmt  = "content mismatch at pos %v, got %v, expected %v"
)

func fetchResponseEquals(r1, r2 *protov3.FetchResponse) error {
	if r1.StartTime != r2.StartTime {
		return fmt.Errorf(errMetadataMismatchFmt, "StartTime", r1.StartTime, r2.StartTime)
	}

	if r1.StopTime != r2.StopTime {
		return fmt.Errorf(errMetadataMismatchFmt, "StopTime", r1.StopTime, r2.StopTime)
	}

	if r1.XFilesFactor != r2.XFilesFactor {
		return fmt.Errorf(errMetadataMismatchFmt, "XFilesFactor", r1.XFilesFactor, r2.XFilesFactor)
	}

	if r1.Name != r2.Name {
		return fmt.Errorf(errMetadataMismatchFmt, "Name", r1.Name, r2.Name)
	}

	if r1.StepTime != r2.StepTime {
		return fmt.Errorf(errMetadataMismatchFmt, "StepTime", r1.StepTime, r2.StepTime)
	}

	if r1.ConsolidationFunc != r2.ConsolidationFunc {
		return fmt.Errorf(errMetadataMismatchFmt, "ConsolidationFunc", r1.ConsolidationFunc, r2.ConsolidationFunc)
	}

	if len(r1.Values) != len(r2.Values) {
		return fmt.Errorf(errLengthMismatchFmt, r1.Values, r2.Values)
	}

	for i := range r1.Values {
		if math.IsNaN(r1.Values[i]) && math.IsNaN(r2.Values[i]) {
			continue
		}
		if r1.Values[i] != r2.Values[i] {
			return fmt.Errorf(errContentMismatchFmt, i, r1.Values[i], r2.Values[i])
		}
	}

	return nil
}

func TestMergeValues(t *testing.T) {
	tests := []mergeValuesData{
		{
			name: "simple 1",
			// 60 seconds
			m1: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			},
			// 120 seconds
			m2: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         0,
				StepTime:          120,
				ConsolidationFunc: "average",
				Values:            []float64{1, 3, 5, 7, 9},
			},

			expectedResult: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			},

			expectedError: errors.Errors{},
		},
		{
			name: "simple 2",
			// 60 seconds
			m1: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         0,
				StepTime:          120,
				ConsolidationFunc: "average",
				Values:            []float64{1, 3, 5, 7, 9},
			},
			// 120 seconds
			m2: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			},

			expectedResult: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			},

			expectedError: errors.Errors{},
		},
		{
			name: "fill the gaps simple",
			// 60 seconds
			m1: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, math.NaN(), 6, 7, 8, 9, math.NaN(), 11, 12, 13, 14, 15, 16, math.NaN(), math.NaN(), math.NaN(), 20},
			},
			// 120 seconds
			m2: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, math.NaN(), math.NaN(), 5, 6, 7, 8, 9, math.NaN(), 11, 12, math.NaN(), 14, 15, 16, 17, 18, math.NaN(), 20},
			},

			expectedResult: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, math.NaN(), 11, 12, 13, 14, 15, 16, 17, 18, math.NaN(), 20},
			},

			expectedError: errors.Errors{},
		},
		{
			name: "fill the gaps 1",
			// 60 seconds
			m1: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         0,
				StepTime:          120,
				ConsolidationFunc: "average",
				Values:            []float64{1, math.NaN(), 5, 7, 9, 11, 13, 15, 17, 19},
			},
			// 120 seconds
			m2: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, math.NaN(), math.NaN(), 5, 6, 7, 8, 9, math.NaN(), 11, 12, math.NaN(), 14, 15, 16, 17, 18, math.NaN(), 20},
			},

			expectedResult: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, math.NaN(), math.NaN(), 5, 6, 7, 8, 9, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},

			expectedError: errors.Errors{},
		},
		{
			name: "fill end of metric",
			// 60 seconds
			m1: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, math.NaN(), 6, 7, 8, 9, math.NaN(), 11, 12, 13, 14, 15, 16, math.NaN(), math.NaN(), math.NaN()},
			},
			// 120 seconds
			m2: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, math.NaN(), math.NaN(), 5, 6, 7, 8, 9, math.NaN(), 11, 12, math.NaN(), 14, 15, 16, 17, 18, math.NaN(), 20},
			},

			expectedResult: protov3.FetchResponse{
				Name:              "foo",
				StartTime:         60,
				StepTime:          60,
				ConsolidationFunc: "average",
				Values:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, math.NaN(), 11, 12, 13, 14, 15, 16, 17, 18, math.NaN(), 20},
			},

			expectedError: errors.Errors{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.m1.StopTime <= test.m1.StartTime {
				test.m1.StopTime = test.m1.StartTime + int64(len(test.m1.Values))*test.m1.StepTime
			}
			if test.m2.StopTime <= test.m2.StartTime {
				test.m2.StopTime = test.m2.StartTime + int64(len(test.m2.Values))*test.m2.StepTime
			}
			if test.expectedResult.StopTime <= test.expectedResult.StartTime {
				test.expectedResult.StopTime = test.expectedResult.StartTime + int64(len(test.expectedResult.Values))*test.expectedResult.StepTime
			}
			err := types.MergeFetchResponses(&test.m1, &test.m2, "test")
			if err == nil {
				err = &errors.Errors{}
			}
			if len(err.Errors) != len(test.expectedError.Errors) && err.HaveFatalErrors != test.expectedError.HaveFatalErrors {
				t.Fatalf("unexpected error: '%v'", err)
			}

			err2 := fetchResponseEquals(&test.m1, &test.expectedResult)
			if err2 != nil {
				t.Fatalf("unexpected difference: '%v'\n    got     : %+v\n    expected: %+v\n", err, test.m1, test.expectedResult)
			}
		})
	}
}
