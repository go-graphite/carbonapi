package main

import "testing"

func TestInterval(t *testing.T) {

	var tests = []struct {
		t       string
		seconds int32
	}{
		{"1s", 1},
		{"2d", 2 * 60 * 60 * 24},
		{"10hours", 60 * 60 * 10},
	}

	for _, tt := range tests {
		if secs, _ := intervalString(tt.t); secs != tt.seconds {
			t.Errorf("intervalString(%q)=%d, want %d\n", tt.t, secs, tt.seconds)
		}
	}
}
