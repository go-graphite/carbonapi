package metrics

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// A UHistogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type UHistogram interface {
	HistogramInterface
	SetLabels([]string) UHistogram
	AddLabelPrefix(string) UHistogram
	SetNameTotal(string) UHistogram
	Snapshot() UHistogram
	Add(v uint64)
	Weights() []uint64
}

// GetOrRegisterHistogram returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterUFixedHistogram(name string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewUFixedHistogram(startVal, endVal, width)
	}).(UHistogram)
}

// GetOrRegisterHistogramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterUFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewUFixedHistogram(startVal, endVal, width)
	}).(UHistogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredUFixedHistogram(name string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUFixedHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredUFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUFixedHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterUVHistogram(name string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewUVHistogram(weights, names)
	}).(UHistogram)
}

func GetOrRegisterUVHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewUVHistogram(weights, names)
	}).(UHistogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredUVHistogram(name string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUVHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredUVHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUVHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type UHistogramSnapshot struct {
	weights        []uint64 // Sorted weights, by <=
	weightsAliases []string
	labels         []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *UHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *UHistogramSnapshot) Labels() []string {
	return h.labels
}

func (UHistogramSnapshot) SetLabels([]string) UHistogram {
	panic("SetLabels called on a UHistogramSnapshot")
}

func (UHistogramSnapshot) AddLabelPrefix(string) UHistogram {
	panic("AddLabelPrefix called on a UHistogramSnapshot")
}

func (UHistogramSnapshot) SetNameTotal(string) UHistogram {
	panic("SetNameTotal called on a UHistogramSnapshot")
}

func (h *UHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *UHistogramSnapshot) Weights() []uint64 {
	return h.weights
}

func (h *UHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *UHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *UHistogramSnapshot) Add(v uint64) {
	panic("Add called on a UHistogramSnapshot")
}

func (h *UHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a UHistogramSnapshot")
}

func (h *UHistogramSnapshot) Snapshot() UHistogram {
	return h
}

func (h *UHistogramSnapshot) IsSummed() bool { return false }

type UHistogramStorage struct {
	weights        []uint64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	labels         []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *UHistogramStorage) Labels() []string {
	return h.labels
}

func (h *UHistogramStorage) SetLabels(labels []string) {
	h.lock.Lock()
	for i := 0; i < Min(len(h.labels), len(labels)); i++ {
		h.labels[i] = labels[i]
	}
	h.lock.Unlock()
}

func (h *UHistogramStorage) AddLabelPrefix(labelPrefix string) {
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

func (h *UHistogramStorage) SetNameTotal(total string) {
	h.lock.Lock()
	h.total = total
	h.lock.Unlock()
}

func (h *UHistogramStorage) NameTotal() string {
	return h.total
}

func (h *UHistogramStorage) Weights() []uint64 {
	return h.weights
}

func (h *UHistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *UHistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *UHistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *UHistogramStorage) IsSummed() bool { return false }

func (h *UHistogramStorage) Snapshot() UHistogram {
	return &UHistogramSnapshot{
		labels:         h.labels,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *UHistogramStorage) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

// A UFixedHistogram is implementation of UHistogram with fixed-size buckets.
type UFixedHistogram struct {
	UHistogramStorage
	start uint64
	width uint64
}

func NewUFixedHistogram(startVal, endVal, width uint64) *UFixedHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	n := endVal - startVal
	count := n/width + 2
	if n%width != 0 {
		count++
	}
	weights := make([]uint64, count)
	weightsAliases := make([]string, count)
	labels := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal
	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(endVal+width, 10)))
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = math.MaxUint64
			weightsAliases[i] = "inf"
			labels[i] = ".inf"
		} else {
			weights[i] = ge
			weightsAliases[i] = strconv.FormatUint(ge, 10)
			labels[i] = "." + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, ge)
			ge += width
		}
	}

	return &UFixedHistogram{
		UHistogramStorage: UHistogramStorage{
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

func (h *UFixedHistogram) Add(v uint64) {
	var n uint64
	if v > h.start {
		n = v - h.start
		if n%h.width == 0 {
			n /= h.width
		} else {
			n = n/h.width + 1
		}
		if n >= uint64(len(h.buckets)) {
			n = uint64(len(h.buckets)) - 1
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *UFixedHistogram) SetLabels(labels []string) UHistogram {
	h.UHistogramStorage.SetLabels(labels)
	return h
}

func (h *UFixedHistogram) AddLabelPrefix(labelPrefix string) UHistogram {
	h.UHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *UFixedHistogram) SetNameTotal(total string) UHistogram {
	h.UHistogramStorage.SetNameTotal(total)
	return h
}

// A UVHistogram is implementation of UHistogram with varibale-size buckets.
type UVHistogram struct {
	UHistogramStorage
}

func NewUVHistogram(weights []uint64, labels []string) *UVHistogram {
	w := make([]uint64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	// last := w[len(w)-2] + 1
	lbls := make([]string, len(w))

	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(last, 10)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			weightsAliases[i] = "inf"
			if i >= len(labels) || labels[i] == "" {
				lbls[i] = ".inf"
			} else {
				lbls[i] = labels[i]
			}
			w[i] = math.MaxUint64
		} else {
			weightsAliases[i] = strconv.FormatUint(w[i], 10)
			if i >= len(labels) || labels[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = labels[i]
			}
		}
	}

	return &UVHistogram{
		UHistogramStorage: UHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *UVHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *UVHistogram) Snapshot() UHistogram {
	return &UHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		labels:         h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *UVHistogram) Add(v uint64) {
	n := searchUint64Ge(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *UVHistogram) SetLabels(labels []string) UHistogram {
	h.UHistogramStorage.SetLabels(labels)
	return h
}

func (h *UVHistogram) AddLabelPrefix(labelPrefix string) UHistogram {
	h.UHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *UVHistogram) SetNameTotal(total string) UHistogram {
	h.UHistogramStorage.SetNameTotal(total)
	return h
}
