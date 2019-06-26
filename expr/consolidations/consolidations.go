package consolidations

import (
	"math"
	"strconv"
	"strings"

	"github.com/wangjohn/quickselect"
	"gonum.org/v1/gonum/mat"
)

// ConsolidationToFunc contains a map of graphite-compatible consolidation functions definitions to actual functions that can do aggregation
// TODO(civil): take into account xFilesFactor
var ConsolidationToFunc = map[string]func([]float64) float64{
	"average":  AggMean,
	"avg_zero": AggMeanZero,
	"avg":      AggMean,
	"count":    summarizeToAggregate("count"),
	"diff":     summarizeToAggregate("diff"),
	"max":      AggMax,
	"maximum":  AggMax,
	"median":   summarizeToAggregate("median"),
	"min":      AggMin,
	"minimum":  AggMin,
	"multiply": summarizeToAggregate("multiply"),
	"range":    summarizeToAggregate("range"),
	"sum":      AggSum,
	"stddev":   summarizeToAggregate("stddev"),
	"first":    AggFirst,
	"last":     AggLast,
}

var AvailableSummarizers = []string{"sum", "total", "avg", "average", "avg_zero", "max", "min", "last", "range", "median", "multiply", "diff", "count", "stddev"}

// AvgValue returns average of list of values
func AvgValue(f64s []float64) float64 {
	var t float64
	var elts int
	for _, v := range f64s {
		if math.IsNaN(v) {
			continue
		}
		elts++
		t += v
	}
	return t / float64(elts)
}

// VarianceValue gets variances of list of values
func VarianceValue(f64s []float64) float64 {
	var squareSum float64
	var elts int

	mean := AvgValue(f64s)
	if math.IsNaN(mean) {
		return mean
	}

	for _, v := range f64s {
		if math.IsNaN(v) {
			continue
		}
		elts++
		squareSum += (mean - v) * (mean - v)
	}
	return squareSum / float64(elts)
}

// Percentile returns percent-th percentile. Can interpolate if needed
func Percentile(data []float64, percent float64, interpolate bool) float64 {
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

func summarizeToAggregate(f string) func([]float64) float64 {
	return func(v []float64) float64 {
		return SummarizeValues(f, v)
	}
}

// SummarizeValues summarizes values
func SummarizeValues(f string, values []float64) float64 {
	rv := 0.0

	if len(values) == 0 {
		return math.NaN()
	}

	switch f {
	case "sum", "total":
		for _, av := range values {
			rv += av
		}

	case "avg", "average", "avg_zero":
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
		rv = values[len(values)-1]
	case "range":
		vMax := math.Inf(-1)
		vMin := math.Inf(1)
		for _, av := range values {
			if av > vMax {
				vMax = av
			}
			if av < vMin {
				vMin = av
			}
		}
		rv = vMax - vMin
	case "median":
		rv = Percentile(values, 50, true)
	case "multiply":
		rv = values[0]
		for _, av := range values[1:] {
			rv *= av
		}
	case "diff":
		rv = values[0]
		for _, av := range values[1:] {
			rv -= av
		}
	case "count":
		rv = float64(len(values))
	case "stddev":
		rv = math.Sqrt(VarianceValue(values))
	default:
		f = strings.Split(f, "p")[1]
		percent, err := strconv.ParseFloat(f, 64)
		if err == nil {
			rv = Percentile(values, percent, true)
		}
	}

	return rv
}

var consolidateFuncs []string

// AvailableConsolidationFuncs lists all available consolidation functions
func AvailableConsolidationFuncs() []string {
	if len(consolidateFuncs) == 0 {
		for name := range ConsolidationToFunc {
			consolidateFuncs = append(consolidateFuncs, name)
		}
	}
	return consolidateFuncs
}

// AggMean computes mean (sum(v)/len(v), excluding NaN points) of values
func AggMean(v []float64) float64 {
	var sum float64
	var n int
	for _, vv := range v {
		if !math.IsNaN(vv) {
			sum += vv
			n++
		}
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// AggMeanZero computes mean (sum(v)/len(v), replacing NaN points with 0
func AggMeanZero(v []float64) float64 {
	var sum float64
	var n int
	for _, vv := range v {
		if !math.IsNaN(vv) {
			sum += vv
		}
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// AggMax computes max of values
func AggMax(v []float64) float64 {
	var m = math.Inf(-1)
	var abs = true
	for _, vv := range v {
		if !math.IsNaN(vv) {
			abs = false
			if m < vv {
				m = vv
			}
		}
	}
	if abs {
		return math.NaN()
	}
	return m
}

// AggMin computes min of values
func AggMin(v []float64) float64 {
	var m = math.Inf(1)
	var abs = true
	for _, vv := range v {
		if !math.IsNaN(vv) {
			abs = false
			if m > vv {
				m = vv
			}
		}
	}
	if abs {
		return math.NaN()
	}
	return m
}

// AggSum computes sum of values
func AggSum(v []float64) float64 {
	var sum float64
	var abs = true
	for _, vv := range v {
		if !math.IsNaN(vv) {
			sum += vv
			abs = false
		}
	}
	if abs {
		return math.NaN()
	}
	return sum
}

// AggFirst returns first point
func AggFirst(v []float64) float64 {
	var m = math.Inf(-1)
	var abs = true
	if len(v) > 0 {
		return v[0]
	}
	if abs {
		return math.NaN()
	}
	return m
}

// AggLast returns last point
func AggLast(v []float64) float64 {
	var m = math.Inf(-1)
	var abs = true
	if len(v) > 0 {
		return v[len(v)-1]
	}
	if abs {
		return math.NaN()
	}
	return m
}

// MaxValue returns maximum from the list
func MaxValue(f64s []float64) float64 {
	m := math.Inf(-1)
	for _, v := range f64s {
		if math.IsNaN(v) {
			continue
		}
		if v > m {
			m = v
		}
	}
	return m
}

// MinValue returns minimal from the list
func MinValue(f64s []float64) float64 {
	m := math.Inf(1)
	for _, v := range f64s {
		if math.IsNaN(v) {
			continue
		}
		if v < m {
			m = v
		}
	}
	return m
}

// CurrentValue returns last non-absent value (if any), otherwise returns NaN
func CurrentValue(f64s []float64) float64 {
	for i := len(f64s) - 1; i >= 0; i-- {
		if !math.IsNaN(f64s[i]) {
			return f64s[i]
		}
	}

	return math.NaN()
}

// Vandermonde creates a Vandermonde matrix
func Vandermonde(absent []float64, deg int) *mat.Dense {
	e := []float64{}
	for i := range absent {
		if math.IsNaN(absent[i]) {
			continue
		}
		v := 1
		for j := 0; j < deg+1; j++ {
			e = append(e, float64(v))
			v *= i
		}
	}
	return mat.NewDense(len(e)/(deg+1), deg+1, e)
}

// Poly computes polynom with specified coefficients
func Poly(x float64, coeffs ...float64) float64 {
	y := coeffs[0]
	v := 1.0
	for _, c := range coeffs[1:] {
		v *= x
		y += c * v
	}
	return y
}
