package main

import (
	"strconv"
	"testing"
)

func Test_resortErr(t *testing.T) {
	tests := []struct {
		errStr string
		want   string
	}{
		{
			errStr: "b: connection refused\na: connection refused\n",
			want:   "a: connection refused\nb: connection refused\n",
		},
		{
			errStr: "a: connection refused\nb: connection refused\n",
			want:   "a: connection refused\nb: connection refused\n",
		},
		{
			errStr: "",
			want:   "",
		},
		{
			errStr: "\n",
			want:   "\n",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := resortErr(tt.errStr); got != tt.want {
				t.Errorf("resortErr(%q) = %q, want %q", tt.errStr, got, tt.want)
			}
		})
	}
}
