package metrics

import (
	"sync"
)

// GetOrRegisterDiffer returns an existing Differ or constructs and registers a
// new StandardDiffer.
func GetOrRegisterDiffer(name string, r Registry, d int64) Gauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewDiffer(d)
	}).(Gauge)
}

// GetOrRegisterDifferT returns an existing Differ or constructs and registers a
// new StandardDiffer.
func GetOrRegisterDifferT(name string, tagsMap map[string]string, r Registry, d int64) Gauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewDiffer(d)
	}).(Gauge)
}

// NewDiffer constructs a new StandardDiffer.
func NewDiffer(d int64) Gauge {
	if UseNilMetrics {
		return NilGauge{}
	}
	return &StandardDiffer{prev: d, value: d}
}

// NewRegisteredDiffer constructs and registers a new StandardDiffer.
func NewRegisteredDiffer(name string, r Registry, d int64) Gauge {
	c := NewDiffer(d)
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredDifferT constructs and registers a new StandardDiffer.
func NewRegisteredDifferT(name string, tagsMap map[string]string, r Registry, d int64) Gauge {
	c := NewDiffer(d)
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// StandardDiffer is the standard implementation of a Differ and uses the
// sync/atomic package to manage a single int64 value.
type StandardDiffer struct {
	prev  int64
	value int64
	mutex sync.Mutex
}

// Snapshot returns a read-only copy of the Differ.
func (g *StandardDiffer) Snapshot() Gauge {
	return GaugeSnapshot(g.Value())
}

// Clear sets the DownCounter to zero.
func (g *StandardDiffer) Clear() int64 {
	g.mutex.Lock()
	v := g.value - g.prev
	g.value = 0
	g.prev = 0
	g.mutex.Unlock()
	return v
}

// Update updates the Differ's value.
func (g *StandardDiffer) Update(v int64) {
	g.mutex.Lock()
	g.prev = g.value
	g.value = v
	g.mutex.Unlock()
}

// Value returns the Differ's current value.
func (g *StandardDiffer) Value() int64 {
	g.mutex.Lock()
	v := g.value - g.prev
	g.mutex.Unlock()
	return v
}
