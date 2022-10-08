package metrics

import (
	"sync"
	"time"
)

// Updated defines the metrics which need to be async updated with Tick.
type Updated interface {
	Tick()
}

// meterUpdater ticks meters every 5s from a single goroutine.
// meters are references in a set for future stopping.
type meterUpdater struct {
	sync.RWMutex
	started bool
	meters  map[Updated]struct{}
	ticker  *time.Ticker
	close   chan struct{}
}

var updater = meterUpdater{
	meters: make(map[Updated]struct{}),
	close:  make(chan struct{}),
}

func (ma *meterUpdater) Register(m Updated) {
	ma.Lock()
	defer ma.Unlock()

	ma.meters[m] = struct{}{}

	if !ma.started {
		ma.ticker = time.NewTicker(5e9)
		ma.started = true
		go ma.start()
	}
}

func (ma *meterUpdater) Unregister(m Updated) {
	ma.Lock()
	delete(ma.meters, m)
	if len(ma.meters) == 0 {
		ma.stop()
	}
	ma.Unlock()
}

func (ma *meterUpdater) stop() {
	if ma.started {
		ma.close <- struct{}{}
		ma.started = false
	}
}

func (ma *meterUpdater) Stop() {
	ma.RLock()
	defer ma.RUnlock()
	ma.stop()
}

func (ma *meterUpdater) StopIfEmpty() {
	ma.RLock()
	defer ma.RUnlock()
	if len(ma.meters) == 0 {
		ma.stop()
	}
}

// Ticks meters on the scheduled interval
func (ma *meterUpdater) start() {
	for {
		select {
		case <-ma.close:
			return
		case <-ma.ticker.C:
			ma.tickMeters()
		}
	}
}

func (ma *meterUpdater) tickMeters() {
	ma.RLock()
	defer ma.RUnlock()
	for meter := range ma.meters {
		meter.Tick()
	}
}
