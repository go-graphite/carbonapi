package metrics

import (
	"math"
	"sync/atomic"
)

// FGauges hold a float64 value that can be set arbitrarily.
//
// Plain: {PREFIX}.{NAME}
//
// Tagged: {TAG_PREFIX}.{NAME}
type FGauge interface {
	Snapshot() FGauge
	Update(float64)
	Value() float64
}

// GetOrRegisterFGauge returns an existing FGauge or constructs and registers a
// new StandardFGauge.
func GetOrRegisterFGauge(name string, r Registry) FGauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewFGauge()).(FGauge)
}

// GetOrRegisterFGaugeT returns an existing FGauge or constructs and registers a
// new StandardFGauge.
func GetOrRegisterFGaugeT(name string, tagsMap map[string]string, r Registry) FGauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewFGauge()).(FGauge)
}

// NewFGauge constructs a new StandardFGauge.
func NewFGauge() FGauge {
	if UseNilMetrics {
		return NilFGauge{}
	}
	return &StandardFGauge{
		value: 0.0,
	}
}

// NewRegisteredFGauge constructs and registers a new StandardFGauge.
func NewRegisteredFGauge(name string, r Registry) FGauge {
	c := NewFGauge()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredFGaugeT constructs and registers a new StandardFGauge.
func NewRegisteredFGaugeT(name string, tagsMap map[string]string, r Registry) FGauge {
	c := NewFGauge()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// NewFunctionalGauge constructs a new FunctionalGauge.
func NewFunctionalFGauge(f func() float64) FGauge {
	if UseNilMetrics {
		return NilFGauge{}
	}
	return &FunctionalFGauge{value: f}
}

// NewRegisteredFunctionalGauge constructs and registers a new StandardGauge.
func NewRegisteredFunctionalFGauge(name string, r Registry, f func() float64) FGauge {
	c := NewFunctionalFGauge(f)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredFunctionalGaugeT constructs and registers a new StandardGauge.
func NewRegisteredFunctionalFGaugeT(name string, tagsMap map[string]string, r Registry, f func() float64) FGauge {
	c := NewFunctionalFGauge(f)
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// FGaugeSnapshot is a read-only copy of another FGauge.
type FGaugeSnapshot float64

// Snapshot returns the snapshot.
func (g FGaugeSnapshot) Snapshot() FGauge { return g }

// Update panics.
func (FGaugeSnapshot) Update(float64) {
	panic("Update called on a FGaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g FGaugeSnapshot) Value() float64 { return float64(g) }

// NilGauge is a no-op Gauge.
type NilFGauge struct{}

// Snapshot is a no-op.
func (NilFGauge) Snapshot() FGauge { return NilFGauge{} }

// Update is a no-op.
func (NilFGauge) Update(v float64) {}

// Value is a no-op.
func (NilFGauge) Value() float64 { return 0.0 }

// StandardFGauge is the standard implementation of a FGauge and uses
// sync.Mutex to manage a single float64 value.
type StandardFGauge struct {
	value uint64
}

// Snapshot returns a read-only copy of the gauge.
func (g *StandardFGauge) Snapshot() FGauge {
	return FGaugeSnapshot(g.Value())
}

// Update updates the gauge's value.
func (g *StandardFGauge) Update(v float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(v))
}

// Value returns the gauge's current value.
func (g *StandardFGauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.value))
}

// FunctionalFGauge returns value from given function
type FunctionalFGauge struct {
	value func() float64
}

// Value returns the gauge's current value.
func (g FunctionalFGauge) Value() float64 {
	return g.value()
}

// Snapshot returns the snapshot.
func (g FunctionalFGauge) Snapshot() FGauge { return FGaugeSnapshot(g.Value()) }

// Update panics.
func (FunctionalFGauge) Update(float64) {
	panic("Update called on a FunctionalFGauge")
}
