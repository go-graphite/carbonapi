package metrics

import (
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	RuntimeNames struct {
		MemStats struct {
			Alloc         string
			BuckHashSys   string
			Frees         string
			HeapAlloc     string
			HeapIdle      string
			HeapInUse     string
			HeapObjects   string
			HeapReleased  string
			HeapSys       string
			LastGC        string
			Lookups       string
			Mallocs       string
			MCacheInUse   string
			MCacheSys     string
			MSpanInuse    string
			MSpanSys      string
			NextGC        string
			NumGC         string
			GCCPUFraction string
			// PauseNs       Histogram
			PauseTotalNs string
			StackInUse   string
			StackSys     string
			Sys          string
			TotalAlloc   string
		}
		NumCgoCall   string
		NumGoroutine string
		NumThread    string
	}
	memStats       runtime.MemStats
	runtimeMetrics struct {
		MemStats struct {
			Alloc       Gauge
			BuckHashSys Gauge
			// DebugGC       Gauge
			// EnableGC      Gauge
			Frees         Rate
			HeapAlloc     Gauge
			HeapIdle      Gauge
			HeapInUse     Gauge
			HeapObjects   Gauge
			HeapReleased  Gauge
			HeapSys       Gauge
			LastGC        Gauge
			Lookups       Rate
			Mallocs       Rate
			MCacheInUse   Gauge
			MCacheSys     Gauge
			MSpanInUse    Gauge
			MSpanSys      Gauge
			NextGC        Gauge
			NumGC         Rate
			GCCPUFraction FGauge
			// PauseNs       Histogram
			PauseTotalNs Gauge
			StackInUse   Gauge
			StackSys     Gauge
			Sys          Gauge
			TotalAlloc   Gauge
		}
		NumCgoCall   Rate
		NumGoroutine Gauge
		NumThread    Gauge
		// ReadMemStats Timer
	}
	// frees       uint64
	// lookups     uint64
	// mallocs     uint64
	// numGC       uint32
	// numCgoCalls int64

	threadCreateProfile        = pprof.Lookup("threadcreate")
	registerRuntimeMetricsOnce = sync.Once{}
)

func init() {
	RuntimeNames.MemStats.Alloc = "runtime.mem_stats.alloc_bytes"
	RuntimeNames.MemStats.BuckHashSys = "runtime.mem_stats.buck_hash_sys_bytes"
	RuntimeNames.MemStats.Frees = "runtime.mem_stats.frees"
	RuntimeNames.MemStats.HeapAlloc = "runtime.mem_stats.heap_alloc_bytes"
	RuntimeNames.MemStats.HeapIdle = "runtime.mem_stats.heap_idle_bytes"
	RuntimeNames.MemStats.HeapInUse = "runtime.mem_stats.heap_inuse_bytes"
	RuntimeNames.MemStats.HeapObjects = "runtime.mem_stats.heap_objects"
	RuntimeNames.MemStats.HeapReleased = "runtime.mem_stats.heap_released_bytes"
	RuntimeNames.MemStats.HeapSys = "runtime.mem_stats.heap_sys_bytes"
	RuntimeNames.MemStats.LastGC = "runtime.mem_stats.last_gc"
	RuntimeNames.MemStats.Lookups = "runtime.mem_stats.lookups"
	RuntimeNames.MemStats.Mallocs = "runtime.mem_stats.mallocs"
	RuntimeNames.MemStats.MCacheInUse = "runtime.mem_stats.mcache_inuse"
	RuntimeNames.MemStats.MCacheSys = "runtime.mem_stats.mcache_sys"
	RuntimeNames.MemStats.MSpanInuse = "runtime.mem_stats.mspan_inuse"
	RuntimeNames.MemStats.MSpanSys = "runtime.mem_stats.mspan_sys"
	RuntimeNames.MemStats.NextGC = "runtime.mem_stats.next_gc"
	RuntimeNames.MemStats.NumGC = "runtime.mem_stats.num_gc"
	RuntimeNames.MemStats.GCCPUFraction = "runtime.mem_stats.gcccpu_fraction"
	// RuntimeNames.MemStats.PauseNs = r.Register("runtime.mem_stats.pause_ns"
	RuntimeNames.MemStats.PauseTotalNs = "runtime.mem_stats.pause_total_ns"
	RuntimeNames.MemStats.StackInUse = "runtime.mem_stats.stack_in_use_bytes"
	RuntimeNames.MemStats.StackSys = "runtime.mem_stats.stack_sys_bytes"
	RuntimeNames.MemStats.Sys = "runtime.mem_stats.sys"
	RuntimeNames.MemStats.TotalAlloc = "runtime.mem_stats.total_alloc_bytes"

	RuntimeNames.NumCgoCall = "runtime.num_cgo_call"
	RuntimeNames.NumGoroutine = "runtime.num_goroutine"
	RuntimeNames.NumThread = "runtime.num_thread"
}

// Capture new values for the Go runtime statistics exported in
// runtime.MemStats.  This is designed to be called as a goroutine.
func CaptureRuntimeMemStats(d time.Duration) {
	for range time.Tick(d) {
		CaptureRuntimeMemStatsOnce()
	}
}

// Capture new values for the Go runtime statistics exported in
// runtime.MemStats.  This is designed to be called in a background
// goroutine.  Giving a registry which has not been given to
// RegisterRuntimeMemStats will panic.
//
// Be very careful with this because runtime.ReadMemStats calls the C
// functions runtime·semacquire(&runtime·worldsema) and runtime·stoptheworld()
// and that last one does what it says on the tin.
func CaptureRuntimeMemStatsOnce() {
	t := time.Now().UnixNano()
	runtime.ReadMemStats(&memStats) // This takes 50-200us.
	// runtimeMetrics.ReadMemStats.UpdateSince(t)

	runtimeMetrics.MemStats.Alloc.Update(int64(memStats.Alloc))
	runtimeMetrics.MemStats.BuckHashSys.Update(int64(memStats.BuckHashSys))
	// if memStats.DebugGC {
	// 	runtimeMetrics.MemStats.DebugGC.Update(1)
	// } else {
	// 	runtimeMetrics.MemStats.DebugGC.Update(0)
	// }
	// if memStats.EnableGC {
	// 	runtimeMetrics.MemStats.EnableGC.Update(1)
	// } else {
	// 	runtimeMetrics.MemStats.EnableGC.Update(0)
	// }

	runtimeMetrics.MemStats.Frees.UpdateTs(int64(memStats.Frees), t)
	runtimeMetrics.MemStats.HeapAlloc.Update(int64(memStats.HeapAlloc))
	runtimeMetrics.MemStats.HeapIdle.Update(int64(memStats.HeapIdle))
	runtimeMetrics.MemStats.HeapInUse.Update(int64(memStats.HeapInuse))
	runtimeMetrics.MemStats.HeapObjects.Update(int64(memStats.HeapObjects))
	runtimeMetrics.MemStats.HeapReleased.Update(int64(memStats.HeapReleased))
	runtimeMetrics.MemStats.HeapSys.Update(int64(memStats.HeapSys))
	runtimeMetrics.MemStats.LastGC.Update(int64(memStats.LastGC))
	runtimeMetrics.MemStats.Lookups.UpdateTs(int64(memStats.Lookups), t)
	runtimeMetrics.MemStats.Mallocs.UpdateTs(int64(memStats.Mallocs), t)
	runtimeMetrics.MemStats.MCacheInUse.Update(int64(memStats.MCacheInuse))
	runtimeMetrics.MemStats.MCacheSys.Update(int64(memStats.MCacheSys))
	runtimeMetrics.MemStats.MSpanInUse.Update(int64(memStats.MSpanInuse))
	runtimeMetrics.MemStats.MSpanSys.Update(int64(memStats.MSpanSys))
	runtimeMetrics.MemStats.NextGC.Update(int64(memStats.NextGC))
	runtimeMetrics.MemStats.NumGC.UpdateTs(int64(memStats.NumGC), t)
	runtimeMetrics.MemStats.GCCPUFraction.Update(gcCPUFraction(&memStats))

	// <https://code.google.com/p/go/source/browse/src/pkg/runtime/mgc0.c>
	// i := numGC % uint32(len(memStats.PauseNs))
	// ii := memStats.NumGC % uint32(len(memStats.PauseNs))
	// if memStats.NumGC-numGC >= uint32(len(memStats.PauseNs)) {
	// 	for i = 0; i < uint32(len(memStats.PauseNs)); i++ {
	// 		runtimeMetrics.MemStats.PauseNs.Update(int64(memStats.PauseNs[i]))
	// 	}
	// } else {
	// 	if i > ii {
	// 		for ; i < uint32(len(memStats.PauseNs)); i++ {
	// 			runtimeMetrics.MemStats.PauseNs.Update(int64(memStats.PauseNs[i]))
	// 		}
	// 		i = 0
	// 	}
	// 	for ; i < ii; i++ {
	// 		runtimeMetrics.MemStats.PauseNs.Update(int64(memStats.PauseNs[i]))
	// 	}
	// }
	// frees = memStats.Frees
	// lookups = memStats.Lookups
	// mallocs = memStats.Mallocs
	// numGC = memStats.NumGC

	runtimeMetrics.MemStats.PauseTotalNs.Update(int64(memStats.PauseTotalNs))
	runtimeMetrics.MemStats.StackInUse.Update(int64(memStats.StackInuse))
	runtimeMetrics.MemStats.StackSys.Update(int64(memStats.StackSys))
	runtimeMetrics.MemStats.Sys.Update(int64(memStats.Sys))
	runtimeMetrics.MemStats.TotalAlloc.Update(int64(memStats.TotalAlloc))

	runtimeMetrics.NumCgoCall.UpdateTs(int64(numCgoCall()), t)

	runtimeMetrics.NumGoroutine.Update(int64(runtime.NumGoroutine()))

	runtimeMetrics.NumThread.Update(int64(threadCreateProfile.Count()))
}

// Register runtimeMetrics for the Go runtime statistics exported in runtime and
// specifically runtime.MemStats.  The runtimeMetrics are named by their
// fully-qualified Go symbols, i.e. runtime.MemStats.Alloc.
func RegisterRuntimeMemStats(r Registry) {
	registerRuntimeMetricsOnce.Do(func() {
		if nil == r {
			r = DefaultRegistry
		}

		runtimeMetrics.MemStats.Alloc = NewGauge()
		runtimeMetrics.MemStats.BuckHashSys = NewGauge()
		// runtimeMetrics.MemStats.DebugGC = NewGauge()
		// runtimeMetrics.MemStats.EnableGC = NewGauge()
		// runtimeMetrics.MemStats.Frees = NewDiffer(int64(memStats.Frees))
		runtimeMetrics.MemStats.Frees = NewRate()
		runtimeMetrics.MemStats.HeapAlloc = NewGauge()
		runtimeMetrics.MemStats.HeapIdle = NewGauge()
		runtimeMetrics.MemStats.HeapInUse = NewGauge()
		runtimeMetrics.MemStats.HeapObjects = NewGauge()
		runtimeMetrics.MemStats.HeapReleased = NewGauge()
		runtimeMetrics.MemStats.HeapSys = NewGauge()
		runtimeMetrics.MemStats.LastGC = NewGauge()
		// runtimeMetrics.MemStats.Lookups = NewDiffer(int64(memStats.Lookups))
		runtimeMetrics.MemStats.Lookups = NewRate()
		// runtimeMetrics.MemStats.Mallocs = NewDiffer(int64(memStats.Mallocs))
		runtimeMetrics.MemStats.Mallocs = NewRate()
		runtimeMetrics.MemStats.MCacheInUse = NewGauge()
		runtimeMetrics.MemStats.MCacheSys = NewGauge()
		runtimeMetrics.MemStats.MSpanInUse = NewGauge()
		runtimeMetrics.MemStats.MSpanSys = NewGauge()
		runtimeMetrics.MemStats.NextGC = NewGauge()
		// runtimeMetrics.MemStats.NumGC = NewDiffer(int64(memStats.NextGC))
		runtimeMetrics.MemStats.NumGC = NewRate()
		runtimeMetrics.MemStats.GCCPUFraction = NewFGauge()
		// runtimeMetrics.MemStats.PauseNs = NewHistogram(NewExpDecaySample(1028, 0.015))
		runtimeMetrics.MemStats.PauseTotalNs = NewGauge()
		runtimeMetrics.MemStats.StackInUse = NewGauge()
		runtimeMetrics.MemStats.StackSys = NewGauge()
		runtimeMetrics.MemStats.Sys = NewGauge()
		runtimeMetrics.MemStats.TotalAlloc = NewGauge()
		// runtimeMetrics.NumCgoCall = NewDiffer(numCgoCall())
		runtimeMetrics.NumCgoCall = NewRate()
		runtimeMetrics.NumGoroutine = NewGauge()
		runtimeMetrics.NumThread = NewGauge()
		// runtimeMetrics.ReadMemStats = NewTimer()
		r.Register(RuntimeNames.MemStats.Alloc, runtimeMetrics.MemStats.Alloc)
		r.Register(RuntimeNames.MemStats.BuckHashSys, runtimeMetrics.MemStats.BuckHashSys)
		// r.Register("runtime.mem_stats.DebugGC", runtimeMetrics.MemStats.DebugGC)
		// r.Register("runtime.mem_stats.EnableGC", runtimeMetrics.MemStats.EnableGC)
		r.Register(RuntimeNames.MemStats.Frees, runtimeMetrics.MemStats.Frees)
		r.Register(RuntimeNames.MemStats.HeapAlloc, runtimeMetrics.MemStats.HeapAlloc)
		r.Register(RuntimeNames.MemStats.HeapIdle, runtimeMetrics.MemStats.HeapIdle)
		r.Register(RuntimeNames.MemStats.HeapInUse, runtimeMetrics.MemStats.HeapInUse)
		r.Register(RuntimeNames.MemStats.HeapObjects, runtimeMetrics.MemStats.HeapObjects)
		r.Register(RuntimeNames.MemStats.HeapReleased, runtimeMetrics.MemStats.HeapReleased)
		r.Register(RuntimeNames.MemStats.HeapSys, runtimeMetrics.MemStats.HeapSys)
		r.Register(RuntimeNames.MemStats.LastGC, runtimeMetrics.MemStats.LastGC)
		r.Register(RuntimeNames.MemStats.Lookups, runtimeMetrics.MemStats.Lookups)
		r.Register(RuntimeNames.MemStats.Mallocs, runtimeMetrics.MemStats.Mallocs)
		r.Register(RuntimeNames.MemStats.MCacheInUse, runtimeMetrics.MemStats.MCacheInUse)
		r.Register(RuntimeNames.MemStats.MCacheSys, runtimeMetrics.MemStats.MCacheSys)
		r.Register(RuntimeNames.MemStats.MCacheInUse, runtimeMetrics.MemStats.MSpanInUse)
		r.Register(RuntimeNames.MemStats.MCacheSys, runtimeMetrics.MemStats.MSpanSys)
		r.Register(RuntimeNames.MemStats.NextGC, runtimeMetrics.MemStats.NextGC)
		r.Register(RuntimeNames.MemStats.NumGC, runtimeMetrics.MemStats.NumGC)
		r.Register(RuntimeNames.MemStats.GCCPUFraction, runtimeMetrics.MemStats.GCCPUFraction)
		// r.Register("runtime.mem_stats.pause_ns", runtimeMetrics.MemStats.PauseNs)
		r.Register(RuntimeNames.MemStats.PauseTotalNs, runtimeMetrics.MemStats.PauseTotalNs)
		r.Register(RuntimeNames.MemStats.StackInUse, runtimeMetrics.MemStats.StackInUse)
		r.Register(RuntimeNames.MemStats.StackSys, runtimeMetrics.MemStats.StackSys)
		r.Register(RuntimeNames.MemStats.Sys, runtimeMetrics.MemStats.Sys)
		r.Register(RuntimeNames.MemStats.TotalAlloc, runtimeMetrics.MemStats.TotalAlloc)
		r.Register(RuntimeNames.NumCgoCall, runtimeMetrics.NumCgoCall)
		r.Register(RuntimeNames.NumGoroutine, runtimeMetrics.NumGoroutine)
		r.Register(RuntimeNames.NumThread, runtimeMetrics.NumThread)
		// r.Register("runtime.read_mem_stats", runtimeMetrics.ReadMemStats)
	})
}
