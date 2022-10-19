package http

import (
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

// NullSender  is disabled sender (if stat need to be disabled)
type NullSender struct{}

func (NullSender) Inc(string, int64, float32, ...statsd.Tag) error                    { return nil }
func (NullSender) Dec(string, int64, float32, ...statsd.Tag) error                    { return nil }
func (NullSender) Gauge(string, int64, float32, ...statsd.Tag) error                  { return nil }
func (NullSender) GaugeDelta(string, int64, float32, ...statsd.Tag) error             { return nil }
func (NullSender) Timing(string, int64, float32, ...statsd.Tag) error                 { return nil }
func (NullSender) TimingDuration(string, time.Duration, float32, ...statsd.Tag) error { return nil }
func (NullSender) Set(string, string, float32, ...statsd.Tag) error                   { return nil }
func (NullSender) SetInt(string, int64, float32, ...statsd.Tag) error                 { return nil }
func (NullSender) Raw(string, string, float32, ...statsd.Tag) error                   { return nil }
func (NullSender) NewSubStatter(string) statsd.SubStatter                             { return NullSender{} }
func (NullSender) SetPrefix(string)                                                   {}
func (NullSender) SetSamplerFunc(statsd.SamplerFunc)                                  {}
func (NullSender) Close() error                                                       { return nil }

var Gstatsd statsd.Statter = NullSender{}
