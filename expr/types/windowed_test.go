package types

import (
	"math"
	"math/rand"
	"slices"
	"strconv"
	"testing"

	"github.com/go-graphite/carbonapi/expr/consolidations"
)

func sameMedian(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

func TestWindowedMedian(t *testing.T) {
	nan := math.NaN()
	tests := []struct {
		window int
		values []float64
		want   []float64
	}{
		{window: 3, values: []float64{1, nan, 3}, want: []float64{2}},
		{window: 4, values: []float64{5, 1, nan, 9, 7}, want: []float64{5, 7}},
		{window: 3, values: []float64{1, 2, 3, nan, nan, nan, 10}, want: []float64{2, 2.5, 3, nan, 10}},
		{window: 2, values: []float64{4, 4, 4, 2}, want: []float64{4, 4, 3}},
		{window: 1, values: []float64{nan, 6}, want: []float64{nan, 6}},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.window), func(t *testing.T) {
			w := &Windowed{Data: make([]float64, tt.window)}
			var got []float64
			for i, v := range tt.values {
				w.Push(v)
				if i >= tt.window-1 {
					got = append(got, w.Median())
				}
			}
			if !slices.EqualFunc(got, tt.want, sameMedian) {
				t.Errorf("%v: got %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestWindowedMedianMatchesPercentile(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	special := []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, math.Copysign(0, -1)}

	for _, window := range []int{1, 2, 3, 10, 61} {
		w := &Windowed{Data: make([]float64, window)}
		for i := 0; i < 5000; i++ {
			if i == 2500 {
				w.Reset()
			}
			v := float64(r.Intn(20))
			if r.Intn(5) == 0 {
				v = special[r.Intn(len(special))]
			}
			w.Push(v)

			want := consolidations.Percentile(w.Data, 50, true)
			got := w.Median()
			if !sameMedian(got, want) {
				t.Fatalf("window %d, push %d: got %v, want %v, data %v", window, i, got, want, w.Data)
			}
		}
	}
}

func BenchmarkWindowedMedian(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	data := make([]float64, 8640)
	for i := range data {
		if i%17 == 0 {
			data[i] = math.NaN()
		} else {
			data[i] = float64(r.Intn(1000))
		}
	}

	for _, window := range []int{4, 60, 360, 600, 3600} {
		b.Run(strconv.Itoa(window), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				w := &Windowed{Data: make([]float64, window)}
				for j, v := range data {
					w.Push(v)
					if j >= window {
						_ = w.Median()
					}
				}
			}
		})
	}
}
