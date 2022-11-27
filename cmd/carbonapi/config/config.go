package config

import (
	"encoding/json"
	"time"

	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/interfaces"
	"github.com/go-graphite/carbonapi/limiter"
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
	Type                string        `mapstructure:"type"`
	Size                int           `mapstructure:"size_mb"`
	MemcachedServers    []string      `mapstructure:"memcachedServers"`
	DefaultTimeoutSec   int32         `mapstructure:"defaultTimeoutSec"`
	ShortTimeoutSec     int32         `mapstructure:"shortTimeoutSec"`
	ShortDuration       time.Duration `mapstructure:"shortDuration"`
	ShortUntilOffsetSec int64         `mapstructure:"shortUntilOffsetSec"`
}

type GraphiteConfig struct {
	Pattern  string
	Host     string
	Statsd   string
	Interval time.Duration
	Prefix   string
}

type Define struct {
	Name     string `mapstructure:"name"`
	Template string `mapstructure:"template"`
}

type ExpvarConfig struct {
	Listen       string `mapstructure:"listen"`
	Enabled      bool   `mapstructure:"enabled"`
	PProfEnabled bool   `mapstructure:"pprofEnabled"`
}

type Listener struct {
	Address string `mapstructure:"address"`

	// Server TLS
	ServerTLSConfig zipperTypes.TLSConfig `mapstructure:"serverTLSConfig"`

	// Client TLS
	ClientTLSConfig zipperTypes.TLSConfig `mapstructure:"clientTLSConfig"`
}

type DurationTruncate struct {
	Duration time.Duration
	Truncate time.Duration
}

type ConfigType struct {
	ExtrapolateExperiment      bool               `mapstructure:"extrapolateExperiment"`
	Logger                     []zapwriter.Config `mapstructure:"logger"`
	Listen                     string             `mapstructure:"listen"`
	Listeners                  []Listener         `mapstructure:"listeners"`
	Buckets                    int                `mapstructure:"buckets"`
	Concurency                 int                `mapstructure:"concurency"`
	ResponseCacheConfig        CacheConfig        `mapstructure:"cache"`
	BackendCacheConfig         CacheConfig        `mapstructure:"backendCache"`
	Cpus                       int                `mapstructure:"cpus"`
	TimezoneString             string             `mapstructure:"tz"`
	UnicodeRangeTables         []string           `mapstructure:"unicodeRangeTables"`
	Graphite                   GraphiteConfig     `mapstructure:"graphite"`
	IdleConnections            int                `mapstructure:"idleConnections"`
	PidFile                    string             `mapstructure:"pidFile"`
	SendGlobsAsIs              *bool              `mapstructure:"sendGlobsAsIs"`
	AlwaysSendGlobsAsIs        *bool              `mapstructure:"alwaysSendGlobsAsIs"`
	ExtractTagsFromArgs        bool               `mapstructure:"extractTagsFromArgs"`
	MaxBatchSize               int                `mapstructure:"maxBatchSize"`
	Zipper                     string             `mapstructure:"zipper"`
	Upstreams                  zipperCfg.Config   `mapstructure:"upstreams"`
	ExpireDelaySec             int32              `mapstructure:"expireDelaySec"`
	GraphiteWeb09Compatibility bool               `mapstructure:"graphite09compat"`
	IgnoreClientTimeout        bool               `mapstructure:"ignoreClientTimeout"`
	DefaultColors              map[string]string  `mapstructure:"defaultColors"`
	GraphTemplates             string             `mapstructure:"graphTemplates"`
	FunctionsConfigs           map[string]string  `mapstructure:"functionsConfig"`
	HeadersToPass              []string           `mapstructure:"headersToPass"`
	HeadersToLog               []string           `mapstructure:"headersToLog"`
	Define                     []Define           `mapstructure:"define"`
	Prefix                     string             `mapstructure:"prefix"`
	Expvar                     ExpvarConfig       `mapstructure:"expvar"`
	NotFoundStatusCode         int                `mapstructure:"notFoundStatusCode"`
	HTTPResponseStackTrace     bool               `mapstructure:"httpResponseStackTrace"`
	UseCachingDNSResolver      bool               `mapstructure:"useCachingDNSResolver"`
	CachingDNSRefreshTime      time.Duration      `mapstructure:"cachingDNSRefreshTime"`

	TruncateTimeMap map[time.Duration]time.Duration `mapstructure:"truncateTime"`
	TruncateTime    []DurationTruncate              `mapstructure:"-" json:"-"` // produce from TruncateTimeMap and sort in reverse order

	ResponseCache cache.BytesCache `mapstructure:"-" json:"-"`
	BackendCache  cache.BytesCache `mapstructure:"-" json:"-"`

	DefaultTimeZone *time.Location `mapstructure:"-" json:"-"`

	// ZipperInstance is API entry to carbonzipper
	ZipperInstance interfaces.CarbonZipper `mapstructure:"-" json:"-"`

	// Limiter limits concurrent zipper requests
	Limiter limiter.SimpleLimiter `mapstructure:"-" json:"-"`
}

// skipcq: CRT-P0003
func (c ConfigType) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return "Failed to marshal config: " + err.Error()
	} else {
		return string(data)
	}
}

var Config = ConfigType{
	ExtrapolateExperiment: false,
	Buckets:               10,
	Concurency:            1000,
	MaxBatchSize:          100,
	ResponseCacheConfig: CacheConfig{
		Type:              "mem",
		DefaultTimeoutSec: 60,
		ShortTimeoutSec:   0,
		ShortDuration:     0,
	},
	BackendCacheConfig: CacheConfig{
		Type:              "null",
		DefaultTimeoutSec: 0,
		ShortTimeoutSec:   0,
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

	ResponseCache: cache.NullCache{},
	BackendCache:  cache.NullCache{},

	DefaultTimeZone: time.Local,
	Logger:          []zapwriter.Config{DefaultLoggerConfig},

	Upstreams: zipperCfg.Config{
		Buckets:          10,
		SlowLogThreshold: 1 * time.Second,
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
	Prefix:                     "",
	Expvar: ExpvarConfig{
		Listen:       "",
		Enabled:      true,
		PProfEnabled: false,
	},
	NotFoundStatusCode:     200,
	HTTPResponseStackTrace: true,
	UseCachingDNSResolver:  false,
	CachingDNSRefreshTime:  1 * time.Minute,
}
