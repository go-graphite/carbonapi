package types

import (
	"time"

	"github.com/go-graphite/carbonapi/pkg/tlsconfig"
)

type BackendsV2 struct {
	Backends                  []BackendV2   `mapstructure:"backends"`
	MaxIdleConnsPerHost       int           `mapstructure:"maxIdleConnsPerHost"`
	ConcurrencyLimitPerServer int           `mapstructure:"concurrencyLimit"`
	Timeouts                  Timeouts      `mapstructure:"timeouts"`
	KeepAliveInterval         time.Duration `mapstructure:"keepAliveInterval"`
	MaxTries                  int           `mapstructure:"maxTries"`
	MaxBatchSize              *int          `mapstructure:"maxBatchSize"`
}

type BackendV2 struct {
	GroupName                 string                 `mapstructure:"groupName"`
	Protocol                  string                 `mapstructure:"protocol"`
	LBMethod                  string                 `mapstructure:"lbMethod"` // Valid: rr/roundrobin, broadcast/all
	Servers                   []string               `mapstructure:"servers"`
	Timeouts                  *Timeouts              `mapstructure:"timeouts"`
	ConcurrencyLimit          *int                   `mapstructure:"concurrencyLimit"`
	KeepAliveInterval         *time.Duration         `mapstructure:"keepAliveInterval"`
	MaxIdleConnsPerHost       *int                   `mapstructure:"maxIdleConnsPerHost"`
	MaxTries                  *int                   `mapstructure:"maxTries"`
	MaxBatchSize              *int                   `mapstructure:"maxBatchSize"`
	BackendOptions            map[string]interface{} `mapstructure:"backendOptions"`
	ForceAttemptHTTP2         bool                   `mapstructure:"forceAttemptHTTP2"`
	DoMultipleRequestsIfSplit bool                   `mapstructure:"doMultipleRequestsIfSplit"`
	IdleConnectionTimeout     *time.Duration         `mapstructure:"idleConnectionTimeout"`
	TLSClientConfig           *tlsconfig.TLSConfig   `mapstructure:"tlsClientConfig"`
	FetchClientType           string                 `mapstructure:"fetchClientType"` // Valid: std, fasthttp
}

func (b *BackendV2) FillDefaults() {
	if b.Timeouts == nil {
		b.Timeouts = &Timeouts{}
	}

	if b.Timeouts.Render == 0 {
		b.Timeouts.Render = 10000 * time.Second
	}

	if b.Timeouts.Find == 0 {
		b.Timeouts.Find = 10000 * time.Second
	}

	if b.Timeouts.Connect == 0 {
		b.Timeouts.Connect = 200 * time.Millisecond
	}

	if b.FetchClientType == "" {
		b.FetchClientType = "std"
	}
}
