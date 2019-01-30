package config

import (
	"time"

	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/interfaces"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/tagdb"
	zipperCfg "github.com/go-graphite/carbonapi/zipper/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"

	"github.com/lomik/zapwriter"
)

var DefaultLoggerConfig = zapwriter.Config{
	Logger:           "",
	File:             "stdout",
	Level:            "info",
	Encoding:         "console",
	EncodingTime:     "iso8601",
	EncodingDuration: "seconds",
}

type CacheConfig struct {
	Type              string   `mapstructure:"type"`
	Size              int      `mapstructure:"size_mb"`
	MemcachedServers  []string `mapstructure:"memcachedServers"`
	DefaultTimeoutSec int32    `mapstructure:"defaultTimeoutSec"`
}

type GraphiteConfig struct {
	Pattern  string
	Host     string
	Interval time.Duration
	Prefix   string
}

var Config = struct {
	ExtrapolateExperiment      bool               `mapstructure:"extrapolateExperiment"`
	Logger                     []zapwriter.Config `mapstructure:"logger"`
	Listen                     string             `mapstructure:"listen"`
	Buckets                    int                `mapstructure:"buckets"`
	Concurency                 int                `mapstructure:"concurency"`
	Cache                      CacheConfig        `mapstructure:"cache"`
	Cpus                       int                `mapstructure:"cpus"`
	TimezoneString             string             `mapstructure:"tz"`
	UnicodeRangeTables         []string           `mapstructure:"unicodeRangeTables"`
	Graphite                   GraphiteConfig     `mapstructure:"graphite"`
	IdleConnections            int                `mapstructure:"idleConnections"`
	PidFile                    string             `mapstructure:"pidFile"`
	SendGlobsAsIs              bool               `mapstructure:"sendGlobsAsIs"`
	AlwaysSendGlobsAsIs        bool               `mapstructure:"alwaysSendGlobsAsIs"`
	MaxBatchSize               int                `mapstructure:"maxBatchSize"`
	Zipper                     string             `mapstructure:"zipper"`
	Upstreams                  zipperCfg.Config   `mapstructure:"upstreams"`
	ExpireDelaySec             int32              `mapstructure:"expireDelaySec"`
	GraphiteWeb09Compatibility bool               `mapstructure:"graphite09compat"`
	IgnoreClientTimeout        bool               `mapstructure:"ignoreClientTimeout"`
	DefaultColors              map[string]string  `mapstructure:"defaultColors"`
	GraphTemplates             string             `mapstructure:"graphTemplates"`
	FunctionsConfigs           map[string]string  `mapstructure:"functionsConfig"`
	TagDB                      tagdb.Config       `mapstructure:"tagDB"`

	QueryCache cache.BytesCache `mapstructure:"-"`
	FindCache  cache.BytesCache `mapstructure:"-"`

	DefaultTimeZone *time.Location `mapstructure:"-"`

	// ZipperInstance is API entry to carbonzipper
	ZipperInstance interfaces.CarbonZipper `mapstructure:"-"`

	// Limiter limits concurrent zipper requests
	Limiter limiter.SimpleLimiter `mapstructure:"-"`

	TagDBProxy *tagdb.Http `mapstructure:"-"`
}{
	ExtrapolateExperiment: false,
	Listen:                "[::]:8081",
	Buckets:               10,
	Concurency:            20,
	SendGlobsAsIs:         false,
	AlwaysSendGlobsAsIs:   false,
	MaxBatchSize:          100,
	Cache: CacheConfig{
		Type:              "mem",
		DefaultTimeoutSec: 60,
	},
	TimezoneString: "",
	Graphite: GraphiteConfig{
		Pattern:  "{prefix}.{fqdn}",
		Host:     "",
		Interval: 60 * time.Second,
		Prefix:   "carbon.api",
	},
	Cpus:            0,
	IdleConnections: 10,
	PidFile:         "",

	QueryCache: cache.NullCache{},
	FindCache:  cache.NullCache{},

	DefaultTimeZone: time.Local,
	Logger:          []zapwriter.Config{DefaultLoggerConfig},

	Upstreams: zipperCfg.Config{
		Timeouts: zipperTypes.Timeouts{
			Render:  10000 * time.Second,
			Find:    2 * time.Second,
			Connect: 200 * time.Millisecond,
		},
		KeepAliveInterval: 30 * time.Second,

		MaxIdleConnsPerHost: 100,
	},
	ExpireDelaySec:             10 * 60,
	GraphiteWeb09Compatibility: false,

	TagDB: tagdb.Config{
		MaxConcurrentConnections: 10,
		MaxTries:                 3,
		Timeout:                  60 * time.Second,
		KeepAliveInterval:        30 * time.Second,
	},
}
