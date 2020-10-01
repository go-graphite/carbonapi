package config

import (
	"time"

	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/zipper/types"
)

// Config is a structure that contains zipper-related configuration bits
type Config struct {
	ConcurrencyLimitPerServer int              `mapstructure:"concurrencyLimitPerServer"`
	MaxIdleConnsPerHost       int              `mapstructure:"maxIdleConnsPerHost"`
	Backends                  []string         `mapstructure:"backends"`
	BackendsV2                types.BackendsV2 `mapstructure:"backendsv2"`
	MaxBatchSize              *int             `mapstructure:"maxBatchSize"`
	FallbackMaxBatchSize      int              `mapstructure:"-"`
	MaxTries                  int              `mapstructure:"maxTries"`
	DoMultipleRequestsIfSplit bool             `mapstructure:"doMultipleRequestsIfSplit"`

	CarbonSearch   types.CarbonSearch
	CarbonSearchV2 types.CarbonSearchV2

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
	newConfig := &Config{
		ConcurrencyLimitPerServer: oldConfig.ConcurrencyLimitPerServer,
		MaxIdleConnsPerHost:       oldConfig.MaxIdleConnsPerHost,
		Backends:                  oldConfig.Backends,
		BackendsV2:                oldConfig.BackendsV2,
		MaxBatchSize:              oldConfig.MaxBatchSize,
		FallbackMaxBatchSize:      oldConfig.FallbackMaxBatchSize,
		MaxTries:                  oldConfig.MaxTries,

		CarbonSearch:   oldConfig.CarbonSearch,
		CarbonSearchV2: oldConfig.CarbonSearchV2,

		ExpireDelaySec:       oldConfig.ExpireDelaySec,
		TLDCacheDisabled:     oldConfig.TLDCacheDisabled,
		InternalRoutingCache: oldConfig.InternalRoutingCache,
		Timeouts:             oldConfig.Timeouts,
		KeepAliveInterval:    oldConfig.KeepAliveInterval,
		ScaleToCommonStep:    oldConfig.ScaleToCommonStep,
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
	if newConfig.CarbonSearch.Backend != "" {
		newConfig.CarbonSearchV2.BackendsV2 = types.BackendsV2{
			Backends: []types.BackendV2{{
				GroupName:                 newConfig.CarbonSearch.Backend,
				Protocol:                  "carbonapi_v2_pb",
				LBMethod:                  "roundrobin",
				Servers:                   []string{newConfig.CarbonSearch.Backend},
				Timeouts:                  &newConfig.Timeouts,
				DoMultipleRequestsIfSplit: true,
				ConcurrencyLimit:          &newConfig.ConcurrencyLimitPerServer,
				KeepAliveInterval:         &newConfig.KeepAliveInterval,
				MaxIdleConnsPerHost:       &newConfig.MaxIdleConnsPerHost,
				MaxTries:                  &newConfig.MaxTries,
			}},
			MaxIdleConnsPerHost:       newConfig.MaxIdleConnsPerHost,
			ConcurrencyLimitPerServer: newConfig.ConcurrencyLimitPerServer,
			Timeouts:                  newConfig.Timeouts,
			KeepAliveInterval:         newConfig.KeepAliveInterval,
			MaxTries:                  newConfig.MaxTries,
		}

		newConfig.CarbonSearchV2.Prefix = newConfig.CarbonSearch.Prefix
	}

	// Convert old config format to new one
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
	}

	if newConfig.BackendsV2.MaxBatchSize == nil {
		newConfig.BackendsV2.MaxBatchSize = newConfig.MaxBatchSize
	}

	for i := range newConfig.CarbonSearchV2.Backends {
		if newConfig.CarbonSearchV2.Backends[i].MaxBatchSize == nil {
			newConfig.CarbonSearchV2.Backends[i].MaxBatchSize = newConfig.CarbonSearchV2.MaxBatchSize
		}
	}

	if newConfig.CarbonSearchV2.MaxBatchSize == nil {
		newConfig.CarbonSearchV2.MaxBatchSize = newConfig.MaxBatchSize
	}

	for i := range newConfig.CarbonSearchV2.Backends {
		if newConfig.CarbonSearchV2.Backends[i].MaxBatchSize == nil {
			newConfig.CarbonSearchV2.Backends[i].MaxBatchSize = newConfig.CarbonSearchV2.MaxBatchSize
		}
	}

	newConfig.isSanitized = true
	return newConfig
}
