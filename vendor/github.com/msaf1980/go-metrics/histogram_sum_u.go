package metrics

import (
	"math"
	"sort"
	"strconv"
)

// GetOrRegisterHistogram returns an existing UHistogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumUHistogram(name string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedSumUHistogram(startVal, endVal, width)
	}).(UHistogram)
}

// GetOrRegisterSumUHistogramT returns an existing UHistogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumUHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedSumUHistogram(startVal, endVal, width)
	}).(UHistogram)
}

// NewRegisteredFixedSumUHistogram constructs and registers a new FixedSumUHistogram (prometheus-like histogram).
func NewRegisteredFixedSumUHistogram(name string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumUHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedSumUHistogramT constructs and registers a new FixedSumUHistogram (prometheus-like histogram).
func NewRegisteredFixedSumUHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumUHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVSumUHistogram(name string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVSumUHistogram(weights, names)
	}).(UHistogram)
}

func GetOrRegisterVSumUHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVSumUHistogram(weights, names)
	}).(UHistogram)
}

// NewRegisteredVSumUHistogram constructs and registers a new VSumUHistogram (prometheus-like histogram).
func NewRegisteredVSumUHistogram(name string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumUHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVSumUHistogramT constructs and registers a new VSumUHistogram (prometheus-like histogram).
func NewRegisteredVSumUHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumUHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type SumUHistogramSnapshot struct {
	weights        []uint64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *SumUHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *SumUHistogramSnapshot) Labels() []string {
	return h.names
}

func (SumUHistogramSnapshot) SetLabels([]string) UHistogram {
	panic("SetLabels called on a UHistogramSnapshot")
}

func (SumUHistogramSnapshot) AddLabelPrefix(string) UHistogram {
	panic("AddLabelPrefix called on a SumUHistogramSnapshot")
}
func (SumUHistogramSnapshot) SetNameTotal(string) UHistogram {
	panic("SetNameTotal called on a SumUHistogramSnapshot")
}

func (h *SumUHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *SumUHistogramSnapshot) Weights() []uint64 {
	return h.weights
}

func (h *SumUHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with UHistogramInterface
func (h *SumUHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *SumUHistogramSnapshot) Add(v uint64) {
	panic("Add called on a SumUHistogramSnapshot")
}

func (h *SumUHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a UHistogramSnapshot")
}

func (h *SumUHistogramSnapshot) Snapshot() UHistogram {
	return h
}

func (SumUHistogramSnapshot) IsSummed() bool { return true }

// A FixedSumUHistogram is implementation of prometheus-like UHistogram with fixed-size buckets.
type FixedSumUHistogram struct {
	UHistogramStorage
	start uint64
	width uint64
}

func NewFixedSumUHistogram(startVal, endVal, width uint64) *FixedSumUHistogram {
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
			weights[i] = ge
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

	return &FixedSumUHistogram{
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

func (h *FixedSumUHistogram) Add(v uint64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *FixedSumUHistogram) SetLabels(labels []string) UHistogram {
	h.UHistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedSumUHistogram) AddLabelPrefix(labelPrefix string) UHistogram {
	h.UHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedSumUHistogram) SetNameTotal(total string) UHistogram {
	h.UHistogramStorage.SetNameTotal(total)
	return h
}

func (h *FixedSumUHistogram) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *FixedSumUHistogram) IsSummed() bool { return true }

// A VSumUHistogram is implementation of prometheus-like UHistogram with varibale-size buckets.
type VSumUHistogram struct {
	UHistogramStorage
}

func NewVSumUHistogram(weights []uint64, names []string) *VSumUHistogram {
	w := make([]uint64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
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
			w[i] = math.MaxUint64
		} else {
			weightsAliases[i] = strconv.FormatUint(w[i], 10)
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = names[i]
			}
		}
	}

	return &VSumUHistogram{
		UHistogramStorage: UHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VSumUHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VSumUHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with UHistogramInterface
func (h *VSumUHistogram) Interface() HistogramInterface {
	return h
}

func (h *VSumUHistogram) Snapshot() UHistogram {
	return &SumUHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VSumUHistogram) Add(v uint64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *VSumUHistogram) SetLabels(labels []string) UHistogram {
	h.UHistogramStorage.SetLabels(labels)
	return h
}

func (h *VSumUHistogram) AddLabelPrefix(labelPrefix string) UHistogram {
	h.UHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *VSumUHistogram) SetNameTotal(total string) UHistogram {
	h.UHistogramStorage.SetNameTotal(total)
	return h
}

func (h *VSumUHistogram) IsSummed() bool { return true }
