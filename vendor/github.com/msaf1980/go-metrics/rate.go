package metrics

import (
	"sync"
	"time"
)

type RateNames interface {
	SetName(string) Rate
	Name() string
	SetRateName(string) Rate
	RateName() string
}

// Rate hold an int64 value and timestamp (current and previous) and return diff and diff/s.
type Rate interface {
	RateNames
	Snapshot() Rate
	Clear() (int64, float64)
	Update(v int64)
	UpdateTs(v int64, timestamp_ns int64)
	Values() (int64, float64)
}

// GetOrRegisterRate returns an existing Rate or constructs and registers a
// new StandardRate.
func GetOrRegisterRate(name string, r Registry) Rate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewRate).(Rate)
}

// GetOrRegisterRateT returns an existing Rate or constructs and registers a
// new StandardRate.
func GetOrRegisterRateT(name string, tagsMap map[string]string, r Registry) Rate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewRate).(Rate)
}

// NewRate constructs a new StandardRate.
func NewRate() Rate {
	if UseNilMetrics {
		return NilRate{}
	}
	return &StandardRate{name: ".value", rateName: ".rate"}
}

// NewRegisteredRate constructs and registers a new StandardRate.
func NewRegisteredRate(name string, r Registry) Rate {
	c := NewRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredRateT constructs and registers a new StandardRate.
func NewRegisteredRateT(name string, tagsMap map[string]string, r Registry) Rate {
	c := NewRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// RateSnapshot is a read-only copy of another Rate.
type RateSnapshot struct {
	name     string
	rateName string
	value    int64
	rate     float64
}

func newRateSnapshot(v int64, rate float64, name, rateName string) Rate {
	return &RateSnapshot{value: v, rate: rate, name: name, rateName: rateName}
}

// Snapshot returns the snapshot.
func (g *RateSnapshot) Snapshot() Rate { return g }

func (g *RateSnapshot) Name() string {
	return g.name
}

func (g *RateSnapshot) RateName() string {
	return g.rateName
}

func (RateSnapshot) SetName(string) Rate {
	panic("SetName called on a RateSnapshot")
}
func (RateSnapshot) SetRateName(string) Rate {
	panic("SetRateName called on a RateSnapshot")
}

// Clear panics.
func (RateSnapshot) Clear() (int64, float64) {
	panic("Clear called on a RateSnapshot")
}

// Update panics.
func (RateSnapshot) Update(int64) {
	panic("Update called on a RateSnapshot")
}

// Update panics.
func (RateSnapshot) UpdateTs(int64, int64) {
	panic("Update called on a RateSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g *RateSnapshot) Values() (int64, float64) { return g.value, g.rate }

// NilRate is a no-op Rate.
type NilRate struct{}

func (NilRate) Name() string {
	return ""
}

func (NilRate) RateName() string {
	return ""
}

func (g NilRate) SetName(string) Rate { return g }

func (g NilRate) SetRateName(string) Rate { return g }

// Snapshot is a no-op.
func (NilRate) Snapshot() Rate { return NilRate{} }

// Clear is a no-op.
func (NilRate) Clear() (int64, float64) { return 0, 0 }

// Update is a no-op.
func (NilRate) Update(int64) {}

// UpdateTs is a no-op.
func (NilRate) UpdateTs(int64, int64) {}

// Value is a no-op.
func (NilRate) Values() (int64, float64) { return 0, 0 }

// StandardRate is the standard implementation of a Rate and uses the
// sync/atomic package to manage a single int64 value.
type StandardRate struct {
	name     string
	rateName string
	prev     int64
	prevTs   int64
	value    int64
	valueTs  int64
	mutex    sync.Mutex
}

// Snapshot returns a read-only copy of the Rate.
func (g *StandardRate) Snapshot() Rate {
	v, rate := g.Values()
	return newRateSnapshot(v, rate, g.name, g.rateName)
}

// Clear sets the DownCounter to zero.
func (g *StandardRate) Clear() (int64, float64) {
	var (
		v int64
		p int64
		d int64
	)
	g.mutex.Lock()
	v = g.value
	if g.prevTs > 0 {
		p = g.prev
		d = g.valueTs - g.prevTs
	}
	g.value = 0
	g.prev = 0
	g.valueTs = 0
	g.prevTs = 0
	g.mutex.Unlock()
	if d == 0 {
		// broken values
		return v, 0
	}
	return v, 1e9 * float64(v-p) / float64(d)
}

// Update updates the Rate's value.
func (g *StandardRate) Update(v int64) {
	ts := time.Now().UnixNano()
	g.mutex.Lock()
	g.prev = g.value
	g.prevTs = g.valueTs
	g.value = v
	g.valueTs = ts
	g.mutex.Unlock()
}

// UpdateTs updates the Rate's value.
func (g *StandardRate) UpdateTs(v int64, ts int64) {
	g.mutex.Lock()
	g.prev = g.value
	g.prevTs = g.valueTs
	g.value = v
	g.valueTs = ts
	g.mutex.Unlock()
}

// Value returns the Rate's current value.
func (g *StandardRate) Values() (int64, float64) {
	var (
		v int64
		p int64
		d int64
	)
	g.mutex.Lock()
	v = g.value
	if g.prevTs > 0 {
		p = g.prev
		d = g.valueTs - g.prevTs
	}
	g.mutex.Unlock()
	if d == 0 {
		// first or immediate try
		return v, 0
	}
	return v, 1e9 * float64(v-p) / float64(d)
}

func (g *StandardRate) SetName(name string) Rate {
	g.mutex.Lock()
	g.name = name
	g.mutex.Unlock()
	return g
}

func (g *StandardRate) Name() string {
	return g.name
}

func (g *StandardRate) SetRateName(rateName string) Rate {
	g.mutex.Lock()
	g.rateName = rateName
	g.mutex.Unlock()
	return g
}

func (g *StandardRate) RateName() string {
	return g.rateName
}
