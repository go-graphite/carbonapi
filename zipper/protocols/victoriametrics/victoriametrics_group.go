package victoriametrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ansel1/merry"
	"github.com/valyala/fastjson"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"

	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus"

	"go.uber.org/zap"
)

func init() {
	aliases := []string{"victoriametrics", "vm"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// VictoriaMetricsGroup is a protocol group that can query victoria-metrics
type VictoriaMetricsGroup struct {
	types.BackendServer

	groupName string
	servers   []string
	protocol  string

	client *http.Client

	limiter              limiter.ServerLimiter
	logger               *zap.Logger
	timeout              types.Timeouts
	maxTries             int
	maxMetricsPerRequest int

	step                 int64
	maxPointsPerQuery    int64
	forceMinStepInterval time.Duration
	vmClusterTenantID    string

	startDelay           prometheus.StartDelay
	probeVersionInterval time.Duration
	fallbackVersion      string

	httpQuery  *helper.HttpQuery
	parserPool fastjson.ParserPool

	featureSet atomic.Value // *vmSupportedFeatures
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "victoriametrics"), zap.String("protocol", config.Protocol), zap.String("name", config.GroupName))
	httpClient := helper.GetHTTPClient(logger, config)

	step := int64(15)
	var vmClusterTenantID string = ""
	vmClusterTenantIDI, ok := config.BackendOptions["vmclustertenantid"]
	if ok {
		vmClusterTenantID = vmClusterTenantIDI.(string)
	}
	stepI, ok := config.BackendOptions["step"]
	if ok {
		stepNew, ok := stepI.(string)
		if ok {
			if stepNew[len(stepNew)-1] >= '0' && stepNew[len(stepNew)-1] <= '9' {
				stepNew += "s"
			}
			t, err := time.ParseDuration(stepNew)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "step"),
					zap.String("option_value", stepNew),
					zap.Error(err),
				)
			}
			step = int64(t.Seconds())
		} else {
			logger.Fatal("failed to parse step",
				zap.String("type_parsed", fmt.Sprintf("%T", stepI)),
				zap.String("type_expected", "string"),
			)
		}
	}

	maxPointsPerQuery := int64(11000)
	mppqI, ok := config.BackendOptions["max_points_per_query"]
	if ok {
		mppq, ok := mppqI.(int)
		if !ok {
			logger.Fatal("failed to parse max_points_per_query",
				zap.String("type_parsed", fmt.Sprintf("%T", mppqI)),
				zap.String("type_expected", "int"),
			)
		}

		maxPointsPerQuery = int64(mppq)
	}

	var forceMinStepInterval time.Duration
	fmsiI, ok := config.BackendOptions["force_min_step_interval"]
	if ok {
		fmsiS, ok := fmsiI.(string)
		if !ok {
			logger.Fatal("failed to parse force_min_step_interval",
				zap.String("type_parsed", fmt.Sprintf("%T", fmsiI)),
				zap.String("type_expected", "time.Duration"),
			)
		}
		var err error
		forceMinStepInterval, err = time.ParseDuration(fmsiS)
		if err != nil {
			logger.Fatal("failed to parse force_min_step_interval",
				zap.String("value_provided", fmsiS),
				zap.String("type_expected", "time.Duration"),
			)
		}
	}

	delay := prometheus.StartDelay{
		IsSet:      false,
		IsDuration: false,
		T:          -1,
	}
	startI, ok := config.BackendOptions["start"]
	if ok {
		delay.IsSet = true
		startNew, ok := startI.(string)
		if ok {
			startNewInt, err := strconv.Atoi(startNew)
			if err != nil {
				d, err2 := time.ParseDuration(startNew)
				if err2 != nil {
					logger.Fatal("failed to parse option",
						zap.String("option_name", "start"),
						zap.String("option_value", startNew),
						zap.Errors("errors", []error{err, err2}),
					)
				}
				delay.IsDuration = true
				delay.D = d
			} else {
				delay.T = int64(startNewInt)
			}
		}
	}

	periodicProbe := false
	probeVersionInterval, _ := time.ParseDuration("600s")
	probeVersionIntervalParam, ok := config.BackendOptions["probe_version_interval"]
	if ok {
		probeVersionIntervalStr, ok := probeVersionIntervalParam.(string)
		if ok && probeVersionIntervalStr != "never" {
			interval, err := time.ParseDuration(probeVersionIntervalStr)
			if err != nil {
				logger.Fatal("failed to parse option",
					zap.String("option_name", "start"),
					zap.String("option_value", probeVersionIntervalStr),
					zap.Errors("errors", []error{err}),
				)
			}
			probeVersionInterval = interval
			periodicProbe = true
		} else {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "start"),
				zap.Any("option_value", probeVersionIntervalParam),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	} else {
		periodicProbe = true
	}

	fallbackVersion := "v0.0.0"
	fallbackVersionParam, ok := config.BackendOptions["fallback_version"]
	if ok {
		fallbackVersion, ok = fallbackVersionParam.(string)
		if !ok {
			logger.Fatal("failed to parse option",
				zap.String("option_name", "start"),
				zap.Any("option_value", probeVersionIntervalParam),
				zap.Errors("errors", []error{fmt.Errorf("not a string")}),
			)
		}
	}

	httpQuery := helper.NewHttpQuery(config.GroupName, config.Servers, *config.MaxTries, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv2PB)

	c := &VictoriaMetricsGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		protocol:             config.Protocol,
		timeout:              *config.Timeouts,
		maxTries:             *config.MaxTries,
		maxMetricsPerRequest: *config.MaxBatchSize,

		step:                 step,
		maxPointsPerQuery:    maxPointsPerQuery,
		vmClusterTenantID:    vmClusterTenantID,
		startDelay:           delay,
		probeVersionInterval: probeVersionInterval,
		fallbackVersion:      fallbackVersion,

		client:  httpClient,
		limiter: limiter,
		logger:  logger,

		httpQuery: httpQuery,
	}

	promLogger := logger.With(zap.String("subclass", "prometheus"))
	c.BackendServer, _ = prometheus.NewWithEverythingInitialized(promLogger, config, tldCacheDisabled, requireSuccessAll, limiter, step, maxPointsPerQuery, forceMinStepInterval, delay, httpQuery, httpClient)

	c.updateFeatureSet(context.Background())

	logger.Info("periodic probe for version change",
		zap.Bool("enabled", periodicProbe),
		zap.Duration("interval", c.probeVersionInterval),
	)
	if periodicProbe {
		go c.probeVMVersion(context.Background())
	}

	return c, nil
}

func New(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool) (types.BackendServer, merry.Error) {
	if config.ConcurrencyLimit == nil {
		return nil, types.ErrConcurrencyLimitNotSet
	}
	if len(config.Servers) == 0 {
		return nil, types.ErrNoServersSpecified
	}
	l := limiter.NewServerLimiter([]string{config.GroupName}, *config.ConcurrencyLimit)

	return NewWithLimiter(logger, config, tldCacheDisabled, requireSuccessAll, l)
}
