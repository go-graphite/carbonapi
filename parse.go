package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	"github.com/gogo/protobuf/proto"
)

func parseExpr(e string) (*expr, string, error) {

	// skip whitespace
	for len(e) > 1 && e[0] == ' ' {
		e = e[1:]
	}

	if len(e) == 0 {
		return nil, "", ErrMissingExpr
	}

	if '0' <= e[0] && e[0] <= '9' || e[0] == '-' || e[0] == '+' {
		val, e, err := parseConst(e)
		return &expr{val: val, etype: etConst}, e, err
	}

	if e[0] == '\'' || e[0] == '"' {
		val, e, err := parseString(e)
		return &expr{valStr: val, etype: etString}, e, err
	}

	name, e := parseName(e)

	if name == "" {
		return nil, e, ErrMissingArgument
	}

	if e != "" && e[0] == '(' {
		exp := &expr{target: name, etype: etFunc}

		argString, args, e, err := parseArgList(e)
		exp.argString = argString
		exp.args = args

		return exp, e, err
	}

	return &expr{target: name}, e, nil
}

var (
	ErrMissingExpr         = errors.New("missing expression")
	ErrMissingComma        = errors.New("missing comma")
	ErrMissingQuote        = errors.New("missing quote")
	ErrUnexpectedCharacter = errors.New("unexpected character")
)

func parseArgList(e string) (string, []*expr, string, error) {

	var args []*expr

	if e[0] != '(' {
		panic("arg list should start with paren")
	}

	argString := e[1:]

	e = e[1:]

	for {
		var arg *expr
		var err error
		arg, e, err = parseExpr(e)
		if err != nil {
			return "", nil, e, err
		}
		args = append(args, arg)

		if e == "" {
			return "", nil, "", ErrMissingComma
		}

		if e[0] == ')' {
			return argString[:len(argString)-len(e)], args, e[1:], nil
		}

		if e[0] != ',' && e[0] != ' ' {
			return "", nil, "", ErrUnexpectedCharacter
		}

		e = e[1:]
	}
}

func isNameChar(r byte) bool {
	return false ||
		'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		'0' <= r && r <= '9' ||
		r == '.' || r == '_' || r == '-' || r == '*' || r == '?' || r == ':' ||
		r == '[' || r == ']'
}

func isDigit(r byte) bool {
	return '0' <= r && r <= '9'
}

func parseConst(s string) (float64, string, error) {

	var i int
	// All valid characters for a floating-point constant
	// Just slurp them all in and let ParseFloat sort 'em out
	for i < len(s) && (isDigit(s[i]) || s[i] == '.' || s[i] == '+' || s[i] == '-' || s[i] == 'e' || s[i] == 'E') {
		i++
	}

	v, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, "", err
	}

	return v, s[i:], err
}

func parseName(s string) (string, string) {

	var i int

FOR:
	for braces := 0; i < len(s); i++ {

		if isNameChar(s[i]) {
			continue
		}

		switch s[i] {
		case '{':
			braces++
		case '}':
			if braces == 0 {
				break FOR

			}
			braces--
		case ',':
			if braces == 0 {
				break FOR
			}
		default:
			break FOR
		}

	}

	if i == len(s) {
		return s, ""
	}

	return s[:i], s[i:]
}

func parseString(s string) (string, string, error) {

	if s[0] != '\'' && s[0] != '"' {
		panic("string should start with open quote")
	}

	match := s[0]

	s = s[1:]

	var i int
	for i < len(s) && s[i] != match {
		i++
	}

	if i == len(s) {
		return "", "", ErrMissingQuote

	}

	return s[:i], s[i+1:], nil
}

var (
	ErrBadType           = errors.New("bad type")
	ErrMissingArgument   = errors.New("missing argument")
	ErrMissingTimeseries = errors.New("missing time series")
)

func getStringArg(e *expr, n int) (string, error) {
	if len(e.args) <= n {
		return "", ErrMissingArgument
	}

	if e.args[n].etype != etString {
		return "", ErrBadType
	}

	return e.args[n].valStr, nil
}

func getStringArgDefault(e *expr, n int, s string) (string, error) {
	if len(e.args) <= n {
		return s, nil
	}

	if e.args[n].etype != etString {
		return "", ErrBadType
	}

	return e.args[n].valStr, nil
}

func getIntervalArg(e *expr, n int, defaultSign int) (int32, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	if e.args[n].etype != etString {
		return 0, ErrBadType
	}

	seconds, err := intervalString(e.args[n].valStr, defaultSign)
	if err != nil {
		return 0, ErrBadType
	}

	return seconds, nil
}

func getFloatArg(e *expr, n int) (float64, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	if e.args[n].etype != etConst {
		return 0, ErrBadType
	}

	return e.args[n].val, nil
}

func getFloatArgDefault(e *expr, n int, v float64) (float64, error) {
	if len(e.args) <= n {
		return v, nil
	}

	if e.args[n].etype != etConst {
		return 0, ErrBadType
	}

	return e.args[n].val, nil
}

func getIntArg(e *expr, n int) (int, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	if e.args[n].etype != etConst {
		return 0, ErrBadType
	}

	return int(e.args[n].val), nil
}

func getIntArgs(e *expr, n int) ([]int, error) {

	if len(e.args) <= n {
		return nil, ErrMissingArgument
	}

	var ints []int

	for i := n; i < len(e.args); i++ {
		a, err := getIntArg(e, i)
		if err != nil {
			return nil, err
		}
		ints = append(ints, a)
	}

	return ints, nil
}

func getIntArgDefault(e *expr, n int, d int) (int, error) {
	if len(e.args) <= n {
		return d, nil
	}

	if e.args[n].etype != etConst {
		return 0, ErrBadType
	}

	return int(e.args[n].val), nil
}

func getBoolArgDefault(e *expr, n int, b bool) (bool, error) {
	if len(e.args) <= n {
		return b, nil
	}

	if e.args[n].etype != etName {
		return false, ErrBadType
	}

	// names go into 'target'
	switch e.args[n].target {
	case "False", "false":
		return false, nil
	case "True", "true":
		return true, nil
	}

	return false, ErrBadType
}

func getSeriesArg(arg *expr, from, until int32, values map[metricRequest][]*metricData) ([]*metricData, error) {

	if arg.etype != etName && arg.etype != etFunc {
		return nil, ErrMissingTimeseries
	}
	a := evalExpr(arg, from, until, values)

	if len(a) == 0 {
		return nil, ErrMissingTimeseries
	}

	return a, nil
}

func getSeriesArgs(e []*expr, from, until int32, values map[metricRequest][]*metricData) ([]*metricData, error) {

	var args []*metricData

	for _, arg := range e {
		a, err := getSeriesArg(arg, from, until, values)
		if err != nil {
			return nil, err
		}
		args = append(args, a...)
	}

	if len(args) == 0 {
		return nil, ErrMissingTimeseries
	}

	return args, nil
}

func evalExpr(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	switch e.etype {
	case etName:
		return values[metricRequest{metric: e.target, from: from, until: until}]
	case etConst:
		p := metricData{FetchResponse: pb.FetchResponse{Name: proto.String(e.target), Values: []float64{e.val}}}
		return []*metricData{&p}
	}

	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.args) == 0 {
		return nil
	}

	execFn := getFn(e.target)
	if execFn == nil {
		logger.Logf("unknown function in evalExpr: %q\n", e.target)
		return nil
	}

	return execFn(e, from, until, values)
}

// Total (sortByTotal), max (sortByMaxima), min (sortByMinima) sorting
// For 'min', we actually store 1/v so the sorting logic is the same
type byVals struct {
	vals   []float64
	series []*metricData
}

func (s byVals) Len() int { return len(s.series) }
func (s byVals) Swap(i, j int) {
	s.series[i], s.series[j] = s.series[j], s.series[i]
	s.vals[i], s.vals[j] = s.vals[j], s.vals[i]
}
func (s byVals) Less(i, j int) bool {
	// actually "greater than"
	return s.vals[i] > s.vals[j]
}

// ByName sorts metrics by name
type ByName []*metricData

func (s ByName) Len() int           { return len(s) }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByName) Less(i, j int) bool { return s[i].GetName() < s[j].GetName() }

type seriesFunc func(*metricData, *metricData) *metricData

type aggregateFunc func([]float64) float64

func aggregateSeries(e *expr, args []*metricData, function aggregateFunc) []*metricData {
	length := len(args[0].Values)
	r := *args[0]
	r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, e.argString))
	r.Values = make([]float64, length)
	r.IsAbsent = make([]bool, length)

	for i := range args[0].Values {
		var values []float64
		for _, arg := range args {
			if !arg.IsAbsent[i] {
				values = append(values, arg.Values[i])
			}
		}

		r.Values[i] = math.NaN()
		if len(values) > 0 {
			r.Values[i] = function(values)
		}

		r.IsAbsent[i] = math.IsNaN(r.Values[i])
	}

	return []*metricData{&r}
}

func getBuckets(start, stop, bucketSize int32) int32 {
	return int32(math.Ceil(float64(stop-start) / float64(bucketSize)))
}

func alignStartToInterval(start, stop, bucketSize int32) int32 {
	for _, v := range []int32{86400, 3600, 60} {
		if bucketSize >= v {
			start -= start % v
			break
		}
	}

	return start
}

func alignToBucketSize(start, stop, bucketSize int32) (int32, int32) {
	start = int32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
	newStop := int32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())

	// check if a partial bucket is needed
	if stop != newStop {
		newStop += bucketSize
	}

	return start, newStop
}

func extractMetric(m string) string {

	// search for a metric name in `m'
	// metric name is defined to be a series of name characters terminated by a comma

	start := 0
	end := 0
	curlyBraces := 0
	for end < len(m) {
		if m[end] == '{' {
			curlyBraces++
		} else if m[end] == '}' {
			curlyBraces--
		} else if m[end] == ')' || (m[end] == ',' && curlyBraces == 0) {
			return m[start:end]
		} else if !(isNameChar(m[end]) || m[end] == ',') {
			start = end + 1
		}

		end++
	}

	return m[start:end]
}

func contains(a []int, i int) bool {
	for _, aa := range a {
		if aa == i {
			return true
		}
	}
	return false
}

// Based on github.com/dgryski/go-onlinestats
// Copied here because we don't need the rest of the package, and we only need
// a small part of this type which we need to modify anyway.

// Note that this uses a slightly unstable but faster implementation of
// standard deviation.  This is also required to be compatible with graphite.

type Windowed struct {
	data   []float64
	head   int
	length int
	sum    float64
	sumsq  float64
	nans   int
}

func (w *Windowed) Push(n float64) {
	old := w.data[w.head]

	w.length++

	w.data[w.head] = n
	w.head++
	if w.head >= len(w.data) {
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

func (w *Windowed) Len() int {
	if w.length < len(w.data) {
		return w.length - w.nans
	}

	return len(w.data) - w.nans
}

func (w *Windowed) Stdev() float64 {
	l := w.Len()

	if l == 0 {
		return 0
	}

	n := float64(l)
	return math.Sqrt(n*w.sumsq-(w.sum*w.sum)) / n
}

func (w *Windowed) Mean() float64 { return w.sum / float64(w.Len()) }

type metricHeapElement struct {
	idx int
	val float64
}

type metricHeap []metricHeapElement

func (m metricHeap) Len() int           { return len(m) }
func (m metricHeap) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m metricHeap) Less(i, j int) bool { return m[i].val < m[j].val }

func (m *metricHeap) Push(x interface{}) {
	*m = append(*m, x.(metricHeapElement))
}

func (m *metricHeap) Pop() interface{} {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[0 : n-1]
	return x
}
