package aliasByRedis

import (
	"context"
	"strings"
	"time"

	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"github.com/gomodule/redigo/redis"
)

func prepareMetric(metric string) (string, []string) {
	parts := strings.Split(metric, ".")
	return parts[len(parts)-1], parts
}

func redisGetHash(name, key string, c redis.Conn, timeout time.Duration) (string, error) {
	v, err := redis.DoWithTimeout(c, timeout, "HGET", key, name)
	return redis.String(v, err)
}

type Database struct {
	Enabled            bool
	MaxIdleConnections *int
	IdleTimeout        *time.Duration
	PingInterval       *time.Duration
	Address            *string
	DatabaseNumber     *int
	Username           *string
	Password           *string
	ConnectTimeout     *time.Duration
	QueryTimeout       *time.Duration
	KeepAliveInterval  *time.Duration
	UseTLS             *bool
	TLSSkipVerify      *bool
}

type aliasByRedis struct {
	interfaces.FunctionBase

	address      string
	dialOptions  []redis.DialOption
	queryTimeout time.Duration
	pool         *redis.Pool
	pingInterval *time.Duration
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "aliasByRedis"))
	if configFile == "" {
		logger.Debug("no config file specified",
			zap.String("message", "this function requrires config file to work properly"),
		)
		return nil
	}
	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil {
		logger.Error("failed to read config file",
			zap.Error(err),
		)
		return nil
	}

	cfg := Database{}

	err = v.Unmarshal(&cfg)
	if err != nil {
		logger.Error("failed to parse config",
			zap.Error(err),
		)
		return nil
	}

	logger.Info("will use configuration",
		zap.Any("config", cfg),
	)

	if !cfg.Enabled {
		logger.Warn("aliasByRedis config found but aliasByRedis is disabled")
		return nil
	}

	f := &aliasByRedis{
		address:      "127.0.0.1:6379",
		dialOptions:  make([]redis.DialOption, 0),
		queryTimeout: 250 * time.Millisecond,
	}

	if cfg.Address != nil {
		f.address = *cfg.Address
	}

	if cfg.DatabaseNumber != nil {
		f.dialOptions = append(f.dialOptions, redis.DialDatabase(*cfg.DatabaseNumber))
	}

	if cfg.Username != nil {
		f.dialOptions = append(f.dialOptions, redis.DialUsername(*cfg.Username))
	}

	if cfg.Password != nil {
		f.dialOptions = append(f.dialOptions, redis.DialPassword(*cfg.Password))
	}

	if cfg.KeepAliveInterval != nil {
		f.dialOptions = append(f.dialOptions, redis.DialKeepAlive(*cfg.KeepAliveInterval))
	}

	if cfg.ConnectTimeout != nil {
		f.dialOptions = append(f.dialOptions, redis.DialConnectTimeout(*cfg.ConnectTimeout))
	}

	if cfg.UseTLS != nil {
		f.dialOptions = append(f.dialOptions, redis.DialUseTLS(*cfg.UseTLS))
	}

	if cfg.TLSSkipVerify != nil {
		f.dialOptions = append(f.dialOptions, redis.DialTLSSkipVerify(*cfg.TLSSkipVerify))
	}

	f.pingInterval = cfg.PingInterval

	maxIdle := 1
	if cfg.MaxIdleConnections != nil && *cfg.MaxIdleConnections > 1 {
		maxIdle = *cfg.MaxIdleConnections
	}

	idleTimeout := 60 * time.Second
	if cfg.IdleTimeout != nil {
		idleTimeout = *cfg.IdleTimeout
	}

	f.pool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", f.address, f.dialOptions...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if f.pingInterval == nil || time.Since(t) < *f.pingInterval {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aliasByRedis"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasByRedis) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	redisHashName, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	keepPath, err := e.GetBoolArgDefault(2, false)
	if err != nil {
		return nil, err
	}
	_ = keepPath

	redisConnection, err := f.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer redisConnection.Close()

	results := make([]*types.MetricData, len(args))

	for i, a := range args {
		var r *types.MetricData
		name, nodes := prepareMetric(a.Tags["name"])
		redisName, err := redisGetHash(name, redisHashName, redisConnection, f.queryTimeout)
		if err == nil {
			if keepPath {
				nodes[len(nodes)-1] = redisName
				r = a.CopyName(strings.Join(nodes, "."))
			} else {
				r = a.CopyName(redisName)
			}
		} else {
			r = a.CopyLink()
		}
		results[i] = r
	}

	return results, nil
}

func (f *aliasByRedis) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByHash": {
			Description: "Takes a seriesList, extracts first part of a metric name and use it as a field name for HGET redis query. Key name is specified by argument.\n\n.. code-block:: none\n\n  &target=aliasByRedis(some.metric, \"redis_key_name\")",
			Function:    "aliasByRedis(seriesList, keyName[, keepPath])",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByRedis",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "redisKey",
					Required: true,
					Type:     types.String,
				},
				{
					Name: "keepPath",
					Type: types.Boolean,
					Default: &types.Suggestion{
						Value: false,
						Type:  types.SBool,
					},
				},
			},
			NameChange: true, // name changed
			TagsChange: true, // name tag changed
		},
	}
}
