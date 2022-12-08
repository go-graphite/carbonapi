package metrics

import (
	"math"
	"strconv"
	"strings"
)

// GetOrRegisterFHistogram returns an existing FHistogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumFHistogram(name string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedSumFHistogram(startVal, endVal, width)
	}).(FHistogram)
}

// GetOrRegisterSumFHistogramT returns an existing FHistogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumFHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedSumFHistogram(startVal, endVal, width)
	}).(FHistogram)
}

// NewRegisteredFixedSumFHistogram constructs and registers a new FixedSumFHistogram (prometheus-like histogram).
func NewRegisteredFixedSumFHistogram(name string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumFHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedSumFHistogramT constructs and registers a new FixedSumFHistogram (prometheus-like histogram).
func NewRegisteredFixedSumFHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumFHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVSumFHistogram(name string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVSumFHistogram(weights, names)
	}).(FHistogram)
}

func GetOrRegisterVSumFHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVSumFHistogram(weights, names)
	}).(FHistogram)
}

// NewRegisteredVSumFHistogram constructs and registers a new VSumFHistogram (prometheus-like histogram).
func NewRegisteredVSumFHistogram(name string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumFHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVSumFHistogramT constructs and registers a new VSumFHistogram (prometheus-like histogram).
func NewRegisteredVSumFHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumFHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type SumFHistogramSnapshot struct {
	weights        []float64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *SumFHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *SumFHistogramSnapshot) Labels() []string {
	return h.names
}

func (SumFHistogramSnapshot) SetLabels([]string) FHistogram {
	panic("SetLabels called on a FHistogramSnapshot")
}

func (SumFHistogramSnapshot) AddLabelPrefix(string) FHistogram {
	panic("AddLabelPrefix called on a SumFHistogramSnapshot")
}
func (SumFHistogramSnapshot) SetNameTotal(string) FHistogram {
	panic("SetNameTotal called on a SumFHistogramSnapshot")
}

func (h *SumFHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *SumFHistogramSnapshot) Weights() []float64 {
	return h.weights
}

func (h *SumFHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with FHistogramInterface
func (h *SumFHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *SumFHistogramSnapshot) Add(v float64) {
	panic("Add called on a SumFHistogramSnapshot")
}

func (h *SumFHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a FHistogramSnapshot")
}

func (h *SumFHistogramSnapshot) Snapshot() FHistogram {
	return h
}

func (SumFHistogramSnapshot) IsSummed() bool { return true }

// A FixedSumFHistogram is implementation of prometheus-like FHistogram with fixed-size buckets.
type FixedSumFHistogram struct {
	FHistogramStorage
	start float64
	width float64
}

func NewFixedSumFHistogram(startVal, endVal, width float64) *FixedSumFHistogram {
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

	return &FixedSumFHistogram{
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

func (h *FixedSumFHistogram) Add(v float64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *FixedSumFHistogram) SetLabels(labels []string) FHistogram {
	h.FHistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedSumFHistogram) AddLabelPrefix(labelPrefix string) FHistogram {
	h.FHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedSumFHistogram) SetNameTotal(total string) FHistogram {
	h.FHistogramStorage.SetNameTotal(total)
	return h
}

func (h *FixedSumFHistogram) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *FixedSumFHistogram) IsSummed() bool { return true }

// A VSumFHistogram is implementation of prometheus-like FHistogram with varibale-size buckets.
type VSumFHistogram struct {
	FHistogramStorage
}

func NewVSumFHistogram(weights []float64, names []string) *VSumFHistogram {
	if !IsSortedSliceFloat64Le(weights) {
		panic(ErrUnsortedWeights)
	}
	w := make([]float64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
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

	return &VSumFHistogram{
		FHistogramStorage: FHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VSumFHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VSumFHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with FHistogramInterface
func (h *VSumFHistogram) Interface() HistogramInterface {
	return h
}

func (h *VSumFHistogram) Snapshot() FHistogram {
	return &SumFHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VSumFHistogram) Add(v float64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *VSumFHistogram) SetLabels(labels []string) FHistogram {
	h.FHistogramStorage.SetLabels(labels)
	return h
}

func (h *VSumFHistogram) AddLabelPrefix(labelPrefix string) FHistogram {
	h.FHistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *VSumFHistogram) SetNameTotal(total string) FHistogram {
	h.FHistogramStorage.SetNameTotal(total)
	return h
}

func (h *VSumFHistogram) IsSummed() bool { return true }
