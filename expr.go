package main

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
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

	// skip whitespace
	for e[0] == ' ' {
		e = e[1:]
	}

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

	switch e.target {
	case "absolute": // absolute(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			for i, v := range a.Values {
				if a.IsAbsent[i] {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = math.Abs(v)
			}
			return r
		})

	case "alias": // alias(seriesList, newName)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		alias, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		r := *arg[0]
		r.Name = proto.String(alias)
		return []*metricData{&r}

	case "aliasByMetric": // aliasByMetric(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			metric := extractMetric(a.GetName())
			part := strings.Split(metric, ".")
			r.Name = proto.String(part[len(part)-1])
			r.Values = a.Values
			r.IsAbsent = a.IsAbsent
			return r
		})

	case "aliasByNode": // aliasByNode(seriesList, *nodes)
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		fields, err := getIntArgs(e, 1)
		if err != nil {
			return nil
		}

		var results []*metricData

		for _, a := range args {

			metric := extractMetric(a.GetName())
			nodes := strings.Split(metric, ".")

			var name []string
			for _, f := range fields {
				if f >= len(nodes) {
					continue
				}
				name = append(name, nodes[f])
			}

			r := *a
			r.Name = proto.String(strings.Join(name, "."))
			results = append(results, &r)
		}

		return results

	case "aliasSub": // aliasSub(seriesList, search, replace)
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		search, err := getStringArg(e, 1)
		if err != nil {
			return nil
		}

		replace, err := getStringArg(e, 2)
		if err != nil {
			return nil
		}

		re, err := regexp.Compile(search)
		if err != nil {
			return nil
		}

		var results []*metricData

		for _, a := range args {
			metric := extractMetric(a.GetName())

			r := *a
			r.Name = proto.String(re.ReplaceAllString(metric, replace))
			results = append(results, &r)
		}

		return results

	case "asPercent": // asPercent(seriesList, total=None)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		var getTotal func(i int) float64
		var formatName func(a *metricData) string

		if len(e.args) == 1 {
			getTotal = func(i int) float64 {
				var t float64
				var atLeastOne bool
				for _, a := range arg {
					if a.IsAbsent[i] {
						continue
					}
					atLeastOne = true
					t += a.Values[i]
				}
				if !atLeastOne {
					t = math.NaN()
				}

				return t
			}
			formatName = func(a *metricData) string {
				return fmt.Sprintf("asPercent(%s)", a.GetName())
			}
		} else if len(e.args) == 2 && e.args[1].etype == etConst {
			total, err := getFloatArg(e, 1)
			if err != nil {
				return nil
			}
			getTotal = func(i int) float64 { return total }
			formatName = func(a *metricData) string {
				return fmt.Sprintf("asPercent(%s,%g)", a.GetName(), total)
			}
		} else if len(e.args) == 2 && (e.args[1].etype == etName || e.args[1].etype == etFunc) {
			total, err := getSeriesArg(e.args[1], from, until, values)
			if err != nil || len(total) != 1 {
				return nil
			}
			getTotal = func(i int) float64 {
				if total[0].IsAbsent[i] {
					return math.NaN()
				}
				return total[0].Values[i]
			}
			var totalString string
			if e.args[1].etype == etName {
				totalString = e.args[1].target
			} else {
				totalString = fmt.Sprintf("%s(%s)", e.args[1].target, e.args[1].argString)
			}
			formatName = func(a *metricData) string {
				return fmt.Sprintf("asPercent(%s,%s)", a.GetName(), totalString)
			}
		} else {
			return nil
		}

		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(formatName(a))
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))
			results = append(results, &r)
		}

		for i := range results[0].Values {

			total := getTotal(i)

			for j := range results {
				r := results[j]
				a := arg[j]

				if a.IsAbsent[i] || math.IsNaN(total) {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}

				r.Values[i] = (a.Values[i] / total) * 100
			}
		}
		return results

	case "avg", "averageSeries": // averageSeries(*seriesLists)
		args, err := getSeriesArgs(e.args, from, until, values)
		if err != nil {
			return nil
		}

		r := *args[0]
		r.Name = proto.String(fmt.Sprintf("averageSeries(%s)", e.argString))
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

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
		return []*metricData{&r}

	case "averageAbove", "averageBelow", "currentAbove", "currentBelow": // averageAbove(seriesList, n), averageBelow(seriesList, n), currentAbove(seriesList, n), currentBelow(seriesList, n)
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		n, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		var results []*metricData
		for _, a := range args {
			switch e.target {
			case "averageAbove":
				if avgValue(a.Values, a.IsAbsent) >= float64(n) {
					results = append(results, a)
				}
			case "averageBelow":
				if avgValue(a.Values, a.IsAbsent) <= float64(n) {
					results = append(results, a)
				}
			case "currentAbove":
				if currentValue(a.Values, a.IsAbsent) >= float64(n) {
					results = append(results, a)
				}
			case "currentBelow":
				if currentValue(a.Values, a.IsAbsent) <= float64(n) {
					results = append(results, a)
				}
			}
		}
		return results

	case "derivative": // derivative(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			prev := a.Values[0]
			for i, v := range a.Values {
				if i == 0 || a.IsAbsent[i] {
					r.IsAbsent[i] = true
					continue
				}

				r.Values[i] = v - prev
				prev = v
			}
			return r
		})

	case "diffSeries": // diffSeries(*seriesLists)
		if len(e.args) < 2 {
			return nil
		}

		minuend, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		subtrahends, err := getSeriesArgs(e.args[1:], from, until, values)
		if err != nil {
			return nil
		}

		// FIXME: need more error checking on minuend, subtrahends here
		r := *minuend[0]
		r.Name = proto.String(fmt.Sprintf("diffSeries(%s)", e.argString))
		r.Values = make([]float64, len(minuend[0].Values))
		r.IsAbsent = make([]bool, len(minuend[0].Values))

		for i, v := range minuend[0].Values {

			if minuend[0].IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			}

			var sub float64
			var atLeastOne bool
			for _, s := range subtrahends {
				if s.IsAbsent[i] {
					continue
				}
				atLeastOne = true
				sub += s.Values[i]
			}

			if atLeastOne {
				r.Values[i] = v - sub
			} else {
				r.IsAbsent[i] = true
			}
		}
		return []*metricData{&r}

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

		if numerator[0].GetStepTime() != denominator[0].GetStepTime() || len(numerator[0].Values) != len(denominator[0].Values) {
			return nil
		}

		r := *numerator[0]
		r.Name = proto.String(fmt.Sprintf("divideSeries(%s)", e.argString))
		r.Values = make([]float64, len(numerator[0].Values))
		r.IsAbsent = make([]bool, len(numerator[0].Values))

		for i, v := range numerator[0].Values {

			if numerator[0].IsAbsent[i] || denominator[0].IsAbsent[i] || denominator[0].Values[i] == 0 {
				r.IsAbsent[i] = true
				continue
			}

			r.Values[i] = v / denominator[0].Values[i]
		}
		return []*metricData{&r}

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

		var results []*metricData

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

		var results []*metricData

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

	case "groupByNode": // groupByNode(seriesList, nodeNum, callback)
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		field, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		callback, err := getStringArg(e, 2)
		if err != nil {
			return nil
		}

		var results []*metricData

		groups := make(map[string][]*metricData)

		for _, a := range args {

			metric := extractMetric(a.GetName())
			nodes := strings.Split(metric, ".")
			node := nodes[field]

			groups[node] = append(groups[node], a)
		}

		for k, v := range groups {

			// create a stub context to evaluate the callback in
			nexpr, _, err := parseExpr(fmt.Sprintf("%s(%s)", callback, k))
			if err != nil {
				return nil
			}

			nvalues := map[metricRequest][]*metricData{
				metricRequest{k, from, until}: v,
			}

			r := evalExpr(nexpr, from, until, nvalues)
			if r != nil {
				results = append(results, r...)
			}
		}

		return results

	case "lowestAverage", "lowestCurrent": // lowestAverage(seriesList, n) , lowestCurrent(seriesList, n)

		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		n, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}
		var results []*metricData

		// we have fewer arguments than we want result series
		if len(arg) < n {
			return arg
		}

		var mh metricHeap

		var compute func([]float64, []bool) float64

		switch e.target {
		case "lowestAverage":
			compute = avgValue
		case "lowestCurrent":
			compute = currentValue
		}

		for i, a := range arg {
			m := compute(a.Values, a.IsAbsent)
			heap.Push(&mh, metricHeapElement{idx: i, val: m})
		}

		results = make([]*metricData, n)

		// results should be ordered ascending
		for i := 0; i < n; i++ {
			v := heap.Pop(&mh).(metricHeapElement)
			results[i] = arg[v.idx]
		}

		return results

	case "highestAverage", "highestCurrent", "highestMax": // highestAverage(seriesList, n) , highestCurrent(seriesList, n), highestMax(seriesList, n)

		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		n, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}
		var results []*metricData

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

		results = make([]*metricData, n)

		// results should be ordered ascending
		for len(mh) > 0 {
			v := heap.Pop(&mh).(metricHeapElement)
			results[len(mh)] = arg[v.idx]
		}

		return results

	case "hitcount": // hitcount(seriesList, intervalString, alignToInterval=False)
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		bucketSize, err := getIntervalArg(e, 1, 1)
		if err != nil {
			return nil
		}

		alignToInterval, err := getBoolArgDefault(e, 3, false)
		if err != nil {
			return nil
		}

		start := args[0].GetStartTime()
		stop := args[0].GetStopTime()

		if alignToInterval {
			start = int32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
			stop = int32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
		}

		buckets := (stop - start) / bucketSize

		var results []*metricData

		for _, arg := range args {

			var name string
			switch len(e.args) {
			case 2:
				name = fmt.Sprintf("hitcount(%s,'%s')", arg.GetName(), e.args[1].valStr)
			case 3:
				name = fmt.Sprintf("hitcount(%s,'%s',%s)", arg.GetName(), e.args[1].valStr, e.args[2].target)
			}

			r := metricData{FetchResponse: pb.FetchResponse{
				Name:      proto.String(name),
				Values:    make([]float64, buckets, buckets+1),
				IsAbsent:  make([]bool, buckets, buckets+1),
				StepTime:  proto.Int32(bucketSize),
				StartTime: proto.Int32(start),
				StopTime:  proto.Int32(stop),
			}}

			bucketStart := args[0].GetStartTime() // unadjusted
			bucketEnd := r.GetStartTime() + bucketSize
			values := make([]float64, 0, bucketSize/arg.GetStepTime())
			t := bucketStart
			ridx := 0
			skipped := 0
			var count float64
			for i, v := range arg.Values {

				if !arg.IsAbsent[i] {
					values = append(values, v)
				} else {
					skipped++
				}

				t += arg.GetStepTime()

				count += v * float64(arg.GetStepTime())

				if t >= bucketEnd {
					rv := count

					if math.IsNaN(rv) {
						rv = 0
						r.IsAbsent[ridx] = true
					}

					r.Values[ridx] = rv
					ridx++
					bucketStart += bucketSize
					bucketEnd += bucketSize
					values = values[:0]
					skipped = 0
					count = 0
				}
			}

			// remaining values
			if len(values) > 0 {
				rv := count
				r.Values = append(r.Values, rv)
				r.IsAbsent = append(r.IsAbsent, false)
			}

			results = append(results, &r)
		}
		return results
	case "integral": // integral(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			current := 0.0
			for i, v := range a.Values {
				if a.IsAbsent[i] || v == 0 {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				current += v
				r.Values[i] = current
			}
			return r
		})

	case "invert": // invert(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			for i, v := range a.Values {
				if a.IsAbsent[i] || v == 0 {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = 1 / v
			}
			return r
		})

	case "keepLastValue": // keepLastValue(seriesList, limit=inf)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		keep, err := getIntArgDefault(e, 1, -1)
		if err != nil {
			return nil
		}
		var results []*metricData

		for _, a := range arg {
			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("keepLastValue(%s)", a.GetName())
			} else {
				name = fmt.Sprintf("keepLastValue(%s,%d)", a.GetName(), keep)
			}

			r := *a
			r.Name = proto.String(name)
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

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

	case "logarithm", "log": // logarithm(seriesList, base=10)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		base, err := getIntArgDefault(e, 1, 10)
		if err != nil {
			return nil
		}
		baseLog := math.Log(float64(base))

		var results []*metricData

		for _, a := range arg {

			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("logarithm(%s)", a.GetName())
			} else {
				name = fmt.Sprintf("logarithm(%s,%d)", a.GetName(), base)
			}

			r := *a
			r.Name = proto.String(name)
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

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

		r := *args[0]
		r.Name = proto.String(fmt.Sprintf("maxSeries(%s)", e.argString))
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		// TODO(dgryski): make sure all series are the same 'size'
		for i := 0; i < len(args[0].Values); i++ {
			var atLeastOne bool
			r.Values[i] = math.Inf(-1)
			for j := 0; j < len(args); j++ {
				if args[j].IsAbsent[i] {
					continue
				}
				atLeastOne = true
				if r.Values[i] < args[j].Values[i] {
					r.Values[i] = args[j].Values[i]
				}
			}

			if !atLeastOne {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		return []*metricData{&r}

	case "minSeries": // minSeries(*seriesLists)
		args, err := getSeriesArgs(e.args, from, until, values)
		if err != nil {
			return nil
		}

		r := *args[0]
		r.Name = proto.String(fmt.Sprintf("minSeries(%s)", e.argString))
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		// TODO(dgryski): make sure all series are the same 'size'
		for i := 0; i < len(args[0].Values); i++ {
			var atLeastOne bool
			r.Values[i] = math.Inf(1)
			for j := 0; j < len(args); j++ {
				if args[j].IsAbsent[i] {
					continue
				}
				atLeastOne = true
				if r.Values[i] > args[j].Values[i] {
					r.Values[i] = args[j].Values[i]
				}
			}

			if !atLeastOne {
				r.Values[i] = 0
				r.IsAbsent[i] = true
			}
		}
		return []*metricData{&r}
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

		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		if scaleByStep {
			windowSize /= int(arg[0].GetStepTime())
		}

		var result []*metricData

		for _, a := range arg {
			w := &Windowed{data: make([]float64, windowSize)}

			r := *a
			r.Name = proto.String(fmt.Sprintf("movingAverage(%s,%d)", a.GetName(), windowSize))
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))
			r.StartTime = proto.Int32(from)
			r.StopTime = proto.Int32(until)

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					// make sure missing values are ignored
					v = math.NaN()
				}
				r.Values[i] = w.Mean()
				w.Push(v)
				if i < windowSize || math.IsNaN(r.Values[i]) {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}
			}
			result = append(result, &r)
		}
		return result

	case "nonNegativeDerivative": // nonNegativeDerivative(seriesList, maxValue=None)
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		maxValue, err := getFloatArgDefault(e, 1, math.NaN())
		if err != nil {
			return nil
		}

		var result []*metricData
		for _, a := range args {
			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("nonNegativeDerivative(%s)", a.GetName())
			} else {
				name = fmt.Sprintf("nonNegativeDerivative(%s,%g)", a.GetName(), maxValue)
			}

			r := *a
			r.Name = proto.String(name)
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

			prev := a.Values[0]
			for i, v := range a.Values {
				if i == 0 || a.IsAbsent[i] || a.IsAbsent[i-1] {
					r.IsAbsent[i] = true
					prev = v
					continue
				}
				diff := v - prev
				if diff >= 0 {
					r.Values[i] = diff
				} else if !math.IsNaN(maxValue) && maxValue >= v {
					r.Values[i] = ((maxValue - prev) + v + 1)
				} else {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}
				prev = v
			}
			result = append(result, &r)
		}
		return result

	case "offsetToZero": // offsetToZero(seriesList)
		return forEachSeriesDo(e, from, until, values, func(a *metricData, r *metricData) *metricData {
			minimum := math.Inf(1)
			for i, v := range a.Values {
				if !a.IsAbsent[i] && v < minimum {
					minimum = v
				}
			}
			for i, v := range a.Values {
				if a.IsAbsent[i] {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}
				r.Values[i] = v - minimum
			}
			return r
		})
	case "scale": // scale(seriesList, factor)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}
		scale, err := getFloatArg(e, 1)
		if err != nil {
			return nil
		}
		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(fmt.Sprintf("scale(%s,%g)", a.GetName(), scale))
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

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

		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(fmt.Sprintf("scaleToSeconds(%s,%d)", a.GetName(), int(seconds)))
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

			factor := seconds / float64(a.GetStepTime())

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

	case "stdev", "stddev": // stdev(seriesList, points, missingThreshold=0.1)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		points, err := getIntArg(e, 1)
		if err != nil {
			return nil
		}

		missingThreshold, err := getFloatArgDefault(e, 2, 0.1)
		if err != nil {
			return nil
		}

		minLen := int((1 - missingThreshold) * float64(points))

		var result []*metricData

		for _, a := range arg {
			w := &Windowed{data: make([]float64, points)}

			r := *a
			r.Name = proto.String(fmt.Sprintf("stdev(%s,%d)", a.GetName(), points))
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					// make sure missing values are ignored
					v = math.NaN()
				}
				w.Push(v)
				r.Values[i] = w.Stdev()
				if math.IsNaN(r.Values[i]) || (i >= minLen && w.Len() < minLen) {
					r.Values[i] = 0
					r.IsAbsent[i] = true
				}
			}
			result = append(result, &r)
		}
		return result

	case "sum", "sumSeries": // sumSeries(*seriesLists)
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArgs(e.args, from, until, values)
		if err != nil {
			return nil
		}

		r := *args[0]
		r.Name = proto.String(fmt.Sprintf("sumSeries(%s)", e.argString))
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}
				atLeastOne[i] = true
				r.Values[i] += v
			}
		}
		for i, v := range atLeastOne {
			if !v {
				r.IsAbsent[i] = true
			}
		}
		return []*metricData{&r}

	case "sumSeriesWithWildcards": // sumSeriesWithWildcards(seriesList, *position)
		// TODO(dgryski): make sure the arrays are all the same 'size'
		args, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		fields, err := getIntArgs(e, 1)
		if err != nil {
			return nil
		}

		var results []*metricData

		groups := make(map[string][]*metricData)

		for _, a := range args {
			metric := extractMetric(a.GetName())
			nodes := strings.Split(metric, ".")
			var s []string
			// Yes, this is O(n^2), but len(nodes) < 10 and len(fields) < 3
			// Iterating an int slice is faster than a map for n ~ 30
			// http://www.antoine.im/posts/someone_is_wrong_on_the_internet
			for i, n := range nodes {
				if !contains(fields, i) {
					s = append(s, n)
				}
			}

			node := strings.Join(s, ".")

			groups[node] = append(groups[node], a)
		}

		for series, args := range groups {
			r := *args[0]
			r.Name = proto.String(fmt.Sprintf("sumSeriesWithWildcards(%s)", series))
			r.Values = make([]float64, len(args[0].Values))
			r.IsAbsent = make([]bool, len(args[0].Values))

			atLeastOne := make([]bool, len(args[0].Values))
			for _, arg := range args {
				for i, v := range arg.Values {
					if arg.IsAbsent[i] {
						continue
					}
					atLeastOne[i] = true
					r.Values[i] += v
				}
			}

			for i, v := range atLeastOne {
				if !v {
					r.IsAbsent[i] = true
				}
			}

			results = append(results, &r)
		}
		return results

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

		start := args[0].GetStartTime()
		stop := args[0].GetStopTime()

		if !alignToFrom {
			start = int32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
			stop = int32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
			// check if a partial bucket is needed
			if stop != args[0].GetStopTime() {
				stop += bucketSize
			}
		} else {
			// adjust for partial buckets
			stop += (stop - start) % bucketSize
		}

		buckets := (stop - start) / bucketSize

		var results []*metricData

		for _, arg := range args {

			var name string
			switch len(e.args) {
			case 2:
				name = fmt.Sprintf("summarize(%s,'%s')", arg.GetName(), e.args[1].valStr)
			case 3:
				name = fmt.Sprintf("summarize(%s,'%s','%s')", arg.GetName(), e.args[1].valStr, e.args[2].valStr)
			case 4:
				name = fmt.Sprintf("summarize(%s,'%s','%s',%s)", arg.GetName(), e.args[1].valStr, e.args[2].valStr, e.args[3].target)
			}

			r := metricData{FetchResponse: pb.FetchResponse{
				Name:      proto.String(name),
				Values:    make([]float64, buckets, buckets),
				IsAbsent:  make([]bool, buckets, buckets),
				StepTime:  proto.Int32(bucketSize),
				StartTime: proto.Int32(start),
				StopTime:  proto.Int32(stop),
			}}
			t := args[0].GetStartTime() // unadjusted
			bucketEnd := r.GetStartTime() + bucketSize
			values := make([]float64, 0, bucketSize/arg.GetStepTime())
			ridx := 0
			skipped := 0
			for i, v := range arg.Values {

				if !arg.IsAbsent[i] {
					values = append(values, v)
				} else {
					skipped++
				}

				t += arg.GetStepTime()

				if t >= bucketEnd {
					rv := summarizeValues(summarizeFunction, values)

					if math.IsNaN(rv) {
						rv = 0
						r.IsAbsent[ridx] = true
					}

					r.Values[ridx] = rv
					ridx++
					bucketEnd += bucketSize
					values = values[:0]
					skipped = 0
				}
			}

			// last partial bucket
			if t < stop {

				rv := summarizeValues(summarizeFunction, values)
				if math.IsNaN(rv) {
					r.Values[ridx] = 0
					r.IsAbsent[ridx] = true
				} else {
					r.Values[ridx] = rv
					r.IsAbsent[ridx] = false
				}

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

		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(fmt.Sprintf("timeShift(%s)", a.GetName()))
			r.StartTime = proto.Int32(a.GetStartTime() - offs)
			r.StopTime = proto.Int32(a.GetStopTime() - offs)
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
		var results []*metricData

		for _, a := range arg {

			var name string
			if len(e.args) == 1 {
				name = fmt.Sprintf("transformNull(%s)", a.GetName())
			} else {
				name = fmt.Sprintf("transformNull(%s,%g)", a.GetName(), defv)
			}

			r := *a
			r.Name = proto.String(name)
			r.Values = make([]float64, len(a.Values))
			r.IsAbsent = make([]bool, len(a.Values))

			for i, v := range a.Values {
				if a.IsAbsent[i] {
					v = defv
				}

				r.Values[i] = v
			}

			results = append(results, &r)
		}
		return results

	case "color": // color(seriesList, theColor) ignored
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		color, err := getStringArg(e, 1) // get color
		if err != nil {
			return nil
		}

		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
			r.color = color

			results = append(results, &r)
		}

		return results

	case "dashed", "drawAsInfinite", "secondYAxis": // ignored
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		var results []*metricData

		for _, a := range arg {
			r := *a
			r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))

			switch e.target {
			case "dashed":
				r.dashed = true
			case "drawAsInfinite":
				r.drawAsInfinite = true
			case "secondYAxis":
				r.secondYAxis = true
			}

			results = append(results, &r)
		}
		return results

	case "limit": // limit(seriesList, n)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		limit, err := getIntArg(e, 1) // get limit
		if err != nil {
			return nil
		}

		return arg[0:limit]

	case "sortByTotal", "sortByName", "sortByMaxima", "sortByMinima": // sortByTotal(seriesList), sortByName(seriesList), sortByMaxima(seriesList), sortByMinima(seriesList)
		arg, err := getSeriesArg(e.args[0], from, until, values)
		if err != nil {
			return nil
		}

		switch e.target {
		case "sortByTotal":
			sort.Sort(ByTotal(arg))
		case "sortByName":
			sort.Sort(ByName(arg))
		case "sortByMaxima":
			sort.Sort(ByMaxi(arg))
		case "sortByMin":
			sort.Sort(ByMini(arg))
		}

		return arg
	}

	log.Printf("unknown function in evalExpr:  %q\n", e.target)

	return nil
}

type ByTotal []*metricData

func (s ByTotal) Len() int      { return len(s) }
func (s ByTotal) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByTotal) Less(i, j int) bool {
	return summarizeValues("sum", s[i].GetValues()) > summarizeValues("sum", s[j].GetValues())
}

type ByName []*metricData

func (s ByName) Len() int           { return len(s) }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByName) Less(i, j int) bool { return s[i].GetName() < s[j].GetName() }

type ByMaxi []*metricData

func (s ByMaxi) Len() int      { return len(s) }
func (s ByMaxi) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByMaxi) Less(i, j int) bool {
	return summarizeValues("max", s[i].GetValues()) > summarizeValues("max", s[j].GetValues())
}

type ByMini []*metricData

func (s ByMini) Len() int      { return len(s) }
func (s ByMini) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByMini) Less(i, j int) bool {
	return summarizeValues("min", s[i].GetValues()) < summarizeValues("min", s[j].GetValues())
}

type seriesFunc func(*metricData, *metricData) *metricData

func forEachSeriesDo(e *expr, from, until int32, values map[metricRequest][]*metricData, function seriesFunc) []*metricData {
	arg, err := getSeriesArg(e.args[0], from, until, values)
	if err != nil {
		return nil
	}
	var results []*metricData

	for _, a := range arg {
		r := *a
		r.Name = proto.String(fmt.Sprintf("%s(%s)", e.target, a.GetName()))
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		results = append(results, function(a, &r))
	}
	return results
}

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
