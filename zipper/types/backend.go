package types

import (
	"time"
)

type BackendsV2 struct {
	Backends                  []BackendV2   `yaml:"backends"`
	MaxIdleConnsPerHost       int           `yaml:"maxIdleConnsPerHost"`
	ConcurrencyLimitPerServer int           `yaml:"concurrencyLimit"`
	Timeouts                  Timeouts      `yaml:"timeouts"`
	KeepAliveInterval         time.Duration `yaml:"keepAliveInterval"`
	MaxTries                  int           `yaml:"maxTries"`
	MaxGlobs                  int           `yaml:"maxGlobs"`
}

type BackendV2 struct {
	GroupName           string         `yaml:"groupName"`
	Protocol            string         `yaml:"protocol"`
	LBMethod            LBMethod       `yaml:"lbMethod"` // Valid: rr/roundrobin, broadcast/all
	Servers             []string       `yaml:"servers"`
	Timeouts            *Timeouts      `yaml:"timeouts"`
	ConcurrencyLimit    *int           `yaml:"concurrencyLimit"`
	KeepAliveInterval   *time.Duration `yaml:"keepAliveInterval"`
	MaxIdleConnsPerHost *int           `yaml:"maxIdleConnsPerHost"`
	MaxTries            *int           `yaml:"maxTries"`
	MaxGlobs            int            `yaml:"maxGlobs"`
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

	if b.ConcurrencyLimit == nil {

	}
}

// CarbonSearch is a structure that contains carbonsearch related configuration bits
type CarbonSearch struct {
	Backend string `yaml:"backend"`
	Prefix  string `yaml:"prefix"`
}

type CarbonSearchV2 struct {
	BackendsV2
	Prefix string `yaml:"prefix"`
}
