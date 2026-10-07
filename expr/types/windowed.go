package types

import (
	"math"
	"slices"
)

// Based on github.com/dgryski/go-onlinestats
// Copied here because we don't need the rest of the package, and we only need
// a small part of this type which we need to modify anyway.

// Note that this uses a slightly unstable but faster implementation of
// standard deviation.  This is also required to be compatible with graphite.

// Windowed is a struct to compute simple windowed stats
type Windowed struct {
	Data   []float64
	head   int
	length int
	sum    float64
	sumsq  float64
	nans   int
	sorted []float64
}

func (w *Windowed) Reset() {
	w.length = 0
	w.head = 0
	w.sum = 0
	w.sumsq = 0
	w.nans = 0
	w.sorted = nil
	for i := range w.Data {
		w.Data[i] = 0
	}
}

// Push pushes data
func (w *Windowed) Push(n float64) {
	if len(w.Data) == 0 {
		return
	}

	old := w.Data[w.head]

	w.length++

	w.Data[w.head] = n
	w.head++
	if w.head >= len(w.Data) {
		w.head = 0
	}

	if !math.IsNaN(old) {
		w.sum -= old
		w.sumsq -= (old * old)
	} else {
		w.nans--
	}

	if !math.IsNaN(n) {
		w.sum += n
		w.sumsq += (n * n)
	} else {
		w.nans++
	}

	// sorted is built by the first Median() call and kept up to date from then on
	if w.sorted != nil {
		w.replaceSorted(old, n)
	}
}

func (w *Windowed) replaceSorted(old, n float64) {
	switch {
	case math.IsNaN(old) && math.IsNaN(n):
	case math.IsNaN(old):
		i, _ := slices.BinarySearch(w.sorted, n)
		w.sorted = slices.Insert(w.sorted, i, n)
	case math.IsNaN(n):
		i, _ := slices.BinarySearch(w.sorted, old)
		w.sorted = slices.Delete(w.sorted, i, i+1)
	default:
		i, _ := slices.BinarySearch(w.sorted, old)
		j, _ := slices.BinarySearch(w.sorted, n)
		if j > i {
			j--
			copy(w.sorted[i:j], w.sorted[i+1:j+1])
		} else {
			copy(w.sorted[j+1:i+1], w.sorted[j:i])
		}
		w.sorted[j] = n
	}
}

// Len returns current len of data
func (w *Windowed) Len() int {
	if w.length < len(w.Data) {
		return w.length - w.nans
	}

	return len(w.Data) - w.nans
}

// Stdev computes standard deviation of data
func (w *Windowed) Stdev() float64 {
	l := w.Len()

	if l == 0 {
		return 0
	}

	n := float64(l)
	return math.Sqrt(n*w.sumsq-(w.sum*w.sum)) / n
}

// SumSQ returns sum of squares
func (w *Windowed) SumSQ() float64 {
	return w.sumsq
}

// Sum returns sum of data
func (w *Windowed) Sum() float64 {
	return w.sum
}

func (w *Windowed) Multiply() float64 {
	var rv = 1.0
	for _, f := range w.Data {
		if !math.IsNaN(rv) {
			rv *= f
		}
	}
	return rv
}

// Mean returns mean value of data
func (w *Windowed) Mean() float64 { return w.sum / float64(w.Len()) }

// MeanZero returns mean value of data, with NaN values replaced with 0
func (w *Windowed) MeanZero() float64 { return w.sum / float64(len(w.Data)) }

func (w *Windowed) Median() float64 {
	if w.sorted == nil {
		w.sorted = make([]float64, 0, len(w.Data))
		for _, v := range w.Data {
			if !math.IsNaN(v) {
				w.sorted = append(w.sorted, v)
			}
		}
		slices.Sort(w.sorted)
	}

	n := len(w.sorted)
	switch {
	case n == 0:
		return math.NaN()
	case n%2 == 1:
		return w.sorted[n/2]
	default:
		return (w.sorted[n/2-1] + w.sorted[n/2]) / 2
	}
}

// Max returns max(values)
func (w *Windowed) Max() float64 {
	rv := math.NaN()
	for _, f := range w.Data {
		if math.IsNaN(rv) || f > rv {
			rv = f
		}
	}
	return rv
}

// Min returns min(values)
func (w *Windowed) Min() float64 {
	rv := math.NaN()
	for _, f := range w.Data {
		if math.IsNaN(rv) || f < rv {
			rv = f
		}
	}
	return rv
}

// Count returns number of non-NaN points
func (w *Windowed) Count() float64 {
	return float64(w.Len())
}

// Diff subtracts series 2 through n from series 1
func (w *Windowed) Diff() float64 {
	rv := w.Data[w.head]
	for i, f := range w.Data {
		if !math.IsNaN(f) && i != w.head {
			rv -= f
		}
	}
	return rv
}

func (w *Windowed) Range() float64 {
	vMax := math.Inf(-1)
	vMin := math.Inf(1)
	for _, f := range w.Data {
		if f > vMax {
			vMax = f
		}
		if f < vMin {
			vMin = f
		}
	}
	return vMax - vMin
}

// Last returns the last data point
func (w *Windowed) Last() float64 {
	if w.head == 0 {
		return w.Data[len(w.Data)-1]
	}

	return w.Data[w.head-1]
}

// IsNonNull checks if the window's data contains only NaN values
// This is to prevent returning -Inf when the window's data contains only NaN values
func (w *Windowed) IsNonNull() bool {
	if len(w.Data) == w.nans {
		return false
	}
	return true
}
