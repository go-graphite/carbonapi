package metrics

import "sync/atomic"

// UGauges hold an int64 value that can be set arbitrarily.
//
//	Graphite naming scheme
//
// Plain: {PREFIX}.{NAME}
//
// Tagged: {TAG_PREFIX}.{NAME}
type UGauge interface {
	Snapshot() UGauge
	Clear() uint64
	Update(uint64)
	Value() uint64
}

// GetOrRegisterUGauge returns an existing UGauge or constructs and registers a
// new StandardUGauge.
func GetOrRegisterUGauge(name string, r Registry) UGauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewUGauge).(UGauge)
}

// GetOrRegisterUGaugeT returns an existing UGauge or constructs and registers a
// new StandardUGauge.
func GetOrRegisterUGaugeT(name string, tagsMap map[string]string, r Registry) UGauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewUGauge).(UGauge)
}

// NewUGauge constructs a new StandardUGauge.
func NewUGauge() UGauge {
	if UseNilMetrics {
		return NilUGauge{}
	}
	return &StandardUGauge{0}
}

// NewRegisteredUGauge constructs and registers a new StandardUGauge.
func NewRegisteredUGauge(name string, r Registry) UGauge {
	c := NewUGauge()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredUGaugeT constructs and registers a new StandardUGauge.
func NewRegisteredUGaugeT(name string, tagsMap map[string]string, r Registry) UGauge {
	c := NewUGauge()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// NewFunctionalUGauge constructs a new FunctionalUGauge.
func NewFunctionalUGauge(f func() uint64) UGauge {
	if UseNilMetrics {
		return NilUGauge{}
	}
	return &FunctionalUGauge{value: f}
}

// NewRegisteredFunctionalUGauge constructs and registers a new StandardUGauge.
func NewRegisteredFunctionalUGauge(name string, r Registry, f func() uint64) UGauge {
	c := NewFunctionalUGauge(f)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredFunctionalUGaugeT constructs and registers a new StandardUGauge.
func NewRegisteredFunctionalUGaugeT(name string, tagsMap map[string]string, r Registry, f func() uint64) UGauge {
	c := NewFunctionalUGauge(f)
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// UGaugeSnapshot is a read-only copy of another UGauge.
type UGaugeSnapshot uint64

// Snapshot returns the snapshot.
func (g UGaugeSnapshot) Snapshot() UGauge { return g }

// Clear panics.
func (g UGaugeSnapshot) Clear() uint64 {
	panic("Clear called on a UGaugeSnapshot")
}

// Update panics.
func (UGaugeSnapshot) Update(uint64) {
	panic("Update called on a UGaugeSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g UGaugeSnapshot) Value() uint64 { return uint64(g) }

// NilUGauge is a no-op UGauge.
type NilUGauge struct{}

// Snapshot is a no-op.
func (NilUGauge) Snapshot() UGauge { return NilUGauge{} }

// Clear is a no-op.
func (NilUGauge) Clear() uint64 { return 0 }

// Update is a no-op.
func (NilUGauge) Update(v uint64) {}

// Value is a no-op.
func (NilUGauge) Value() uint64 { return 0 }

// StandardUGauge is the standard implementation of a UGauge and uses the
// sync/atomic package to manage a single int64 value.
type StandardUGauge struct {
	value uint64
}

// Snapshot returns a read-only copy of the gauge.
func (g *StandardUGauge) Snapshot() UGauge {
	return UGaugeSnapshot(g.Value())
}

// Clear sets the DownCounter to zero.
func (g *StandardUGauge) Clear() uint64 {
	return atomic.SwapUint64(&g.value, 0)
}

// Update updates the gauge's value.
func (g *StandardUGauge) Update(v uint64) {
	atomic.StoreUint64(&g.value, v)
}

// Value returns the gauge's current value.
func (g *StandardUGauge) Value() uint64 {
	return atomic.LoadUint64(&g.value)
}

// FunctionalUGauge returns value from given function
type FunctionalUGauge struct {
	value func() uint64
}

// Value returns the gauge's current value.
func (g FunctionalUGauge) Value() uint64 {
	return g.value()
}

// Snapshot returns the snapshot.
func (g FunctionalUGauge) Snapshot() UGauge { return UGaugeSnapshot(g.Value()) }

// Update panics.
func (FunctionalUGauge) Update(uint64) {
	panic("Update called on a FunctionalUGauge")
}

// Clear panics.
func (g FunctionalUGauge) Clear() uint64 {
	panic("Clear called on a FunctionalUGauge")
}
