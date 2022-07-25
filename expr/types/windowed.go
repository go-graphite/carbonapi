package types

import (
	"math"
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
}

func (w *Windowed) Reset() {
	w.length = 0
	w.head = 0
	w.sum = 0
	w.sumsq = 0
	w.nans = 0
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

// Mean returns mean value of data
func (w *Windowed) Mean() float64 { return w.sum / float64(w.Len()) }

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
