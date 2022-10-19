package metrics

import (
	"math"
	"sort"
	"strconv"
)

// GetOrRegisterHistogram returns an existing Histogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedSumHistogram(startVal, endVal, width)
	}).(Histogram)
}

// GetOrRegisterSumHistogramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedSumHistogram(startVal, endVal, width)
	}).(Histogram)
}

// NewRegisteredFixedSumHistogram constructs and registers a new FixedSumHistogram (prometheus-like histogram).
func NewRegisteredFixedSumHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedSumHistogramT constructs and registers a new FixedSumHistogram (prometheus-like histogram).
func NewRegisteredFixedSumHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVSumHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVSumHistogram(weights, names)
	}).(Histogram)
}

func GetOrRegisterVSumHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVSumHistogram(weights, names)
	}).(Histogram)
}

// NewRegisteredVSumHistogram constructs and registers a new VSumHistogram (prometheus-like histogram).
func NewRegisteredVSumHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVSumHistogramT constructs and registers a new VSumHistogram (prometheus-like histogram).
func NewRegisteredVSumHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type SumHistogramSnapshot struct {
	weights        []int64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *SumHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *SumHistogramSnapshot) Labels() []string {
	return h.names
}

func (SumHistogramSnapshot) SetLabels([]string) Histogram {
	panic("SetLabels called on a HistogramSnapshot")
}

func (SumHistogramSnapshot) AddLabelPrefix(string) Histogram {
	panic("AddLabelPrefix called on a SumHistogramSnapshot")
}
func (SumHistogramSnapshot) SetNameTotal(string) Histogram {
	panic("SetNameTotal called on a SumHistogramSnapshot")
}

func (h *SumHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *SumHistogramSnapshot) Weights() []int64 {
	return h.weights
}

func (h *SumHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *SumHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *SumHistogramSnapshot) Add(v int64) {
	panic("Add called on a SumHistogramSnapshot")
}

func (h *SumHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a HistogramSnapshot")
}

func (h *SumHistogramSnapshot) Snapshot() Histogram {
	return h
}

func (SumHistogramSnapshot) IsSummed() bool { return true }

// A FixedSumHistogram is implementation of prometheus-like Histogram with fixed-size buckets.
type FixedSumHistogram struct {
	HistogramStorage
	start int64
	width int64
}

func NewFixedSumHistogram(startVal, endVal, width int64) *FixedSumHistogram {
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

	return &FixedSumHistogram{
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

func (h *FixedSumHistogram) Add(v int64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *FixedSumHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedSumHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedSumHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}

func (h *FixedSumHistogram) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *FixedSumHistogram) IsSummed() bool { return true }

// A VSumHistogram is implementation of prometheus-like Histogram with varibale-size buckets.
type VSumHistogram struct {
	HistogramStorage
}

func NewVSumHistogram(weights []int64, names []string) *VSumHistogram {
	w := make([]int64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	// last := w[len(w)-2] + 1
	lbls := make([]string, len(w))

	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(last, 10)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			if i >= len(names) || names[i] == "" {
				lbls[i] = ".inf"
			} else {
				lbls[i] = names[i]
			}
			weightsAliases[i] = "inf"
			w[i] = math.MaxInt64
		} else {
			weightsAliases[i] = strconv.FormatInt(w[i], 10)
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = names[i]
			}
		}
	}

	return &VSumHistogram{
		HistogramStorage: HistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VSumHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VSumHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *VSumHistogram) Interface() HistogramInterface {
	return h
}

func (h *VSumHistogram) Snapshot() Histogram {
	return &SumHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VSumHistogram) Add(v int64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *VSumHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *VSumHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *VSumHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}

func (h *VSumHistogram) IsSummed() bool { return true }
