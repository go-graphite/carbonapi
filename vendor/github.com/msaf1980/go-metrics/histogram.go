package metrics

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// A HistogramInterface is some strped (no Weights{}, it's not need in registry Each iterator) version of Histogram interface
//
//	Graphite naming scheme
//
// Plain:
//
// {PREFIX}.{NAME}{LABEL_BUCKET1}
//
// {PREFIX}.{NAME}{LABEL_BUCKET2}
//
// {PREFIX}.{NAME}{LABEL_BUCKET_INF}
//
// {PREFIX}.{NAME}{TOTAL}
//
// Tagged:
//
// {TAG_PREFIX}.{NAME}{LABEL_BUCKET1};TAG=VAL;..;le=W1
//
// {TAG_PREFIX}.{NAME}{LABEL_BUCKET2};TAG=VAL;..;le=W2
//
// {TAG_PREFIX}.{NAME}{LABEL_BUCKET_INF};TAG=VAL;..;le=inf
//
// {TAG_PREFIX}{NAME}{TOTAL};TAG=VAL;..
type HistogramInterface interface {
	Clear() []uint64
	Values() []uint64
	Labels() []string
	NameTotal() string
	// Tag aliases values (for le key)
	WeightsAliases() []string
	// If true, is prometheus-like (cummulative, increment in bucket[1]  also increment bucket[0])
	IsSummed() bool
}

// A Histogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type Histogram interface {
	HistogramInterface
	SetLabels([]string) Histogram
	AddLabelPrefix(string) Histogram
	SetNameTotal(string) Histogram
	Snapshot() Histogram
	Add(v int64)
	Weights() []int64
}

// GetOrRegisterHistogram returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFixedHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedHistogram(startVal, endVal, width)
	}).(Histogram)
}

// GetOrRegisterHistogramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedHistogram(startVal, endVal, width)
	}).(Histogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredFixedHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVHistogram(weights, names)
	}).(Histogram)
}

func GetOrRegisterVHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVHistogram(weights, names)
	}).(Histogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredVHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredVHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type HistogramSnapshot struct {
	weights        []int64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *HistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *HistogramSnapshot) Labels() []string {
	return h.names
}

func (HistogramSnapshot) SetLabels([]string) Histogram {
	panic("SetLabels called on a HistogramSnapshot")
}

func (HistogramSnapshot) AddLabelPrefix(string) Histogram {
	panic("AddLabelPrefix called on a HistogramSnapshot")
}
func (HistogramSnapshot) SetNameTotal(string) Histogram {
	panic("SetNameTotal called on a HistogramSnapshot")
}

func (h *HistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *HistogramSnapshot) Weights() []int64 {
	return h.weights
}

func (h *HistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *HistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *HistogramSnapshot) Add(v int64) {
	panic("Add called on a HistogramSnapshot")
}

func (h *HistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a HistogramSnapshot")
}

func (h *HistogramSnapshot) Snapshot() Histogram {
	return h
}

func (HistogramSnapshot) IsSummed() bool { return false }

type HistogramStorage struct {
	weights        []int64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	labels         []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *HistogramStorage) Labels() []string {
	return h.labels
}

func (h *HistogramStorage) SetLabels(labels []string) {
	h.lock.Lock()
	for i := 0; i < Min(len(h.labels), len(labels)); i++ {
		h.labels[i] = labels[i]
	}
	h.lock.Unlock()
}

func (h *HistogramStorage) AddLabelPrefix(labelPrefix string) {
	h.lock.Lock()
	for i := range h.labels {
		if strings.HasPrefix(h.labels[i], ".") {
			h.labels[i] = "." + labelPrefix + h.labels[i][1:]
		} else {
			h.labels[i] = labelPrefix + h.labels[i]
		}
	}
	h.lock.Unlock()
}

func (h *HistogramStorage) SetNameTotal(total string) {
	h.lock.Lock()
	h.total = total
	h.lock.Unlock()
}

func (h *HistogramStorage) NameTotal() string {
	return h.total
}

func (h *HistogramStorage) Weights() []int64 {
	return h.weights
}

func (h *HistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *HistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *HistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *HistogramStorage) Snapshot() Histogram {
	return &HistogramSnapshot{
		names:          h.labels,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *HistogramStorage) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *HistogramStorage) IsSummed() bool { return false }

// A FixedHistogram is implementation of Histogram with fixed-size buckets.
type FixedHistogram struct {
	HistogramStorage
	start int64
	width int64
}

func NewFixedHistogram(startVal, endVal, width int64) *FixedHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	if width < 0 {
		width = -width
	}
	n := endVal - startVal

	count := n/width + 2
	if n%width != 0 {
		count++
	}
	weights := make([]int64, count)
	weightsAliases := make([]string, count)
	labels := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal
	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(endVal+width, 10)))
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = math.MaxInt64
			weightsAliases[i] = "inf"
			labels[i] = ".inf"
		} else {
			weights[i] = ge
			weightsAliases[i] = strconv.FormatInt(ge, 10)
			labels[i] = "." + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, ge)
			ge += width
		}
	}

	return &FixedHistogram{
		HistogramStorage: HistogramStorage{
			weights:        weights,
			weightsAliases: weightsAliases,
			labels:         labels,
			total:          ".total",
			buckets:        buckets,
		},
		start: startVal,
		width: width,
	}
}

func (h *FixedHistogram) Add(v int64) {
	var n int64
	if v > h.start {
		n = v - h.start
		if n%h.width == 0 {
			n /= h.width
		} else {
			n = n/h.width + 1
		}
		if n >= int64(len(h.buckets)) {
			n = int64(len(h.buckets) - 1)
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *FixedHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}

// A VHistogram is implementation of Histogram with varibale-size buckets.
type VHistogram struct {
	HistogramStorage
}

func NewVHistogram(weights []int64, labels []string) *VHistogram {
	w := make([]int64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	// last := w[len(w)-2] + 1
	lbls := make([]string, len(w))

	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(last, 10)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			if i >= len(labels) || labels[i] == "" {
				lbls[i] = ".inf"
			} else {
				lbls[i] = labels[i]
			}
			weightsAliases[i] = "inf"
			w[i] = math.MaxInt64
		} else {
			weightsAliases[i] = strconv.FormatInt(w[i], 10)
			if i >= len(labels) || labels[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = labels[i]
			}
		}
	}

	return &VHistogram{
		HistogramStorage: HistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *VHistogram) Interface() HistogramInterface {
	return h
}

func (h *VHistogram) Snapshot() Histogram {
	return &HistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VHistogram) Add(v int64) {
	n := searchInt64Ge(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *VHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *VHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *VHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}
