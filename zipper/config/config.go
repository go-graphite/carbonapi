package config

import (
	"time"

	"github.com/go-graphite/carbonzipper/pathcache"
	"github.com/go-graphite/carbonzipper/zipper/types"
)

// Config is a structure that contains zipper-related configuration bits
type Config struct {
	ConcurrencyLimitPerServer int
	MaxIdleConnsPerHost       int
	Backends                  []string
	BackendsV2                types.BackendsV2
	MaxGlobs                  int
	MaxTries                  int

	CarbonSearch   types.CarbonSearch
	CarbonSearchV2 types.CarbonSearchV2

	PathCache         pathcache.PathCache
	SearchCache       pathcache.PathCache
	Timeouts          types.Timeouts
	KeepAliveInterval time.Duration `yaml:"keepAliveInterval"`
}
