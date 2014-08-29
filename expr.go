package main

import (
	"errors"
	"fmt"
	"strconv"
)

// expression parser

type exprType int

const (
	etMetric exprType = iota
	etFunc
	etConst
)

type expr struct {
	target    string
	etype     exprType
	val       float64
	args      []*expr
	argString string
}

func (e *expr) metrics() []string {

	switch e.etype {
	case etMetric:
		return []string{e.target}
	case etConst:
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

type namedExpr struct {
	name string
	data []float64
}

func evalExpr(e *expr, values map[string][]namedExpr) []namedExpr {

	switch e.etype {
	case etMetric:
		return values[e.target]
	case etConst:
		return []namedExpr{{name: e.target, data: []float64{e.val}}}
	}

	// evaluate the function
	switch e.target {
	case "sum", "sumSeries":
		// make sure the arrays are all the same 'size'
		var args []namedExpr
		for _, arg := range e.args {
			a := evalExpr(arg, values)
			args = append(args, a...)
		}
		r := namedExpr{
			name: fmt.Sprintf("sum(%s)", e.argString),
			data: make([]float64, len(args[0].data)),
		}
		for _, arg := range args {
			for i, v := range arg.data {
				r.data[i] += v
			}
		}
		return []namedExpr{r}

	case "nonNegativeDerivative":
		arg := evalExpr(e.args[0], values)
		var result []namedExpr
		for _, a := range arg {
			r := namedExpr{
				name: fmt.Sprintf("nonNegativeDerivative(%s)", a.name),
				data: make([]float64, len(a.data)),
			}
			r.data[0] = 0
			for i := 1; i < len(a.data); i++ {
				r.data[i] = a.data[i] - a.data[i-1]
				if r.data[i] < 0 {
					r.data[i] = r.data[i-1]
				}
			}
			result = append(result, r)
		}
		return result

	case "movingAverage":
		arg := evalExpr(e.args[0], values)
		n := evalExpr(e.args[1], values)
		if len(n) != 1 || len(n[0].data) != 1 {
			// fail
			return nil
		}

		windowSize := int(n[0].data[0])

		var result []namedExpr

		for _, a := range arg {
			w := &Windowed{data: make([]float64, windowSize)}
			r := namedExpr{
				name: fmt.Sprintf("movingAverage(%s, %d)", a.name, windowSize),
				data: make([]float64, len(a.data)),
			}
			for i, v := range a.data {
				w.Push(v)
				r.data[i] = w.Mean()
			}
			result = append(result, r)
		}
		return result
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
