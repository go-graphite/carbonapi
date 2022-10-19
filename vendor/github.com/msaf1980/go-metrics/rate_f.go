package metrics

import (
	"sync"
	"time"
)

type FRateNames interface {
	SetName(string) FRate
	Name() string
	SetRateName(string) FRate
	RateName() string
}

// FRate hold an int64 value and timestamp (current and previous) and return diff and diff/s.
type FRate interface {
	FRateNames
	Snapshot() FRate
	Clear() (float64, float64)
	Update(v float64)
	UpdateTs(v float64, timestamp_ns int64)
	Values() (float64, float64)
}

// GetOrRegisterFRate returns an existing FRate or constructs and registers a
// new StandardFRate.
func GetOrRegisterFRate(name string, r Registry) FRate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewFRate).(FRate)
}

// GetOrRegisterFRateT returns an existing FRate or constructs and registers a
// new StandardFRate.
func GetOrRegisterFRateT(name string, tagsMap map[string]string, r Registry) FRate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewFRate).(FRate)
}

// NewFRate constructs a new StandardFRate.
func NewFRate() FRate {
	if UseNilMetrics {
		return NilFRate{}
	}
	return &StandardFRate{name: ".value", rateName: ".rate"}
}

// NewRegisteredFRate constructs and registers a new StandardFRate.
func NewRegisteredFRate(name string, r Registry) FRate {
	c := NewFRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredFRateT constructs and registers a new StandardFRate.
func NewRegisteredFRateT(name string, tagsMap map[string]string, r Registry) FRate {
	c := NewFRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// FRateSnapshot is a read-only copy of another FRate.
type FRateSnapshot struct {
	name     string
	rateName string
	value    float64
	FRate    float64
}

func newFRateSnapshot(v, FRate float64, name, rateName string) FRate {
	return &FRateSnapshot{value: v, FRate: FRate, name: name, rateName: name}
}

// Snapshot returns the snapshot.
func (g *FRateSnapshot) Snapshot() FRate { return g }

func (g *FRateSnapshot) Name() string {
	return g.name
}

func (g *FRateSnapshot) RateName() string {
	return g.rateName
}

func (FRateSnapshot) SetName(string) FRate {
	panic("SetName called on a RateSnapshot")
}
func (FRateSnapshot) SetRateName(string) FRate {
	panic("SetRateName called on a RateSnapshot")
}

// Clear panics.
func (FRateSnapshot) Clear() (float64, float64) {
	panic("Clear called on a FRateSnapshot")
}

// Update panics.
func (FRateSnapshot) Update(float64) {
	panic("Update called on a FRateSnapshot")
}

// Update panics.
func (FRateSnapshot) UpdateTs(float64, int64) {
	panic("Update called on a FRateSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g *FRateSnapshot) Values() (float64, float64) { return g.value, g.FRate }

// NilFRate is a no-op FRate.
type NilFRate struct{}

func (NilFRate) Name() string {
	return ""
}

func (NilFRate) RateName() string {
	return ""
}

func (g NilFRate) SetName(string) FRate { return g }

func (g NilFRate) SetRateName(string) FRate { return g }

// Snapshot is a no-op.
func (NilFRate) Snapshot() FRate { return NilFRate{} }

// Clear is a no-op.
func (NilFRate) Clear() (float64, float64) { return 0, 0 }

// Update is a no-op.
func (NilFRate) Update(float64) {}

// UpdateTs is a no-op.
func (NilFRate) UpdateTs(float64, int64) {}

// Value is a no-op.
func (NilFRate) Values() (float64, float64) { return 0, 0 }

// StandardFRate is the standard implementation of a FRate and uses the
// sync/atomic package to manage a single int64 value.
type StandardFRate struct {
	name     string
	rateName string
	prev     float64
	prevTs   int64
	value    float64
	valueTs  int64
	mutex    sync.Mutex
}

// Snapshot returns a read-only copy of the FRate.
func (g *StandardFRate) Snapshot() FRate {
	v, rate := g.Values()
	return newFRateSnapshot(v, rate, g.name, g.rateName)
}

// Clear sets the DownCounter to zero.
func (g *StandardFRate) Clear() (float64, float64) {
	var (
		v float64
		p float64
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
		// first or immediate try
		return v, 0
	}
	return v, 1e9 * float64(v-p) / float64(d)
}

// Update updates the FRate's value.
func (g *StandardFRate) Update(v float64) {
	ts := time.Now().UnixNano()
	g.mutex.Lock()
	g.prev = g.value
	g.prevTs = g.valueTs
	g.value = v
	g.valueTs = ts
	g.mutex.Unlock()
}

// UpdateTs updates the FRate's value.
func (g *StandardFRate) UpdateTs(v float64, ts int64) {
	g.mutex.Lock()
	g.prev = g.value
	g.prevTs = g.valueTs
	g.value = v
	g.valueTs = ts
	g.mutex.Unlock()
}

// Value returns the FRate's current value.
func (g *StandardFRate) Values() (float64, float64) {
	var (
		v float64
		p float64
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

func (g *StandardFRate) SetName(name string) FRate {
	g.mutex.Lock()
	g.name = name
	g.mutex.Unlock()
	return g
}

func (g *StandardFRate) Name() string {
	return g.name
}

func (g *StandardFRate) SetRateName(rateName string) FRate {
	g.mutex.Lock()
	g.rateName = rateName
	g.mutex.Unlock()
	return g
}

func (g *StandardFRate) RateName() string {
	return g.rateName
}
