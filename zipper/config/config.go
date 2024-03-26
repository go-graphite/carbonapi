package config

import (
	"time"

	"github.com/barkimedes/go-deepcopy"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/zipper/types"
)

// Config is a structure that contains zipper-related configuration bits
type Config struct {
	SumBuckets                bool             `mapstructure:"sumBuckets"`
	Buckets                   int              `mapstructure:"buckets"`
	BucketsWidth              []int64          `mapstructure:"bucketsWidth"`
	BucketsLabels             []string         `mapstructure:"bucketsLabels"`
	ExtendedStat              bool             `mapstructure:"extendedStat"` // extended stat metrics
	SlowLogThreshold          time.Duration    `mapstructure:"slowLogThreshold"`
	ConcurrencyLimitPerServer int              `mapstructure:"concurrencyLimitPerServer"`
	MaxIdleConnsPerHost       int              `mapstructure:"maxIdleConnsPerHost"`
	Backends                  []string         `mapstructure:"backends"`
	BackendsV2                types.BackendsV2 `mapstructure:"backendsv2"`
	MaxBatchSize              *int             `mapstructure:"maxBatchSize"`
	FallbackMaxBatchSize      int              `mapstructure:"-"`
	MaxTries                  int              `mapstructure:"maxTries"`
	DoMultipleRequestsIfSplit bool             `mapstructure:"doMultipleRequestsIfSplit"`
	RequireSuccessAll         bool             `mapstructure:"requireSuccessAll"` // require full success for upstreams queries (for multi-target query)

	ExpireDelaySec       int32
	TLDCacheDisabled     bool `mapstructure:"tldCacheDisabled"`
	InternalRoutingCache time.Duration
	Timeouts             types.Timeouts
	KeepAliveInterval    time.Duration `mapstructure:"keepAliveInterval"`

	// ScaleToCommonStep controls if metrics in one target should be aggregated to common step
	ScaleToCommonStep bool `mapstructure:"scaleToCommonStep"`

	isSanitized bool
}

func (cfg *Config) IsSanitized() bool {
	return cfg.isSanitized
}

var defaultTimeouts = types.Timeouts{
	Render:  10000 * time.Second,
	Find:    100 * time.Second,
	Connect: 200 * time.Millisecond,
}

func sanitizeTimeouts(timeouts, defaultTimeouts types.Timeouts) types.Timeouts {
	if timeouts.Render == 0 {
		timeouts.Render = defaultTimeouts.Render
	}
	if timeouts.Find == 0 {
		timeouts.Find = defaultTimeouts.Find
	}

	if timeouts.Connect == 0 {
		timeouts.Connect = defaultTimeouts.Connect
	}

	return timeouts
}

// SanitizeConfig perform old kind of checks and conversions for zipper's configuration
func SanitizeConfig(logger *zap.Logger, oldConfig Config) *Config {
	// create a full copy of old config
	newConfigPtr, err := deepcopy.Anything(oldConfig)
	if err != nil {
		logger.Fatal("failed to copy old config", zap.Error(err))
	}
	newConfig := newConfigPtr.(Config)

	if len(newConfig.BucketsWidth) > 0 {
		newConfig.Buckets = 0
	}

	if newConfig.MaxBatchSize == nil {
		newConfig.MaxBatchSize = &newConfig.FallbackMaxBatchSize
	}

	newConfig.Timeouts = sanitizeTimeouts(newConfig.Timeouts, defaultTimeouts)

	if newConfig.InternalRoutingCache.Seconds() < 30 {
		logger.Warn("internalRoutingCache is too low",
			zap.String("reason", "this variable is used for internal routing cache, minimum allowed is 30s"),
			zap.String("recommendation", "it's usually good idea to set it to something like 600s"),
		)
		newConfig.InternalRoutingCache = 30 * time.Second
	}

	// Convert old config format to new one
	defaultIdleConnTimeout := 3600 * time.Second
	if newConfig.Backends != nil && len(newConfig.Backends) != 0 {
		newConfig.BackendsV2 = types.BackendsV2{
			Backends: []types.BackendV2{
				{
					GroupName:                 "backends",
					Protocol:                  "carbonapi_v2_pb",
					LBMethod:                  "broadcast",
					Servers:                   newConfig.Backends,
					Timeouts:                  &newConfig.Timeouts,
					ConcurrencyLimit:          &newConfig.ConcurrencyLimitPerServer,
					DoMultipleRequestsIfSplit: true,
					KeepAliveInterval:         &newConfig.KeepAliveInterval,
					MaxIdleConnsPerHost:       &newConfig.MaxIdleConnsPerHost,
					MaxTries:                  &newConfig.MaxTries,
					MaxBatchSize:              newConfig.MaxBatchSize,
					IdleConnectionTimeout:     &defaultIdleConnTimeout,
				},
			},
			MaxIdleConnsPerHost:       newConfig.MaxIdleConnsPerHost,
			ConcurrencyLimitPerServer: newConfig.ConcurrencyLimitPerServer,
			Timeouts:                  newConfig.Timeouts,
			KeepAliveInterval:         newConfig.KeepAliveInterval,
			MaxTries:                  newConfig.MaxTries,
			MaxBatchSize:              newConfig.MaxBatchSize,
		}

		newConfig.DoMultipleRequestsIfSplit = true
	}

	newConfig.BackendsV2.Timeouts = sanitizeTimeouts(newConfig.BackendsV2.Timeouts, newConfig.Timeouts)
	for i := range newConfig.BackendsV2.Backends {
		if newConfig.BackendsV2.Backends[i].Timeouts == nil {
			timeouts := newConfig.BackendsV2.Timeouts
			newConfig.BackendsV2.Backends[i].Timeouts = &timeouts
		}
		timeouts := sanitizeTimeouts(*(newConfig.BackendsV2.Backends[i].Timeouts), newConfig.BackendsV2.Timeouts)
		newConfig.BackendsV2.Backends[i].Timeouts = &timeouts
		if newConfig.BackendsV2.Backends[i].IdleConnectionTimeout == nil {
			newConfig.BackendsV2.Backends[i].IdleConnectionTimeout = &defaultIdleConnTimeout
		}
	}

	if newConfig.BackendsV2.MaxBatchSize == nil {
		newConfig.BackendsV2.MaxBatchSize = newConfig.MaxBatchSize
	}

	newConfig.isSanitized = true
	return &newConfig
}
