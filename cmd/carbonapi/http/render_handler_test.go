package http

import (
	"testing"
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
		_ = backendCacheComputeKey(from, until, targets)
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
		_ = backendCacheComputeKeyAbs(from, until, targets)
	}
}
