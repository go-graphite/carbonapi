package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"code.google.com/p/goprotobuf/proto"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
)

// expression parser

type exprType int

const (
	etMetric exprType = iota
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

func (e *expr) metrics() []string {

	switch e.etype {
	case etMetric:
		return []string{e.target}
	case etConst, etString:
		return nil
	case etFunc:
		var r []string
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

	if e[0] == '\'' {
		val, e, err := parseString(e)
		return &expr{valStr: val, etype: etString}, e, err
	}

	name, e := parseName(e)

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
		r == '.' || r == '_' || r == '-' || r == '*'
}

func isDigit(r byte) bool {
	return '0' <= r && r <= '9'
}

func parseConst(s string) (float64, string, error) {

	var i int
	for i < len(s) && isDigit(s[i]) {
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

	if s[0] != '\'' {
		panic("string should start with open quote")
	}

	s = s[1:]

	var i int
	for i < len(s) && s[i] != '\'' {
		i++
	}

	if i == len(s) {
		return "", "", ErrMissingQuote

	}

	return s[:i], s[i+1:], nil
}

func evalExpr(e *expr, values map[string][]*pb.FetchResponse) []*pb.FetchResponse {

	// TODO(dgryski): this should reuse the FetchResponse structs instead of allocating new ones

	switch e.etype {
	case etMetric:
		return values[e.target]
	case etConst:
		p := pb.FetchResponse{Name: proto.String(e.target), Values: []float64{e.val}}
		return []*pb.FetchResponse{&p}
	}

	// evaluate the function
	switch e.target {
	case "alias":
		arg := evalExpr(e.args[0], values)

		if len(arg) > 1 {
			return nil
		}

		if e.args[1].etype != etString {
			return nil
		}

		r := pb.FetchResponse{
			Name:      proto.String(e.args[1].valStr),
			Values:    arg[0].Values,
			IsAbsent:  arg[0].IsAbsent,
			StepTime:  arg[0].StepTime,
			StartTime: arg[0].StartTime,
			StopTime:  arg[0].StopTime,
		}

		return []*pb.FetchResponse{&r}

	case "keepLastValue":

		arg := evalExpr(e.args[0], values)

		keep := -1

		if len(e.args) > 1 {

			n := evalExpr(e.args[1], values)
			if len(n) != 1 || len(n[0].Values) != 1 {
				// fail
				return nil
			}

			keep = int(n[0].Values[0])

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

	case "movingAverage":
		arg := evalExpr(e.args[0], values)
		n := evalExpr(e.args[1], values)
		if len(n) != 1 || len(n[0].Values) != 1 {
			// fail
			return nil
		}

		windowSize := int(n[0].Values[0])

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
				w.Push(v)
				r.Values[i] = w.Mean()
			}
			result = append(result, &r)
		}
		return result

	case "nonNegativeDerivative":
		arg := evalExpr(e.args[0], values)
		var result []*pb.FetchResponse
		for _, a := range arg {
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
		arg := evalExpr(e.args[0], values)
		n := evalExpr(e.args[1], values)
		if len(n) != 1 || len(n[0].Values) != 1 {
			// fail
			return nil
		}

		scale := n[0].Values[0]

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
		arg := evalExpr(e.args[0], values)
		n := evalExpr(e.args[1], values)
		if len(n) != 1 || len(n[0].Values) != 1 {
			// fail
			return nil
		}

		seconds := n[0].Values[0]

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
		var args []*pb.FetchResponse
		for _, arg := range e.args {
			a := evalExpr(arg, values)
			args = append(args, a...)
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
	}

	return nil
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
