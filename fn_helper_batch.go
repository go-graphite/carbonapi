package main

import (
	"math"

	"github.com/wangjohn/quickselect"
)

func percentile(data []float64, percent float64, interpolate bool) float64 {
	if len(data) == 0 || percent < 0 || percent > 100 {
		return math.NaN()
	}
	if len(data) == 1 {
		return data[0]
	}

	k := (float64(len(data)-1) * percent) / 100
	length := int(math.Ceil(k)) + 1
	quickselect.Float64QuickSelect(data, length)
	top, secondTop := math.Inf(-1), math.Inf(-1)
	for _, val := range data[0:length] {
		if val > top {
			secondTop = top
			top = val
		} else if val > secondTop {
			secondTop = val
		}
	}
	remainder := k - float64(int(k))
	if remainder == 0 || !interpolate {
		return top
	}
	return (top * remainder) + (secondTop * (1 - remainder))
}

type summarizeFunc func(f64s []float64, absent []bool) float64

func maxValue(f64s []float64, absent []bool) float64 {
	m := math.Inf(-1)
	for i, v := range f64s {
		if absent[i] {
			continue
		}
		if v > m {
			m = v
		}
	}
	return m
}

func minValue(f64s []float64, absent []bool) float64 {
	m := math.Inf(1)
	for i, v := range f64s {
		if absent[i] {
			continue
		}
		if v < m {
			m = v
		}
	}
	return m
}

func avgValue(f64s []float64, absent []bool) float64 {
	var t float64
	var elts int
	for i, v := range f64s {
		if absent[i] {
			continue
		}
		elts++
		t += v
	}
	return t / float64(elts)
}

func curValue(f64s []float64, absent []bool) float64 {

	for i := len(f64s) - 1; i >= 0; i-- {
		if !absent[i] {
			return f64s[i]
		}
	}

	return math.NaN()
}

func varianceValue(f64s []float64, absent []bool) float64 {
	var squareSum float64
	var elts int

	mean := avgValue(f64s, absent)
	if math.IsNaN(mean) {
		return mean
	}

	for i, v := range f64s {
		if absent[i] {
			continue
		}
		elts++
		squareSum += (mean - v) * (mean - v)
	}
	return squareSum / float64(elts)
}
