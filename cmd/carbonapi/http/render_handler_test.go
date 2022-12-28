package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	"github.com/lomik/zapwriter"
)

func BenchmarkResponseCacheComputeKey(b *testing.B) {
	var from int64 = 1628876560
	var until int64 = 1628876620
	var maxDataPoints int64 = 1024
	noNullPoints := false
	template := "test"
	targets := []string{
		"test.metric.*.cpu.load_avg",
		"test.metric.*.memory.free",
	}
	format := "json"

	for i := 0; i < b.N; i++ {
		_ = responseCacheComputeKey(from, until, targets, format, maxDataPoints, noNullPoints, template)
	}
}

func BenchmarkBackendCacheComputeKey(b *testing.B) {
	from := "1628876560"
	until := "1628876620"
	targets := []string{
		"test.metric.*.cpu.load_avg",
		"test.metric.*.memory.free",
	}

	for i := 0; i < b.N; i++ {
		_ = backendCacheComputeKey(from, until, targets, 0, true)
	}
}

func BenchmarkBackendCacheComputeKeyAbs(b *testing.B) {
	var from int64 = 1628876560
	var until int64 = 1628876620
	targets := []string{
		"test.metric.*.cpu.load_avg",
		"test.metric.*.memory.free",
	}

	for i := 0; i < b.N; i++ {
		_ = backendCacheComputeKeyAbs(from, until, targets, 0, true)
	}
}

func Test_getCacheTimeout(t *testing.T) {
	cacheConfig := config.CacheConfig{
		ShortTimeoutSec:     60,
		DefaultTimeoutSec:   300,
		ShortDuration:       3 * time.Hour,
		ShortUntilOffsetSec: 120,
	}

	now := int64(1636985018)

	tests := []struct {
		name  string
		now   time.Time
		from  int64
		until int64
		want  int32
	}{
		{
			name:  "short: from = now - 600, until = now - 120",
			now:   time.Unix(now, 0),
			from:  now - 600,
			until: now - 120,
			want:  60,
		},
		{
			name:  "short: from = now - 10800",
			now:   time.Unix(now, 0),
			from:  now - 10800,
			until: now,
			want:  60,
		},
		{
			name:  "short: from = now - 10810, until = now - 120",
			now:   time.Unix(now, 0),
			from:  now - 10800,
			until: now - 120,
			want:  60,
		},
		{
			name:  "short: from = now - 10800, until now - 121",
			now:   time.Unix(now, 0),
			from:  now - 10800,
			until: now - 121,
			want:  300,
		},
		{
			name:  "default: from = now - 10801",
			now:   time.Unix(now, 0),
			from:  now - 10801,
			until: now,
			want:  300,
		},
		{
			name:  "short: from = now - 122, until = now - 121",
			now:   time.Unix(now, 0),
			from:  now - 122,
			until: now - 121,
			want:  300,
		},
	}
	logger := zapwriter.Logger("test")
	r, _ := http.NewRequest("GET", "http://127.0.0.1/render", nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := time.Second * time.Duration(tt.until-tt.from)
			if got := getCacheTimeout(logger, r, tt.now.Unix(), tt.until, duration, &cacheConfig); got != tt.want {
				t.Errorf("getCacheTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
