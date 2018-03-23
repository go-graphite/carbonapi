package config

import (
	"time"

	"github.com/go-graphite/carbonzipper/zipper/types"
)

// Config is a structure that contains zipper-related configuration bits
type Config struct {
	ConcurrencyLimitPerServer int              `mapstructure:"concurrencyLimitPerServer"`
	MaxIdleConnsPerHost       int              `mapstructure:"maxIdleConnsPerHost"`
	Backends                  []string         `mapstructure:"backends"`
	BackendsV2                types.BackendsV2 `mapstructure:"backendsv2"`
	MaxGlobs                  int              `mapstructure:"maxGlobs"`
	MaxTries                  int              `mapstructure:"maxTries"`

	CarbonSearch   types.CarbonSearch
	CarbonSearchV2 types.CarbonSearchV2

	ExpireDelaySec    int32
	Timeouts          types.Timeouts
	KeepAliveInterval time.Duration `yaml:"keepAliveInterval"`
}
