package gosnowth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

// Config values represent configuration information SnowthClient values.
type Config struct {
	sync.RWMutex
	dialTimeout    time.Duration
	discover       bool
	servers        []string
	timeout        time.Duration
	watchInterval  time.Duration
	retries        int64
	connectRetries int64
}

// NewConfig creates and initializes a new SnowthClient configuration value.
func NewConfig(servers ...string) (*Config, error) {
	c := &Config{
		dialTimeout:    500 * time.Millisecond,
		discover:       false,
		servers:        []string{},
		timeout:        10 * time.Second,
		watchInterval:  30 * time.Second,
		retries:        0,
		connectRetries: -1,
	}

	if err := c.SetServers(servers...); err != nil {
		return nil, err
	}

	return c, nil
}

// MarshalJSON encodes a Config value into a JSON format byte slice.
func (c *Config) MarshalJSON() ([]byte, error) {
	c.RLock()
	m := struct {
		DialTimeout    string   `json:"dial_timeout,omitempty"`
		Discover       bool     `json:"discover"`
		Timeout        string   `json:"timeout,omitempty"`
		WatchInterval  string   `json:"watch_interval,omitempty"`
		Retries        int64    `json:"retries,omitempty"`
		ConnectRetries int64    `json:"connect_retries,omitempty"`
		Servers        []string `json:"servers,omitempty"`
	}{}

	if c.dialTimeout != 0 {
		m.DialTimeout = c.dialTimeout.String()
	}

	m.Discover = c.discover
	if c.timeout != 0 {
		m.Timeout = c.timeout.String()
	}

	if c.watchInterval != 0 {
		m.WatchInterval = c.watchInterval.String()
	}

	if c.retries != 0 {
		m.Retries = c.retries
	}

	if c.connectRetries != 0 {
		m.ConnectRetries = c.connectRetries
	}

	if len(c.servers) > 0 {
		m.Servers = make([]string, len(c.servers))
		copy(m.Servers, c.servers)
	}

	c.RUnlock()
	return json.Marshal(m)
}

// UnmarshalJSON decodes a JSON format byte slice into the Config value.
func (c *Config) UnmarshalJSON(b []byte) error {
	m := struct {
		DialTimeout    string   `json:"dial_timeout,omitempty"`
		Discover       bool     `json:"discover"`
		Timeout        string   `json:"timeout,omitempty"`
		WatchInterval  string   `json:"watch_interval,omitempty"`
		Retries        int64    `json:"retries,omitempty"`
		ConnectRetries int64    `json:"connect_retries,omitempty"`
		Servers        []string `json:"servers,omitempty"`
	}{}

	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %w", err)
	}

	if m.DialTimeout != "" {
		d, err := time.ParseDuration(m.DialTimeout)
		if err != nil {
			return fmt.Errorf("unable to parse dial timeout: %w", err)
		}

		if err := c.SetDialTimeout(d); err != nil {
			return err
		}
	}

	c.SetDiscover(m.Discover)
	if m.Timeout != "" {
		d, err := time.ParseDuration(m.Timeout)
		if err != nil {
			return fmt.Errorf("unable to parse timeout: %w", err)
		}

		if err := c.SetTimeout(d); err != nil {
			return err
		}
	}

	if m.WatchInterval != "" {
		d, err := time.ParseDuration(m.WatchInterval)
		if err != nil {
			return fmt.Errorf("unable to parse watch interval: %w", err)
		}

		if err := c.SetWatchInterval(d); err != nil {
			return err
		}
	}

	if m.Retries != 0 {
		c.retries = m.Retries
	}

	if m.ConnectRetries != 0 {
		c.connectRetries = m.ConnectRetries
	}

	if len(m.Servers) > 0 {
		if err := c.SetServers(m.Servers...); err != nil {
			return err
		}
	}

	return nil
}

// DialTimeout gets the initial connection timeout duration for attempts to
// connect to IRONdb. The default value is 500 milliseconds.
func (c *Config) DialTimeout() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c.dialTimeout
}

// SetDialTimeout sets a new dial timeout value.
func (c *Config) SetDialTimeout(t time.Duration) error {
	if t < 0 || t > time.Minute {
		return fmt.Errorf("invalid dial timeout value")
	}

	c.Lock()
	c.dialTimeout = t
	c.Unlock()
	return nil
}

// Discover gets whether the client should attempt to discover other IRONdb
// nodes in the same cluster as the provided node servers.
func (c *Config) Discover() bool {
	c.RLock()
	defer c.RUnlock()
	return c.discover
}

// SetDiscover sets whether the client should attempt to discover other IRONdb
// nodes in the same cluster as the provided node servers.
func (c *Config) SetDiscover(d bool) {
	c.Lock()
	c.discover = d
	c.Unlock()
}

// Timeout gets the timeout duration for HTTP requests to IRONdb. The default
// value is 10 seconds.
func (c *Config) Timeout() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c.timeout
}

// SetTimeout sets a new HTTP timeout duration.
func (c *Config) SetTimeout(t time.Duration) error {
	if t < 0 || t > (5*time.Minute) {
		return fmt.Errorf("invalid timeout value")
	}

	c.Lock()
	c.timeout = t
	c.Unlock()
	return nil
}

// Retries gets the number of times requests will be retried.
func (c *Config) Retries() int64 {
	c.RLock()
	defer c.RUnlock()
	return c.retries
}

// SetRetries sets the number of times requests will be retried.
func (c *Config) SetRetries(num int64) {
	c.Lock()
	c.retries = num
	c.Unlock()
}

// ConnectRetries gets the number of times requests will be retried on
// other nodes when network errors occur..
func (c *Config) ConnectRetries() int64 {
	c.RLock()
	defer c.RUnlock()
	return c.connectRetries
}

// SetConnectRetries sets the number of times requests will be retried on
// other nodes when network errors occur.
func (c *Config) SetConnectRetries(num int64) {
	c.Lock()
	c.connectRetries = num
	c.Unlock()
}

// Servers gets the list of IRONdb node servers to be used by a SnowthClient.
func (c *Config) Servers() []string {
	c.RLock()
	defer c.RUnlock()
	s := make([]string, len(c.servers))
	copy(s, c.servers)
	return s
}

// SetServers assigns a new list of server addresses and returns a pointer to the
// updated configuration value.
func (c *Config) SetServers(servers ...string) error {
	s := []string{}
	for _, svr := range servers {
		if _, err := url.Parse(svr); err != nil {
			return fmt.Errorf("invalid server address " + svr)
		}

		s = append(s, svr)
	}

	c.Lock()
	c.servers = s
	c.Unlock()
	return nil
}

// WatchInterval gets the frequency at which a SnowthClient will check for
// updates to the active status of its nodes if WatchAndUpdate() is called.
func (c *Config) WatchInterval() time.Duration {
	c.RLock()
	defer c.RUnlock()
	return c.watchInterval
}

// SetWatchInterval sets a new interval for watch and update functionality.
func (c *Config) SetWatchInterval(i time.Duration) error {
	if i < 0 || i > (time.Hour*24) {
		return fmt.Errorf("invalid watch interval value")
	}

	c.Lock()
	c.watchInterval = i
	c.Unlock()
	return nil
}
