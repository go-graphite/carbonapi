package main

import (
	"math"
	"strconv"
	"strings"
)

func summarizeValues(f string, values []float64) float64 {
	rv := 0.0

	if len(values) == 0 {
		return math.NaN()
	}

	switch f {
	case "sum":
		for _, av := range values {
			rv += av
		}

	case "avg":
		for _, av := range values {
			rv += av
		}
		rv /= float64(len(values))
	case "max":
		rv = math.Inf(-1)
		for _, av := range values {
			if av > rv {
				rv = av
			}
		}
	case "min":
		rv = math.Inf(1)
		for _, av := range values {
			if av < rv {
				rv = av
			}
		}
	case "last":
		if len(values) > 0 {
			rv = values[len(values)-1]
		}

	default:
		f = strings.Split(f, "p")[1]
		percent, err := strconv.ParseFloat(f, 64)
		if err == nil {
			rv = percentile(values, percent, true)
		}
	}

	return rv
}
