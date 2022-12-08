package metrics

import (
	"math"
	"strconv"
	"strings"
	"sync"
)

// A FHistogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type FHistogram interface {
	HistogramInterface
	SetLabels([]string) FHistogram
	AddLabelPrefix(string) FHistogram
	SetNameTotal(string) FHistogram
	Snapshot() FHistogram
	Add(v float64)
	Weights() []float64
}

// GetOrRegisterFHistogram returns an existing FHistogram or constructs and registers
// a new FFixedHistorgam.
func GetOrRegisterFixedFHistogram(name string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedFHistogram(startVal, endVal, width)
	}).(FHistogram)
}

// GetOrRegisterHistogramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFixedFHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedFHistogram(startVal, endVal, width)
	}).(FHistogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredFixedFHistogram(name string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedFHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredFixedFHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedFHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterFUHistogram(name string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFUHistogram(weights, names)
	}).(FHistogram)
}

func GetOrRegisterFUHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFUHistogram(weights, names)
	}).(FHistogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredFUHistogram(name string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFUHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredFUHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFUHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

func trimFloatZero(f string) string {
	d := strings.IndexByte(f, '.')
	if d == -1 {
		return f
	}
	for i := d + 1; i < len(f); i++ {
		if f[i] != '0' {
			return f
		}
	}
	return f[:d]
}

type NilFHistogram struct{}

func (NilFHistogram) Values() []uint64 {
	return nil
}

func (NilFHistogram) Labels() []string {
	return nil
}

func (NilFHistogram) SetLabels([]string) FHistogram { return NilFHistogram{} }

func (NilFHistogram) AddLabelPrefix(string) FHistogram { return NilFHistogram{} }

func (NilFHistogram) SetNameTotal(string) FHistogram { return NilFHistogram{} }

func (NilFHistogram) NameTotal() string { return "total" }

func (NilFHistogram) Weights() []float64 {
	return nil
}

func (NilFHistogram) WeightsAliases() []string {
	return nil
}

func (h NilFHistogram) Interface() HistogramInterface {
	return h
}

func (h NilFHistogram) Add(v float64) {}

func (h NilFHistogram) Clear() []uint64 {
	return nil
}

func (NilFHistogram) Snapshot() FHistogram { return NilFHistogram{} }

func (NilFHistogram) IsSummed() bool { return false }

type FHistogramSnapshot struct {
	weights        []float64 // Sorted weights, by <=
	weightsAliases []string
	labels         []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *FHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *FHistogramSnapshot) Labels() []string {
	return h.labels
}

func (FHistogramSnapshot) SetLabels([]string) FHistogram {
	panic("SetLabels called on a FHistogramSnapshot")
}

func (FHistogramSnapshot) AddLabelPrefix(string) FHistogram {
	panic("AddLabelPrefix called on a FHistogramSnapshot")
}
func (FHistogramSnapshot) SetNameTotal(string) FHistogram {
	panic("SetNameTotal called on a FHistogramSnapshot")
}

func (h *FHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *FHistogramSnapshot) Weights() []float64 {
	return h.weights
}

func (h *FHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *FHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *FHistogramSnapshot) Add(v float64) {
	panic("Add called on a FHistogramSnapshot")
}

func (h *FHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a FHistogramSnapshot")
}

func (h *FHistogramSnapshot) Snapshot() FHistogram {
	return h
}

func (h *FHistogramSnapshot) IsSummed() bool { return false }

type FHistogramStorage struct {
	weights        []float64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	labels         []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *FHistogramStorage) Labels() []string {
	return h.labels
}

func (h *FHistogramStorage) SetLabels(labels []string) {
	h.lock.Lock()
	for i := 0; i < Min(len(h.labels), len(labels)); i++ {
		h.labels[i] = labels[i]
	}
	h.lock.Unlock()
}

func (h *FHistogramStorage) AddLabelPrefix(labelPrefix string) {
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

func (h *FHistogramStorage) SetNameTotal(total string) {
	h.lock.Lock()
	h.total = total
	h.lock.Unlock()
}

func (h *FHistogramStorage) NameTotal() string {
	return h.total
}

func (h *FHistogramStorage) Weights() []float64 {
	return h.weights
}

func (h *FHistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *FHistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *FHistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *FHistogramStorage) Snapshot() FHistogram {
	return &FHistogramSnapshot{
		labels:         h.labels,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *FHistogramStorage) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *FHistogramStorage) IsSummed() bool { return false }

// A FixedFHistogram is implementation of FHistogram with fixed-size buckets.
type FixedFHistogram struct {
	FHistogramStorage
	start float64
	width float64
}

func NewFixedFHistogram(startVal, endVal, width float64) FHistogram {
	if UseNilMetrics {
		return NilFHistogram{}
	}
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	if width < 0 {
		width = -width
	}
	n := (endVal - startVal) / width
	if n > float64(int(n)) {
		n++
	}

	count := int(n) + 2
	weights := make([]float64, count)
	weightsAliases := make([]string, count)
	labels := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal

	// maxLength := len(strconv.FormatInt(int64(endVal+width)+1, 10))
	// fmtStr := fmt.Sprintf("%%s%%0%dd", maxLength)
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = math.MaxFloat64
			weightsAliases[i] = "inf"
			labels[i] = ".inf"
		} else {
			weights[i] = ge
			// n := int(ge)
			// d := ge - float64(n)
			weightsAliases[i] = strings.ReplaceAll(trimFloatZero(strconv.FormatFloat(ge, 'f', 2, 64)), ".", "_")
			labels[i] = "." + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, n)
			ge += width
		}
	}

	return &FixedFHistogram{
		FHistogramStorage: FHistogramStorage{
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

func (h *FixedFHistogram) Add(v float64) {
	var (
		n int
		f float64
	)
	if v > h.start {
		f = (v - h.start) / h.width
		if f > float64(int(f)) {
			n = int(f) + 1
		} else {
			n = int(f)
		}
		if n >= len(h.buckets) {
			n = len(h.buckets) - 1
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *FixedFHistogram) SetLabels(labels []string) FHistogram {
	h.FHistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedFHistogram) AddLabelPrefix(labelPrefix string) FHistogram {
	h.FHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedFHistogram) SetNameTotal(total string) FHistogram {
	h.FHistogramStorage.SetNameTotal(total)
	return h
}

// A FUHistogram is implementation of FHistogram with varibale-size buckets.
type FUHistogram struct {
	FHistogramStorage
}

func NewFUHistogram(weights []float64, names []string) FHistogram {
	if UseNilMetrics {
		return NilFHistogram{}
	}
	if !IsSortedSliceFloat64Le(weights) {
		panic(ErrUnsortedWeights)
	}
	w := make([]float64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	// last := w[len(w)-2] + 1
	lbls := make([]string, len(w))

	// fmtStr := fmt.Sprintf("%%s%%0%df", len(strconv.FormatFloat(last, 'f', -1, 64)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			weightsAliases[i] = "inf"
			if i >= len(names) || names[i] == "" {
				lbls[i] = ".inf"
			} else {
				lbls[i] = names[i]
			}
			w[i] = math.MaxFloat64
		} else {
			weightsAliases[i] = strings.ReplaceAll(trimFloatZero(strconv.FormatFloat(w[i], 'f', 2, 64)), ".", "_")
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = names[i]
			}
		}
	}

	return &FUHistogram{
		FHistogramStorage: FHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *FUHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *FUHistogram) Snapshot() FHistogram {
	return &FHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		labels:         h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *FUHistogram) Add(v float64) {
	n := SearchFloat64Le(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

func (h *FUHistogram) SetLabels(labels []string) FHistogram {
	h.FHistogramStorage.SetLabels(labels)
	return h
}

func (h *FUHistogram) AddLabelPrefix(labelPrefix string) FHistogram {
	h.FHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FUHistogram) SetNameTotal(total string) FHistogram {
	h.FHistogramStorage.SetNameTotal(total)
	return h
}
