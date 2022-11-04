package consolidations

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ansel1/merry"

	"github.com/wangjohn/quickselect"
	"gonum.org/v1/gonum/mat"
)

var ErrInvalidConsolidationFunc = merry.New("Invalid Consolidation Function")

// ConsolidationToFunc contains a map of graphite-compatible consolidation functions definitions to actual functions that can do aggregation
// TODO(civil): take into account xFilesFactor
var ConsolidationToFunc = map[string]func([]float64) float64{
	"average":  AggMean,
	"avg_zero": AggMeanZero,
	"avg":      AggMean,
	"count":    AggCount,
	"diff":     AggDiff,
	"max":      AggMax,
	"maximum":  AggMax,
	"median":   summarizeToAggregate("median"),
	"min":      AggMin,
	"minimum":  AggMin,
	"multiply": summarizeToAggregate("multiply"),
	"range":    summarizeToAggregate("range"),
	"rangeOf":  summarizeToAggregate("rangeOf"),
	"sum":      AggSum,
	"total":    AggSum,
	"stddev":   summarizeToAggregate("stddev"),
	"first":    AggFirst,
	"last":     AggLast,
	"current":  AggLast,
}

var AvailableSummarizers = []string{"sum", "total", "avg", "average", "avg_zero", "max", "min", "last", "current", "first", "range", "rangeOf", "median", "multiply", "diff", "count", "stddev"}

func CheckValidConsolidationFunc(functionName string) error {
	if _, ok := ConsolidationToFunc[functionName]; ok {
		return nil
	} else {
		// Check if this is a p50 - p99.9 consolidation
		if match, _ := regexp.MatchString("p([0-9]*[.])?[0-9]+", functionName); match {
			return nil
		}
	}
	return ErrInvalidConsolidationFunc.WithMessage("invalid consolidation " + functionName)
}

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
	dataFiltered := make([]float64, 0)
	for _, v := range data {
		if !math.IsNaN(v) {
			dataFiltered = append(dataFiltered, v)
		}
	}

	if len(dataFiltered) == 0 || percent < 0 || percent > 100 {
		return math.NaN()
	}
	if len(dataFiltered) == 1 {
		return dataFiltered[0]
	}

	k := (float64(len(dataFiltered)-1) * percent) / 100
	length := int(math.Ceil(k)) + 1

	_ = quickselect.Float64QuickSelect(dataFiltered, length)
	top, secondTop := math.Inf(-1), math.Inf(-1)
	for _, val := range dataFiltered[0:length] {
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
		return SummarizeValues(f, v, 0)
	}
}

// SummarizeValues summarizes values
func SummarizeValues(f string, values []float64, XFilesFactor float32) float64 {
	rv := 0.0
	total := 0

	notNans := func(values []float64) int {
		t := 0
		for _, av := range values {
			if !math.IsNaN(av) {
				t++
			}
		}
		return t
	}

	if len(values) == 0 {
		return math.NaN()
	}

	switch f {
	case "sum", "total":
		for _, av := range values {
			if !math.IsNaN(av) {
				rv += av
				total++
			}
		}

	case "avg", "average", "avg_zero":
		for _, av := range values {
			if !math.IsNaN(av) {
				rv += av
				total++
			}
		}
		if total == 0 {
			return math.NaN()
		}
		rv /= float64(total)
	case "max":
		rv = math.Inf(-1)
		for _, av := range values {
			if !math.IsNaN(av) {
				total++
				if av > rv {
					rv = av
				}
			}
		}
	case "min":
		rv = math.Inf(1)
		for _, av := range values {
			if !math.IsNaN(av) {
				total++
				if av < rv {
					rv = av
				}
			}
		}
	case "last", "current":
		rv = values[len(values)-1]
		total = notNans(values)
	case "range", "rangeOf":
		vMax := math.Inf(-1)
		vMin := math.Inf(1)
		isNaN := true
		for _, av := range values {
			if !math.IsNaN(av) {
				total++
				isNaN = false
			}
			if av > vMax {
				vMax = av
			}
			if av < vMin {
				vMin = av
			}
		}
		if isNaN {
			rv = math.NaN()
		} else {
			rv = vMax - vMin
		}
	case "median":
		rv = Percentile(values, 50, true)
		total = notNans(values)
	case "multiply":
		rv = 1.0
		for _, v := range values {
			if math.IsNaN(v) {
				rv = math.NaN()
				break
			} else {
				total++
				rv *= v
			}
		}
	case "diff":
		rv = values[0]
		for _, av := range values[1:] {
			if !math.IsNaN(av) {
				total++
				rv -= av
			}
		}
	case "count":
		rv = float64(len(values))
		total = notNans(values)
	case "stddev":
		rv = math.Sqrt(VarianceValue(values))
		total = notNans(values)
	case "first":
		if len(values) > 0 {
			rv = values[0]
		} else {
			rv = math.NaN()
		}
		total = notNans(values)
	default:
		// This processes function percentile functions such as p50 or p99.9.
		// If a function name is passed in that does not match that format,
		// it should be ignored
		fn := strings.Split(f, "p")
		if len(fn) > 1 {
			f = fn[1]
			percent, err := strconv.ParseFloat(f, 64)
			if err == nil {
				total = notNans(values)
				rv = Percentile(values, percent, true)
			}
		}
	}

	if float32(total)/float32(len(values)) < XFilesFactor {
		return math.NaN()
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
	n := len(v)
	n2 := 0
	for _, vv := range v {
		if !math.IsNaN(vv) {
			sum += vv
			n2++
		}
	}

	if n2 == 0 {
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
	if len(v) > 0 {
		i := len(v)
		for i != 0 {
			i--
			if !math.IsNaN(v[i]) {
				return v[i]
			}
		}
	}
	return math.NaN()
}

// AggCount counts non-NaN points
func AggCount(v []float64) float64 {
	n := 0

	for _, vv := range v {
		if !math.IsNaN(vv) {
			n++
		}
	}

	if n == 0 {
		return math.NaN()
	}

	return float64(n)
}

func AggDiff(v []float64) float64 {
	safeValues := make([]float64, 0, len(v))
	for _, vv := range v {
		if !math.IsNaN(vv) {
			safeValues = append(safeValues, vv)
		}
	}

	if len(safeValues) > 0 {
		res := safeValues[0]
		if len(safeValues) == 1 {
			return res
		}

		for _, vv := range safeValues[1:] {
			res -= vv
		}

		return res
	} else {
		return math.NaN()
	}
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
