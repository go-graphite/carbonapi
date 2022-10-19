package http

import (
	"testing"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	zipperCfg "github.com/go-graphite/carbonapi/zipper/config"
	"github.com/lomik/zapwriter"
	"github.com/stretchr/testify/assert"
)

func Test_initRequestsHistogram(t *testing.T) {
	logger := zapwriter.Logger("test")
	tests := []struct {
		name       string
		config     zipperCfg.Config
		wantLabels []string
	}{
		{
			name:   "fixed",
			config: zipperCfg.Config{Buckets: 10},
			wantLabels: []string{
				"_in_0ms_to_100ms", "_in_100ms_to_200ms", "_in_200ms_to_300ms", "_in_300ms_to_400ms", "_in_400ms_to_500ms",
				"_in_500ms_to_600ms", "_in_600ms_to_700ms", "_in_700ms_to_800ms", "_in_800ms_to_900ms", "_in_900ms_to_1000ms",
				"_in_1000ms_to_1100ms",
			},
		},
		{
			name:   "fixed with sum",
			config: zipperCfg.Config{Buckets: 10, SumBuckets: true},
			wantLabels: []string{
				"_to_100ms", "_to_200ms", "_to_300ms", "_to_400ms", "_to_500ms",
				"_to_600ms", "_to_700ms", "_to_800ms", "_to_900ms", "_to_1000ms", "_to_1100ms",
			},
		},
		{
			name:   "variable buckets",
			config: zipperCfg.Config{BucketsWidth: []int64{100, 500, 1000, 5000, 10000}},
			wantLabels: []string{
				"_in_0ms_to_100ms", "_in_100ms_to_500ms", "_in_500ms_to_1000ms", "_in_1000ms_to_5000ms",
				"_in_5000ms_to_10000ms", "_in_10000ms_to_inf",
			},
		},
		{
			name:   "variable buckets with sum",
			config: zipperCfg.Config{BucketsWidth: []int64{100, 500, 1000, 5000, 10000}, SumBuckets: true},
			wantLabels: []string{
				"_to_100ms", "_to_500ms", "_to_1000ms", "_to_5000ms", "_to_10000ms", "_to_inf",
			},
		},
		{
			name: "variable buckets with partial labels",
			config: zipperCfg.Config{
				BucketsWidth:  []int64{100, 500, 1000, 5000, 10000},
				BucketsLabels: []string{"low", "", "", "", "high", "inf", "none"},
			},
			wantLabels: []string{
				"low", "_in_100ms_to_500ms", "_in_500ms_to_1000ms", "_in_1000ms_to_5000ms", "high", "inf",
			},
		},
		{
			name: "variable buckets with sum and partial labels",
			config: zipperCfg.Config{
				BucketsWidth:  []int64{100, 500, 1000, 5000, 10000},
				BucketsLabels: []string{"low", "", "", "", "high", "inf", "none"},
				SumBuckets:    true,
			},
			wantLabels: []string{
				"low", "_to_500ms", "_to_1000ms", "_to_5000ms", "high", "inf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Config.Upstreams = *zipperCfg.SanitizeConfig(logger, tt.config)
			got := initRequestsHistogram()
			assert.Equal(t, tt.wantLabels, got.Labels())
			assert.Equal(t, len(tt.wantLabels), len(got.Values()))
		})
	}
}
