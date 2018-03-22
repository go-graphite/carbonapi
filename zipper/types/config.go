package types

import (
	"time"

	"github.com/go-graphite/carbonzipper/pathcache"
)

// Config is a structure that contains zipper-related configuration bits
type Config struct {
	ConcurrencyLimitPerServer int
	MaxIdleConnsPerHost       int
	Backends                  []string
	BackendsV2                BackendsV2
	MaxGlobs                  int
	MaxTries                  int

	CarbonSearch   CarbonSearch
	CarbonSearchV2 CarbonSearchV2

	PathCache         pathcache.PathCache
	SearchCache       pathcache.PathCache
	Timeouts          Timeouts
	KeepAliveInterval time.Duration `yaml:"keepAliveInterval"`
}
