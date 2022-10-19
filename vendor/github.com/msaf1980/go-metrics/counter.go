package metrics

import "sync/atomic"

// Counters hold an uint64 value that can be incremented only
//
//	Graphite naming scheme
//
// Plain: {PREFIX}.{NAME}
//
// Tagged: {TAG_PREFIX}.{NAME}
type Counter interface {
	Clear() uint64
	Count() uint64
	Add(uint64)
	Snapshot() Counter
}

// GetOrRegisterCounter returns an existing Counter or constructs and registers
// a new StandardCounter.
func GetOrRegisterCounter(name string, r Registry) Counter {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewCounter).(Counter)
}

// GetOrRegisterCounterT returns an existing Counter or constructs and registers
// a new StandardCounter.
func GetOrRegisterCounterT(name string, tagsMap map[string]string, r Registry) Counter {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewCounter).(Counter)
}

// NewCounter constructs a new StandardCounter.
func NewCounter() Counter {
	if UseNilMetrics {
		return NilCounter{}
	}
	return &StandardCounter{0}
}

// NewRegisteredCounter constructs and registers a new StandardCounter.
func NewRegisteredCounter(name string, r Registry) Counter {
	c := NewCounter()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredCounterT constructs and registers a new StandardCounter.
func NewRegisteredCounterT(name string, tagsMap map[string]string, r Registry) Counter {
	c := NewCounter()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// CounterSnapshot is a read-only copy of another Counter.
type CounterSnapshot uint64

// Clear panics.
func (CounterSnapshot) Clear() uint64 {
	panic("Clear called on a CounterSnapshot")
}

// Count returns the count at the time the snapshot was taken.
func (c CounterSnapshot) Count() uint64 { return uint64(c) }

// Inc panics.
func (CounterSnapshot) Add(uint64) {
	panic("Inc called on a CounterSnapshot")
}

// Snapshot returns the snapshot.
func (c CounterSnapshot) Snapshot() Counter { return c }

// NilCounter is a no-op Counter.
type NilCounter struct{}

// Clear is a no-op.
func (NilCounter) Clear() uint64 { return 0 }

// Count is a no-op.
func (NilCounter) Count() uint64 { return 0 }

// Inc is a no-op.
func (NilCounter) Add(i uint64) {}

// Snapshot is a no-op.
func (NilCounter) Snapshot() Counter { return NilCounter{} }

// StandardCounter is the standard implementation of a Counter and uses the
// sync/atomic package to manage a single uint64 value.
type StandardCounter struct {
	count uint64
}

// Clear sets the Counter to zero.
func (c *StandardCounter) Clear() uint64 {
	return atomic.SwapUint64(&c.count, 0)
}

// Count returns the current count.
func (c *StandardCounter) Count() uint64 {
	return atomic.LoadUint64(&c.count)
}

// Inc increments the Counter by the given amount.
func (c *StandardCounter) Add(i uint64) {
	atomic.AddUint64(&c.count, i)
}

// Snapshot returns a read-only copy of the Counter.
func (c *StandardCounter) Snapshot() Counter {
	return CounterSnapshot(c.Count())
}
