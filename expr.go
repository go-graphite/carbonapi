package main

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
)

// expression parser

type exprType int

const (
	etName exprType = iota
	etFunc
	etConst
	etString
)

type expr struct {
	target    string
	etype     exprType
	val       float64
	valStr    string
	args      []*expr
	argString string
}

type metricRequest struct {
	metric string
	from   int32
	until  int32
}

func (e *expr) metrics() []metricRequest {

	switch e.etype {
	case etName:
		return []metricRequest{{metric: e.target}}
	case etConst, etString:
		return nil
	case etFunc:
		var r []metricRequest
		for _, a := range e.args {
			r = append(r, a.metrics()...)
		}
		return r
	}

	return nil
}

func parseExpr(e string) (*expr, string, error) {

	if '0' <= e[0] && e[0] <= '9' {
		val, e, err := parseConst(e)
		return &expr{val: val, etype: etConst}, e, err
	}

	if e[0] == '\'' || e[0] == '"' {
		val, e, err := parseString(e)
		return &expr{valStr: val, etype: etString}, e, err
	}

	name, e := parseName(e)

	if name == "" {
		return nil, "", ErrMissingArgument
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
			return "", nil, "", err
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
		r == '.' || r == '_' || r == '-' || r == '*' || r == ':'
}

func isDigit(r byte) bool {
	return '0' <= r && r <= '9'
}

func parseConst(s string) (float64, string, error) {

	var i int
	for i < len(s) && (isDigit(s[i]) || s[i] == '.') {
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
	for i < len(s) && isNameChar(s[i]) {
		i++
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

func getFloatArg(e *expr, n int) (float64, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
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
	case "false":
		return false, nil
	case "true":
		return true, nil
	}

	return false, ErrBadType
}

func getSeriesArg(arg *expr, values map[metricRequest][]*pb.FetchResponse) ([]*pb.FetchResponse, error) {

	if arg.etype != etName && arg.etype != etFunc {
		return nil, ErrMissingTimeseries
	}
	a := evalExpr(arg, values)

	if len(a) == 0 {
		return nil, ErrMissingTimeseries
	}

	return a, nil
}

func getSeriesArgs(e []*expr, values map[metricRequest][]*pb.FetchResponse) ([]*pb.FetchResponse, error) {

	var args []*pb.FetchResponse

	for _, arg := range e {
		a, err := getSeriesArg(arg, values)
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

func evalExpr(e *expr, values map[metricRequest][]*pb.FetchResponse) []*pb.FetchResponse {

	// TODO(dgryski): group highestAverage exclude timeShift stdev transformNull

	switch e.etype {
	case etName:
		return values[metricRequest{metric: e.target}]
	case etConst:
		p := pb.FetchResponse{Name: proto.String(e.target), Values: []float64{e.val}}
		return []*pb.FetchResponse{&p}
	}

	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.args) == 0 {
		return nil
	}

	switch e.target {
	case "alias":
		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		alias, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(alias),
			Values:    arg[0].Values,
			IsAbsent:  arg[0].IsAbsent,
			StepTime:  arg[0].StepTime,
			StartTime: arg[0].StartTime,
			StopTime:  arg[0].StopTime,
		}

		return []*pb.FetchResponse{&r}

	case "aliasByNode":
		args, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		field, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range args {

			fields := strings.Split(*a.Name, ".")
			if len(fields) < field {
				continue
			}
			r := pb.FetchResponse{
				Name:      proto.String(fields[field]),
				Values:    a.Values,
				IsAbsent:  a.IsAbsent,
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}
			results = append(results, &r)
		}

		return results

	case "avg", "averageSeries":
		args, err := getSeriesArgs(e.args, values)
		if err != nil {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("averageSeries(%s)", e.argString)),
			Values:    make([]float64, len(args[0].Values)),
			IsAbsent:  make([]bool, len(args[0].Values)),
			StepTime:  args[0].StepTime,
			StartTime: args[0].StartTime,
			StopTime:  args[0].StopTime,
		}

		// TODO(dgryski): make sure all series are the same 'size'
		for i := 0; i < len(args[0].Values); i++ {
			var elts int
			for j := 0; j < len(args); j++ {
				if args[j].IsAbsent[i] {
					continue
				}
				elts++
				r.Values[i] += args[j].Values[i]
			}

			if elts > 0 {
				r.Values[i] /= float64(elts)
			} else {
				r.IsAbsent[i] = true
			}
		}
		return []*pb.FetchResponse{&r}

	case "derivative":
		args, err := getSeriesArgs(e.args, values)
		if err != nil {
			return nil
		}
		var result []*pb.FetchResponse
		for _, a := range args {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("derivative(%s)", *a.Name)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}
			prev := a.Values[0]
			for i, v := range a.Values {
				if i == 0 || a.IsAbsent[i] {
					r.IsAbsent[i] = true
					continue
				}

				r.Values[i] = v - prev
				prev = v
			}
			result = append(result, &r)
		}
		return result

	case "diffSeries":
		if len(e.args) != 2 {
			return nil
		}

		minuend, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}

		subtrahend, err := getSeriesArg(e.args[1], values)
		if err != nil {
			return nil
		}

		if len(minuend) != 1 || len(subtrahend) != 1 {
			return nil
		}

		if *minuend[0].StepTime != *subtrahend[0].StepTime || len(minuend[0].Values) != len(subtrahend[0].Values) {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("diffSeries(%s)", e.argString)),
			Values:    make([]float64, len(minuend[0].Values)),
			IsAbsent:  make([]bool, len(minuend[0].Values)),
			StepTime:  minuend[0].StepTime,
			StartTime: minuend[0].StartTime,
			StopTime:  minuend[0].StopTime,
		}

		for i, v := range minuend[0].Values {

			if minuend[0].IsAbsent[i] || subtrahend[0].IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v - subtrahend[0].Values[i]
		}
		return []*pb.FetchResponse{&r}

	case "divideSeries":
		if len(e.args) != 2 {
			return nil
		}

		numerator, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}

		denominator, err := getSeriesArg(e.args[1], values)
		if err != nil {
			return nil
		}

		if len(numerator) != 1 || len(denominator) != 1 {
			return nil
		}

		if *numerator[0].StepTime != *denominator[0].StepTime || len(numerator[0].Values) != len(denominator[0].Values) {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("divideSeries(%s)", e.argString)),
			Values:    make([]float64, len(numerator[0].Values)),
			IsAbsent:  make([]bool, len(numerator[0].Values)),
			StepTime:  numerator[0].StepTime,
			StartTime: numerator[0].StartTime,
			StopTime:  numerator[0].StopTime,
		}

		for i, v := range numerator[0].Values {

			if numerator[0].IsAbsent[i] || denominator[0].IsAbsent[i] || denominator[0].Values[i] == 0 {
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v / denominator[0].Values[i]
		}
		return []*pb.FetchResponse{&r}

	case "highestMax":

		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		n, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}
		var results []*pb.FetchResponse

		// we have fewer arguments than we want result series
		if len(arg) < n {
			return arg
		}

		var mh metricHeap

		for i, a := range arg {
			m := maxFloat(a.Values)

			if len(mh) < n {
				heap.Push(&mh, metricHeapElement{idx: i, val: m})
				continue
			}
			// m is bigger than smallest max found so far
			if mh[0].val < m {
				mh[0].val = m
				mh[0].idx = i
				heap.Fix(&mh, 0)
			}
		}

		results = make([]*pb.FetchResponse, n)

		// results should be ordered ascending
		for len(mh) > 0 {
			v := heap.Pop(&mh).(metricHeapElement)
			results[len(mh)] = arg[v.idx]
		}

		return results

	case "keepLastValue":
		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		keep, err := getIntArgDefault(e, 1, -1)
		if err != nil {
			return nil
		}
		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("keepLastValue(%s)", e.argString)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}

			prev := math.NaN()
			missing := 0

			for i, v := range a.Values {
				if a.IsAbsent[i] {

					if (keep < 0 || missing < keep) && !math.IsNaN(prev) {
						r.Values[i] = prev
						missing++
					} else {
						r.IsAbsent[i] = true
					}

					continue
				}
				missing = 0
				prev = v
				r.Values[i] = v
			}
			results = append(results, &r)
		}
		return results

	case "maxSeries":
		args, err := getSeriesArgs(e.args, values)
		if err != nil {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("maxSeries(%s)", e.argString)),
			Values:    make([]float64, len(args[0].Values)),
			IsAbsent:  make([]bool, len(args[0].Values)),
			StepTime:  args[0].StepTime,
			StartTime: args[0].StartTime,
			StopTime:  args[0].StopTime,
		}

		// TODO(dgryski): make sure all series are the same 'size'
		for i := 0; i < len(args[0].Values); i++ {
			var elts int
			r.Values[i] = math.Inf(-1)
			for j := 0; j < len(args); j++ {
				if args[j].IsAbsent[i] {
					continue
				}
				elts++
				if r.Values[i] < args[j].Values[i] {
					r.Values[i] = args[j].Values[i]
				}
			}

			if elts == 0 {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		return []*pb.FetchResponse{&r}

	case "movingAverage":
		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		windowSize, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		var result []*pb.FetchResponse

		for _, a := range arg {
			w := &Windowed{data: make([]float64, windowSize)}
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("movingAverage(%s,%d)", *a.Name, windowSize)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}
			for i, v := range a.Values {
				if a.IsAbsent[i] {
					// make sure missing values are ignored
					v = 0
				}
				r.Values[i] = w.Mean()
				if math.IsNaN(r.Values[i]) {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}
				w.Push(v)
			}
			result = append(result, &r)
		}
		return result

	case "nonNegativeDerivative":
		args, err := getSeriesArgs(e.args, values)
		if err != nil {
			return nil
		}

		var result []*pb.FetchResponse
		for _, a := range args {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("nonNegativeDerivative(%s)", *a.Name)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}
			prev := a.Values[0]
			for i, v := range a.Values {
				if i == 0 || a.IsAbsent[i] {
					r.IsAbsent[i] = true
					continue
				}

				r.Values[i] = v - prev
				if r.Values[i] < 0 {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}
				prev = v
			}
			result = append(result, &r)
		}
		return result

	case "scale":
		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		scale, err := getFloatArg(e, 1)
		if err != nil {
			return nil
		}
		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("scale(%s)", e.argString)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = v * scale
			}
			results = append(results, &r)
		}
		return results

	case "scaleToSeconds":
		arg, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}
		seconds, err := getFloatArg(e, 1)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("scaleToSeconds(%s)", e.argString)),
				Values:    make([]float64, len(a.Values)),
				StepTime:  a.StepTime,
				IsAbsent:  make([]bool, len(a.Values)),
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}

			factor := seconds / float64(*a.StepTime)

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = v * factor
			}
			results = append(results, &r)
		}
		return results

	case "sum", "sumSeries":
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArgs(e.args, values)
		if err != nil {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(fmt.Sprintf("sumSeries(%s)", e.argString)),
			Values:    make([]float64, len(args[0].Values)),
			IsAbsent:  make([]bool, len(args[0].Values)),
			StepTime:  args[0].StepTime,
			StartTime: args[0].StartTime,
			StopTime:  args[0].StopTime,
		}
		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}
				r.Values[i] += v
			}
		}
		return []*pb.FetchResponse{&r}
	case "summarize":

		// TODO(dgryski): make sure the arrays are all the same 'size'
		// TODO(dgryski): need to implement alignToFrom=false, and make it the default
		args, err := getSeriesArg(e.args[0], values)
		if err != nil {
			return nil
		}

		bucketSizeStr, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		bucketSize, err := intervalString(bucketSizeStr)
		if err != nil {
			return nil
		}

		summarizeFunction, err := getStringArgDefault(e, 2, "sum")
		if err != nil {
			return nil
		}

		alignToFrom, err := getBoolArgDefault(e, 3, false)
		if err != nil {
			return nil
		}

		start := *args[0].StartTime
		stop := *args[0].StopTime

		if !alignToFrom {
			start = int32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
			stop = int32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
		}

		buckets := (stop - start) / bucketSize

		var results []*pb.FetchResponse

		for _, arg := range args {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("summarize(%s)", e.argString)),
				Values:    make([]float64, buckets, buckets+1),
				IsAbsent:  make([]bool, buckets, buckets+1),
				StepTime:  proto.Int32(bucketSize),
				StartTime: proto.Int32(start),
				StopTime:  proto.Int32(stop),
			}
			bucketStart := *args[0].StartTime // unadjusted
			bucketEnd := *r.StartTime + bucketSize
			values := make([]float64, 0, bucketSize / *arg.StepTime)
			t := bucketStart
			ridx := 0
			skipped := 0
			for i, v := range arg.Values {

				if !arg.IsAbsent[i] {
					values = append(values, v)
				} else {
					skipped++
				}

				t += *arg.StepTime

				if t >= bucketEnd {
					rv := summarizeValues(summarizeFunction, values)

					r.Values[ridx] = rv
					ridx++
					bucketStart += bucketSize
					bucketEnd += bucketSize
					values = values[:0]
					skipped = 0
				}
			}

			// remaining values
			if len(values) > 0 {
				rv := summarizeValues(summarizeFunction, values)
				r.Values = append(r.Values, rv)
				r.IsAbsent = append(r.IsAbsent, false)
			}

			results = append(results, &r)
		}
		return results
	}

	log.Printf("unknown function in evalExpr:  %q\n", e.target)

	return nil
}

func summarizeValues(f string, values []float64) float64 {
	rv := 0.0
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
	}

	return rv
}

// From github.com/dgryski/go-onlinestats
// Copied here because we don't need the rest of the package, and we only need
// a small part of this type

type Windowed struct {
	data   []float64
	head   int
	length int
	sum    float64
}

func (w *Windowed) Push(n float64) {
	old := w.data[w.head]

	w.length++

	w.data[w.head] = n
	w.head++
	if w.head >= len(w.data) {
		w.head = 0
	}

	w.sum -= old
	w.sum += n
}

func (w *Windowed) Len() int {
	if w.length < len(w.data) {
		return w.length
	}

	return len(w.data)
}

func (w *Windowed) Mean() float64 { return w.sum / float64(w.Len()) }

func maxFloat(f64s []float64) float64 {
	m := math.Inf(-1)
	for _, v := range f64s {
		if v > m {
			m = v
		}
	}
	return m
}

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
