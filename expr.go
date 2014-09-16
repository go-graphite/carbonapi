package main

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
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

		switch e.target {
		case "movingAverage":
			if len(e.args) != 2 {
				return nil
			}

			var n int32
			var err error

			switch e.args[1].etype {
			case etConst:
				var nint int
				nint, err = getIntArg(e, 1)
				n = int32(nint)
			case etString:
				n, err = getIntervalArg(e, 1, 1)
			default:
				err = ErrBadType
			}
			if err != nil {
				return nil
			}

			for i := range r {
				r[i].from -= n
			}
		case "timeShift":
			offs, err := getIntervalArg(e, 1, -1)
			if err != nil {
				return nil
			}
			for i := range r {
				r[i].from += offs
				r[i].until += offs
			}
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

func getSeriesArg(arg *expr, from, until int32, values map[metricRequest][]*pb.FetchResponse) ([]*pb.FetchResponse, error) {

	if arg.etype != etName && arg.etype != etFunc {
		return nil, ErrMissingTimeseries
	}
	a := evalExpr(arg, from, until, values)

	if len(a) == 0 {
		return nil, ErrMissingTimeseries
	}

	return a, nil
}

func getSeriesArgs(e []*expr, from, until int32, values map[metricRequest][]*pb.FetchResponse) ([]*pb.FetchResponse, error) {

	var args []*pb.FetchResponse

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

func evalExpr(e *expr, from, until int32, values map[metricRequest][]*pb.FetchResponse) []*pb.FetchResponse {

	// TODO(dgryski): stdev

	switch e.etype {
	case etName:
		return values[metricRequest{metric: e.target, from: from, until: until}]
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
	case "alias": // alias(seriesList, newName)
		arg, err := getSeriesArg(e.args[0], from, until, values)
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

	case "aliasByNode": // aliasByNode(seriesList, *nodes)
		// TODO(dgryski): we only support one 'node' argument at the moment
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		field, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range args {

			metric := extractMetric(*a.Name)
			fields := strings.Split(metric, ".")
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

	case "avg", "averageSeries": // averageSeries(*seriesLists)
		args, err := getSeriesArgs(e.args, from, until, values)
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

	case "derivative": // derivative(seriesList)
		args, err := getSeriesArg(e.args[0], from, until, values)
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

	case "diffSeries": // diffSeries(*seriesLists)
		// FIXME(dgryski): only accepts two arguments
		if len(e.args) != 2 {
			return nil
		}

		minuend, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		subtrahend, err := getSeriesArg(e.args[1], from, until, values)
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

	case "divideSeries": // divideSeries(dividendSeriesList, divisorSeriesList)
		if len(e.args) != 2 {
			return nil
		}

		numerator, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		denominator, err := getSeriesArg(e.args[1], from, until, values)
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

	case "exclude": // exclude(seriesList, pattern)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		pat, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		patre, err := regexp.Compile(pat)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range arg {
			if !patre.MatchString(a.GetName()) {
				results = append(results, a)
			}
		}

		return results

	case "grep": // grep(seriesList, pattern)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		pat, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		patre, err := regexp.Compile(pat)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range arg {
			if patre.MatchString(a.GetName()) {
				results = append(results, a)
			}
		}

		return results

	case "group": // group(*seriesLists)
		args, err := getSeriesArgs(e.args, from, until, values)
		if err != nil {
			return nil
		}

		return args

	case "highestAverage", "highestCurrent", "highestMax": // highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)

		arg, err := getSeriesArg(e.args[0], from, until, values)
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

		var compute func([]float64, []bool) float64

		switch e.target {
		case "highestMax":
			compute = maxValue
		case "highestAverage":
			compute = avgValue
		case "highestCurrent":
			compute = currentValue
		}

		for i, a := range arg {
			m := compute(a.Values, a.IsAbsent)

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

	case "keepLastValue": // keepLastValue(seriesList, limit=inf)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		keep, err := getIntArgDefault(e, 1, -1)
		if err != nil {
			return nil
		}
		var results []*pb.FetchResponse

		for _, a := range arg {
			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("keepLastValue(%s)", *a.Name)
			} else {
				name = fmt.Sprintf("keepLastValue(%s,%d)", *a.Name, keep)
			}

			r := pb.FetchResponse{
				Name:      proto.String(name),
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

	case "logarithm": // logarithm(seriesList, base=10)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		base, err := getIntArgDefault(e, 1, 10)
		if err != nil {
			return nil
		}
		baseLog := math.Log(float64(base))

		var results []*pb.FetchResponse

		for _, a := range arg {

			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("logarithm(%s)", *a.Name)
			} else {
				name = fmt.Sprintf("logarithm(%s,%d)", *a.Name, base)
			}

			r := pb.FetchResponse{
				Name:      proto.String(name),
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
				r.Values[i] = math.Log(v) / baseLog
			}
			results = append(results, &r)
		}
		return results

	case "maxSeries": // maxSeries(*seriesLists)
		args, err := getSeriesArgs(e.args, from, until, values)
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

	case "movingAverage": // movingAverage(seriesList, windowSize)
		var n int
		var err error

		var scaleByStep bool

		switch e.args[1].etype {
		case etConst:
			n, err = getIntArg(e, 1)
		case etString:
			var n32 int32
			n32, err = getIntervalArg(e, 1, 1)
			n = int(n32)
			scaleByStep = true
		default:
			err = ErrBadType
		}
		if err != nil {
			return nil
		}

		windowSize := n

		arg, err := getSeriesArg(e.args[0], from-int32(windowSize), until, values)
		if err != nil {
			return nil
		}

		if scaleByStep {
			windowSize /= int(*arg[0].StepTime)
		}

		var result []*pb.FetchResponse

		for _, a := range arg {
			w := &Windowed{data: make([]float64, windowSize)}
			sz := (until - from) / *a.StepTime
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("movingAverage(%s,%d)", *a.Name, windowSize)),
				Values:    make([]float64, sz),
				IsAbsent:  make([]bool, sz),
				StepTime:  a.StepTime,
				StartTime: proto.Int32(from),
				StopTime:  proto.Int32(until),
			}
			ridx := 0
			for i, v := range a.Values {
				if a.IsAbsent[i] {
					// make sure missing values are ignored
					v = 0
				}
				if i > windowSize {
					r.Values[ridx] = w.Mean()
					if math.IsNaN(r.Values[ridx]) {
						r.Values[ridx] = 0
						r.IsAbsent[ridx] = true
					}
					ridx++
				}
				w.Push(v)
			}
			result = append(result, &r)
		}
		return result

	case "nonNegativeDerivative": // nonNegativeDerivative(seriesList, maxValue=None)
		// FIXME(dgryski): support maxValue
		args, err := getSeriesArg(e.args[0], from, until, values)
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

	case "scale": // scale(seriesList, factor)
		arg, err := getSeriesArg(e.args[0], from, until, values)
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
				Name:      proto.String(fmt.Sprintf("scale(%s,%g)", *a.Name, scale)),
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

	case "scaleToSeconds": // scaleToSeconds(seriesList, seconds)
		arg, err := getSeriesArg(e.args[0], from, until, values)
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
				Name:      proto.String(fmt.Sprintf("scaleToSeconds(%s,%d)", *a.Name, int(seconds))),
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

	case "sum", "sumSeries": // sumSeries(*seriesLists)
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArgs(e.args, from, until, values)
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

	case "summarize": // summarize(seriesList, intervalString, func='sum', alignToFrom=False
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		bucketSize, err := getIntervalArg(e, 1, 1)
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

	case "timeShift": // timeShift(seriesList, timeShift, resetEnd=True)
		// FIXME(dgryski): support resetEnd=true

		offs, err := getIntervalArg(e, 1, -1)
		if err != nil {
			return nil
		}

		arg, err := getSeriesArg(e.args[0], from+offs, until+offs, values)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("timeShift(%s)", *a.Name)),
				Values:    a.Values,
				IsAbsent:  a.IsAbsent,
				StepTime:  a.StepTime,
				StartTime: proto.Int32(*a.StartTime + offs),
				StopTime:  proto.Int32(*a.StopTime + offs),
			}

			results = append(results, &r)
		}
		return results

	case "transformNull": // transformNull(seriesList, default=0)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		defv, err := getFloatArgDefault(e, 1, 0)
		if err != nil {
			return nil
		}
		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("transformNull(%s)", e.argString)),
				Values:    make([]float64, len(a.Values)),
				IsAbsent:  make([]bool, len(a.Values)),
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
			}

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					v = defv
				}

				r.Values[i] = v
			}
			results = append(results, &r)
		}
		return results

	case "drawAsInfinite": // ignored
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		var results []*pb.FetchResponse

		for _, a := range arg {
			r := pb.FetchResponse{
				Name:      proto.String(fmt.Sprintf("%s(%s)", e.target, *a.Name)),
				Values:    a.Values,
				IsAbsent:  a.IsAbsent,
				StepTime:  a.StepTime,
				StartTime: a.StartTime,
				StopTime:  a.StopTime,
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

func extractMetric(m string) string {

	// search for a metric name in `m'
	// metric name is defined to be a series of name characters terminated by a comma

	start := 0
	end := 0
	for end < len(m) {
		if !isNameChar(m[end]) {
			if m[end] == ',' || m[end] == ')' {
				return m[start:end]
			}
			start = end + 1
		}

		end++
	}

	return m[start:end]
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

func currentValue(f64s []float64, absent []bool) float64 {

	for i := len(f64s) - 1; i >= 0; i-- {
		if !absent[i] {
			return f64s[i]
		}
	}

	return math.NaN()
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
