package http

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
)

func Test_timestampTruncate(t *testing.T) {
	// reverse sorted
	durations := []config.DurationTruncate{
		{Duration: 30 * 24 * time.Hour, Truncate: 10 * time.Minute},
		{Duration: time.Hour, Truncate: time.Minute},
		{Duration: 10 * time.Minute, Truncate: 10 * time.Second},
		{Duration: 0, Truncate: 2 * time.Second},
	}

	tests := []struct {
		ts        int64
		duration  time.Duration
		durations []config.DurationTruncate
		want      int64
	}{
		{
			ts:        1628876563,
			duration:  5 * time.Minute,
			durations: durations,
			want:      1628876562, // truncation to 2s
		},
		{
			ts:        1628876563,
			duration:  10 * time.Minute,
			durations: durations,
			want:      1628876562, // truncate to 2s
		},
		{
			ts:        1628876563,
			duration:  10*time.Minute + time.Second,
			durations: durations,
			want:      1628876560, // truncate to 10s
		},
		{
			ts:        1628876563,
			duration:  2 * time.Hour,
			durations: durations,
			want:      1628876520, // truncate to 1m
		},
		{
			ts:        1628876563,
			duration:  30 * 24 * time.Hour,
			durations: durations,
			want:      1628876520, // truncate to 1m
		},
		{
			ts:        1628876563,
			duration:  30*24*time.Hour + time.Second,
			durations: durations,
			want:      1628876400, // truncate to 10m
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d -> %d", tt.ts, tt.want), func(t *testing.T) {
			if got := timestampTruncate(tt.ts, tt.duration, tt.durations); got != tt.want {
				t.Errorf("timestampTruncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_queryLengthLimitExceeded(t *testing.T) {
	tests := []struct {
		query     []string
		maxLength uint64
		want      bool
	}{
		{
			query:     []string{"a.b.c.d.e", "a.a.a.a.a.a.a.a.a.a.a.a"},
			maxLength: 20,
			want:      true,
		},
		{
			query:     []string{"a.b.c", "a.b.d"},
			maxLength: 10,
			want:      false,
		},
		{
			query:     []string{"a.b.c", "a.b.d"},
			maxLength: 9,
			want:      true,
		},
		{
			query:     []string{"a.b.c.d.e", "a.a.a.a.a.a.a.a.a.a.a.a"},
			maxLength: 0,
			want:      false,
		},
		{
			query:     []string{"a.b.b.c.*", "a.d.e", "a.b.c", "a.b.e", "a.a.a.b", "a.f.s.x.w.g"},
			maxLength: 30,
			want:      true,
		},
		{
			query:     []string{"a.b.c.d.e", "a.b.c.d.f", "b.a.*"},
			maxLength: 10000,
			want:      false,
		},
		{
			query:     []string{},
			maxLength: 0,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("queryLengthLimitExceeded([%s], %d) -> %t", strings.Join(tt.query, ", "), tt.maxLength, tt.want), func(t *testing.T) {
			if got := queryLengthLimitExceeded(tt.query, tt.maxLength); got != tt.want {
				t.Errorf("queryLengthLimitExceeded() = %t, want %t", got, tt.want)
			}
		})
	}
}
