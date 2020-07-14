package helper

import (
	"testing"

	"github.com/go-graphite/carbonapi/expr/tags"
)

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		expected map[string]string
	}{
		{
			name:   "tagged metric",
			metric: "cpu.usage_idle;cpu=cpu-total;host=test",
			expected: map[string]string{
				"name": "cpu.usage_idle",
				"cpu":  "cpu-total",
				"host": "test",
			},
		},
		{
			name:   "no tags in metric",
			metric: "cpu.usage_idle",
			expected: map[string]string{
				"name": "cpu.usage_idle",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tags.ExtractTags(tt.metric)
			if len(actual) != len(tt.expected) {
				t.Fatalf("amount of tags doesn't match: got %v, expected %v", actual, tt.expected)
			}
			for tag, value := range actual {
				vExpected, ok := tt.expected[tag]
				if !ok {
					t.Fatalf("tag %v not found in %+v", value, actual)
				} else if vExpected != value {
					t.Errorf("unexpected tag-value, got %v, expected %v", value, vExpected)
				}
			}
		})
	}
}

func TestGCD(t *testing.T) {
	tests := []struct {
		arg1     int64
		arg2     int64
		expected int64
	}{
		{
			13,
			17,
			1,
		},
		{
			14,
			21,
			7,
		},
		{
			12,
			16,
			4,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("GDC(%v, %v)=>%v", tt.arg1, tt.arg2, tt.expected), func(t *testing.T) {
			value := GCD(tt.arg1, tt.arg2)
			if value != tt.expected {
				t.Errorf("GCD of %v and %v != %v: %v", tt.arg1, tt.arg2, tt.expected, value)
			}
		})
	}
}

func TestLCM(t *testing.T) {
	tests := []struct {
		args     []int64
		expected int64
	}{
		{
			[]int64{2, 3},
			6,
		},
		{
			[]int64{},
			0,
		},
		{
			[]int64{15},
			15,
		},
		{
			[]int64{10, 15, 20},
			60,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("LMC(%v)=>%v", tt.args, tt.expected), func(t *testing.T) {
			value := LCM(tt.args...)
			if value != tt.expected {
				t.Errorf("LCM of %v != %v: %v", tt.args, tt.expected, value)
			}
		})
	}
}
