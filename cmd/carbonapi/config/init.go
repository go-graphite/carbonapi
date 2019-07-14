package config

import (
	"bytes"
	"expvar"
	"fmt"
	"github.com/ansel1/merry"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/facebookgo/pidfile"
	"github.com/go-graphite/carbonapi/cache"
	"github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/functions/cairo/png"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/rewrite"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/pkg/parser"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var graphTemplates map[string]png.PictureParams

func SetUpConfig(logger *zap.Logger, BuildVersion string) {
	Config.Cache.MemcachedServers = viper.GetStringSlice("cache.memcachedServers")
	if n := viper.GetString("logger.logger"); n != "" {
		Config.Logger[0].Logger = n
	}
	if n := viper.GetString("logger.file"); n != "" {
		Config.Logger[0].File = n
	}
	if n := viper.GetString("logger.level"); n != "" {
		Config.Logger[0].Level = n
	}
	if n := viper.GetString("logger.encoding"); n != "" {
		Config.Logger[0].Encoding = n
	}
	if n := viper.GetString("logger.encodingtime"); n != "" {
		Config.Logger[0].EncodingTime = n
	}
	if n := viper.GetString("logger.encodingduration"); n != "" {
		Config.Logger[0].EncodingDuration = n
	}
	err := zapwriter.ApplyConfig(Config.Logger)
	if err != nil {
		logger.Fatal("failed to initialize logger with requested configuration",
			zap.Any("configuration", Config.Logger),
			zap.Error(err),
		)
	}

	needStackTrace := false
	for _, l := range Config.Logger {
		if strings.ToLower(l.Level) == "debug" {
			needStackTrace = true
			break
		}
	}
	fmt.Printf("\n\n\n\n\nneedStackTrace=%v\n\n\n\n", needStackTrace)
	merry.SetStackCaptureEnabled(needStackTrace)

	if Config.GraphTemplates != "" {
		graphTemplates = make(map[string]png.PictureParams)
		graphTemplatesViper := viper.New()
		b, err := ioutil.ReadFile(Config.GraphTemplates)
		if err != nil {
			logger.Fatal("error reading graphTemplates file",
				zap.String("graphTemplate_path", Config.GraphTemplates),
				zap.Error(err),
			)
		}

		if strings.HasSuffix(Config.GraphTemplates, ".toml") {
			logger.Info("will parse config as toml",
				zap.String("graphTemplate_path", Config.GraphTemplates),
			)
			graphTemplatesViper.SetConfigType("TOML")
		} else {
			logger.Info("will parse config as yaml",
				zap.String("graphTemplate_path", Config.GraphTemplates),
			)
			graphTemplatesViper.SetConfigType("YAML")
		}

		err = graphTemplatesViper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.String("graphTemplate_path", Config.GraphTemplates),
				zap.Error(err),
			)
		}

		for k := range graphTemplatesViper.AllSettings() {
			// we need to explicitly copy	YDivisors and ColorList
			newStruct := png.DefaultParams
			newStruct.ColorList = nil
			newStruct.YDivisors = nil
			sub := graphTemplatesViper.Sub(k)
			err = sub.Unmarshal(&newStruct)
			if err != nil {
				logger.Error("failed to parse graphTemplates config, settings will be ignored",
					zap.String("graphTemplate_path", Config.GraphTemplates),
					zap.Error(err),
				)
			}
			if newStruct.ColorList == nil || len(newStruct.ColorList) == 0 {
				newStruct.ColorList = make([]string, len(png.DefaultParams.ColorList))
				copy(newStruct.ColorList, png.DefaultParams.ColorList)
			}
			if newStruct.YDivisors == nil || len(newStruct.YDivisors) == 0 {
				newStruct.YDivisors = make([]float64, len(png.DefaultParams.YDivisors))
				copy(newStruct.YDivisors, png.DefaultParams.YDivisors)
			}
			graphTemplates[k] = newStruct
		}

		for name, params := range graphTemplates {
			png.SetTemplate(name, params)
		}
	}

	if Config.DefaultColors != nil {
		for name, color := range Config.DefaultColors {
			err = png.SetColor(name, color)
			if err != nil {
				logger.Warn("invalid color specified and will be ignored",
					zap.String("reason", "color must be valid hex rgb or rbga value, e.x. '#c80032', 'c80032', 'c80032ff', etc."),
					zap.Error(err),
				)
			}
		}
	}

	if Config.FunctionsConfigs != nil {
		logger.Info("extra configuration for functions found",
			zap.Any("extra_config", Config.FunctionsConfigs),
		)
	} else {
		Config.FunctionsConfigs = make(map[string]string)
	}

	rewrite.New(Config.FunctionsConfigs)
	functions.New(Config.FunctionsConfigs)

	expvar.NewString("GoVersion").Set(runtime.Version())
	expvar.NewString("BuildVersion").Set(BuildVersion)
	expvar.Publish("config", Config)

	Config.Limiter = limiter.NewSimpleLimiter(Config.Concurency)

	switch Config.Cache.Type {
	case "memcache":
		if len(Config.Cache.MemcachedServers) == 0 {
			logger.Fatal("memcache cache requested but no memcache servers provided")
		}

		logger.Info("memcached configured",
			zap.Strings("servers", Config.Cache.MemcachedServers),
		)
		Config.QueryCache = cache.NewMemcached("capi", Config.Cache.MemcachedServers...)
		// find cache is only used if SendGlobsAsIs is false.
		if !Config.SendGlobsAsIs {
			Config.FindCache = cache.NewExpireCache(0)
		}
	case "mem":
		Config.QueryCache = cache.NewExpireCache(uint64(Config.Cache.Size * 1024 * 1024))

		// find cache is only used if SendGlobsAsIs is false.
		if !Config.SendGlobsAsIs {
			Config.FindCache = cache.NewExpireCache(0)
		}
	case "null":
		// defaults
		Config.QueryCache = cache.NullCache{}
		Config.FindCache = cache.NullCache{}
	default:
		logger.Error("unknown cache type",
			zap.String("cache_type", Config.Cache.Type),
			zap.Strings("known_cache_types", []string{"null", "mem", "memcache"}),
		)
	}

	if Config.TimezoneString != "" {
		fields := strings.Split(Config.TimezoneString, ",")

		if len(fields) != 2 {
			logger.Fatal("unexpected amount of fields in tz",
				zap.String("timezone_string", Config.TimezoneString),
				zap.Int("fields_got", len(fields)),
				zap.Int("fields_expected", 2),
			)
		}

		offs, err := strconv.Atoi(fields[1])
		if err != nil {
			logger.Fatal("unable to parse seconds",
				zap.String("field[1]", fields[1]),
				zap.Error(err),
			)
		}

		Config.DefaultTimeZone = time.FixedZone(fields[0], offs)
		logger.Info("using fixed timezone",
			zap.String("timezone", Config.DefaultTimeZone.String()),
			zap.Int("offset", offs),
		)
	}

	if len(Config.UnicodeRangeTables) != 0 {
		if strings.ToLower(Config.UnicodeRangeTables[0]) == "all" {
			for _, t := range unicode.Scripts {
				parser.RangeTables = append(parser.RangeTables, t)
			}
		} else {
			for _, stringRange := range Config.UnicodeRangeTables {
				t, ok := unicode.Scripts[stringRange]
				if !ok {
					supportedTables := make([]string, 0)
					for tt := range unicode.Scripts {
						supportedTables = append(supportedTables, tt)
					}
					logger.Fatal("unknown unicode table",
						zap.String("specified_table", stringRange),
						zap.Strings("supported_tables", supportedTables),
						zap.String("more_info", "you need to specify the table, by it's alias in unicode"+
							" 10.0.0, see https://golang.org/src/unicode/tables.go?#L3437"),
					)
				}
				parser.RangeTables = append(parser.RangeTables, t)
			}
		}
	} else {
		parser.RangeTables = append(parser.RangeTables, unicode.Latin)
	}

	if Config.Cpus != 0 {
		runtime.GOMAXPROCS(Config.Cpus)
	}

	if Config.PidFile != "" {
		pidfile.SetPidfilePath(Config.PidFile)
		err := pidfile.Write()
		if err != nil {
			logger.Fatal("error during pidfile.Write()",
				zap.Error(err),
			)
		}
	}

	helper.ExtrapolatePoints = Config.ExtrapolateExperiment
	if Config.ExtrapolateExperiment {
		logger.Warn("extraploation experiment is enabled",
			zap.String("reason", "this feature is highly experimental and untested"),
		)
	}

	for _, define := range Config.Define {
		if define.Name == "" {
			logger.Fatal("empty define name")
		}
		err := parser.Define(define.Name, define.Template)
		if err != nil {
			logger.Fatal("unable to compile define template",
				zap.Error(err),
				zap.String("template", define.Template),
			)
		}
	}
}

func SetUpViper(logger *zap.Logger, configPath *string, viperPrefix string) {
	if *configPath != "" {
		b, err := ioutil.ReadFile(*configPath)
		if err != nil {
			logger.Fatal("error reading config file",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}

		if strings.HasSuffix(*configPath, ".toml") {
			logger.Info("will parse config as toml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("TOML")
		} else {
			logger.Info("will parse config as yaml",
				zap.String("config_file", *configPath),
			)
			viper.SetConfigType("YAML")
		}
		err = viper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			logger.Fatal("failed to parse config",
				zap.String("config_path", *configPath),
				zap.Error(err),
			)
		}
	}

	if viperPrefix != "" {
		viper.SetEnvPrefix(viperPrefix)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("tz", "carbonapi_tz")
	viper.SetDefault("listen", "localhost:8081")
	viper.SetDefault("concurency", 20)
	viper.SetDefault("cache.type", "mem")
	viper.SetDefault("cache.size_mb", 0)
	viper.SetDefault("cache.defaultTimeoutSec", 60)
	viper.SetDefault("cache.memcachedServers", []string{})
	viper.SetDefault("cpus", 0)
	viper.SetDefault("tz", "")
	viper.SetDefault("sendGlobsAsIs", false)
	viper.SetDefault("AlwaysSendGlobsAsIs", false)
	viper.SetDefault("maxBatchSize", 100)
	viper.SetDefault("graphite.host", "")
	viper.SetDefault("graphite.interval", "60s")
	viper.SetDefault("graphite.prefix", "carbon.api")
	viper.SetDefault("graphite.pattern", "{prefix}.{fqdn}")
	viper.SetDefault("idleConnections", 10)
	viper.SetDefault("pidFile", "")
	viper.SetDefault("upstreams.internalRoutingCache", "600s")
	viper.SetDefault("upstreams.buckets", 10)
	viper.SetDefault("upstreams.timeouts.global", "10s")
	viper.SetDefault("upstreams.timeouts.afterStarted", "2s")
	viper.SetDefault("upstreams.timeouts.connect", "200ms")
	viper.SetDefault("upstreams.concurrencyLimit", 0)
	viper.SetDefault("upstreams.keepAliveInterval", "30s")
	viper.SetDefault("upstreams.maxIdleConnsPerHost", 100)
	viper.SetDefault("upstreams.carbonsearch.backend", "")
	viper.SetDefault("upstreams.carbonsearch.prefix", "virt.v1.*")
	viper.SetDefault("upstreams.graphite09compat", false)
	viper.SetDefault("expireDelaySec", 600)
	viper.SetDefault("logger", map[string]string{})
	viper.AutomaticEnv()

	err := viper.Unmarshal(&Config)
	if err != nil {
		logger.Fatal("failed to parse config",
			zap.Error(err),
		)
	}
}

func SetUpConfigUpstreams(logger *zap.Logger) {
	if Config.Zipper != "" {
		logger.Warn("found legacy 'zipper' option, will use it instead of any 'upstreams' specified. This will be removed in future versions!")

		Config.Upstreams.Backends = []string{Config.Zipper}
		Config.Upstreams.ConcurrencyLimitPerServer = Config.Concurency
		Config.Upstreams.MaxIdleConnsPerHost = Config.IdleConnections
		Config.Upstreams.MaxBatchSize = Config.MaxBatchSize
		Config.Upstreams.KeepAliveInterval = 10 * time.Second
		// To emulate previous behavior
		Config.Upstreams.Timeouts = zipperTypes.Timeouts{
			Connect: 1 * time.Second,
			Render:  600 * time.Second,
			Find:    600 * time.Second,
		}
	}
	if len(Config.Upstreams.Backends) == 0 && len(Config.Upstreams.BackendsV2.Backends) == 0 {
		logger.Fatal("no backends specified for upstreams!")
	}

}
