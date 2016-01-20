package mstats

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// Var is an atomic variable satisfying expvar.Var
type Var struct {
	atomic.Value
}

func (a *Var) String() string {
	v := a.Load().(uint64)
	return fmt.Sprintf("%d", v)
}

// PauseNS is the total number of nanoseconds the GC has paused the application
var PauseNS Var

// NumGC is the number of collections
var NumGC Var

// Alloc is the number of bytes allocated and not yet freed by the application
var Alloc Var

// Start polls runtime.ReadMemStats with interval d and updates the package level variables
func Start(d time.Duration) {
	for range time.Tick(d) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		PauseNS.Store(m.PauseTotalNs)
		Alloc.Store(m.Alloc)
		NumGC.Store(uint64(m.NumGC))
	}
}
