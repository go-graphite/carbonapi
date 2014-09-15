package main

import (
	"bytes"
	"encoding/json"
	"math"
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

		{"1s", -1, -1},
		{"+2d", 2 * 60 * 60 * 24, -1},
		{"-10hours", -60 * 60 * 10, -1},
	}

	for _, tt := range tests {
		if secs, _ := intervalString(tt.t, tt.sign); secs != tt.seconds {
			t.Errorf("intervalString(%q)=%d, want %d\n", tt.t, secs, tt.seconds)
		}
	}
}

func TestJSONResponse(t *testing.T) {

	tests := []struct {
		jsr jsonResponse
		out []byte
	}{
		{
			jsonResponse{
				Target:     "jsTarget",
				Datapoints: []graphitePoint{{1, 100}, {1.5, 200}, {2.25, 300}, {math.NaN(), 400}},
			},
			[]byte(`{"target":"jsTarget","datapoints":[[1,100],[1.5,200],[2.25,300],[null,400]]}`),
		},
	}

	for _, tt := range tests {
		b, err := json.Marshal(tt.jsr)
		if err != nil {
			t.Errorf("error marshalling %+v: %+v", tt.jsr, err)
			continue
		}
		if !bytes.Equal(b, tt.out) {
			t.Errorf("json.Marshal(%+v)=%+v, want %+v", tt.jsr, string(b), string(tt.out))
		}
	}
}

func TestRawResponse(t *testing.T) {

	r := makeResponse("metric1", []float64{1, 2, math.NaN(), 8, 16, 32}, 60, 1410633660)

	b := marshalRaw(r)

	want := []byte(`metric1,1410633660,1410634020,60|1,2,None,8,16,32` + "\n")

	if !bytes.Equal(b, want) {
		t.Errorf("marshalRaw(...)=%q, want %q", string(b), string(want))
	}
}
