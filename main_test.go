package main

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestInterval(t *testing.T) {

	var tests = []struct {
		t       string
		seconds int32
		sign    int
	}{
		{"1s", 1, 1},
		{"2d", 2 * 60 * 60 * 24, 1},
		{"10hours", 60 * 60 * 10, 1},
		{"7d13h45min21s", 7*24*60*60 + 13*60*60 + 45*60 + 21, 1},
		{"01hours", 60 * 60 * 1, 1},
		{"2d2d", 4 * 60 * 60 * 24, 1},

		{"1s", -1, -1},
		{"+2d", 2 * 60 * 60 * 24, -1},
		{"-10hours", -60 * 60 * 10, -1},
		{"-360h2min", -360*60*60 - 2*60, -1},
	}

	for _, tt := range tests {
		if secs, _ := intervalString(tt.t, tt.sign); secs != tt.seconds {
			t.Errorf("intervalString(%q)=%d, want %d\n", tt.t, secs, tt.seconds)
		}
	}

	var exceptTests = []struct {
		t       string
		seconds int32
		err     string
		sign    int
	}{
		{"10m10s", 0, "unknown time units", 1},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000y", 0, "value out of range", 1},
	}
	for _, tt := range exceptTests {
		secs, err := intervalString(tt.t, tt.sign)
		if secs != tt.seconds {
			t.Errorf("intervalString(%q)=%d, want %d\n", tt.t, secs, tt.seconds)
		}
		if !strings.Contains(err.Error(), tt.err) {
			t.Errorf("Error of intervalString(%q)=%v, expected to contain %v\n", tt.t, err.Error(), tt.err)
		}
	}
}

func TestJSONResponse(t *testing.T) {

	tests := []struct {
		results []*metricData
		out     []byte
	}{
		{
			[]*metricData{
				makeResponse("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				makeResponse("metric2", []float64{2, 2.5, 3.25, 4, 5}, 100, 100),
			},
			[]byte(`[{"target":"metric1","datapoints":[[1,100],[1.5,200],[2.25,300],[null,400]]},{"target":"metric2","datapoints":[[2,100],[2.5,200],[3.25,300],[4,400],[5,500]]}]`),
		},
	}

	for _, tt := range tests {
		b := marshalJSON(tt.results)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalJSON(%+v)=%+v, want %+v", tt.results, string(b), string(tt.out))
		}
	}
}

func TestRawResponse(t *testing.T) {

	tests := []struct {
		results []*metricData
		out     []byte
	}{
		{
			[]*metricData{
				makeResponse("metric1", []float64{1, 1.5, 2.25, math.NaN()}, 100, 100),
				makeResponse("metric2", []float64{2, 2.5, 3.25, 4, 5}, 100, 100),
			},
			[]byte(`metric1,100,500,100|1,1.5,2.25,None` + "\n" + `metric2,100,600,100|2,2.5,3.25,4,5` + "\n"),
		},
	}

	for _, tt := range tests {
		b := marshalRaw(tt.results)
		if !bytes.Equal(b, tt.out) {
			t.Errorf("marshalRaw(%+v)=%+v, want %+v", tt.results, string(b), string(tt.out))
		}
	}
}

func getData(rangeSize int) []float64 {
	var data = make([]float64, rangeSize)
	var r = rand.New(rand.NewSource(99))
	for i := range data {
		data[i] = math.Floor(1000 * r.Float64())
	}

	return data
}

func BenchmarkMarshalJSON(b *testing.B) {
	data := []*metricData{
		makeResponse("metric1", getData(10000), 100, 100),
		makeResponse("metric2", getData(100000), 100, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = marshalJSON(data)
	}
}
