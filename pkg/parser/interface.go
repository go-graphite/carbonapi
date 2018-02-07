package parser

import (
	"errors"
	"strings"
	"fmt"
)

type MetricRequest struct {
	Metric string
	From   int32
	Until  int32
}

type ExprType int

const (
	EtName   ExprType = iota
	EtFunc
	EtConst
	EtString
)

var (
	// ErrMissingExpr is a parse error returned when an expression is missing.
	ErrMissingExpr = errors.New("missing expression")
	// ErrMissingComma is a parse error returned when an expression is missing a comma.
	ErrMissingComma = errors.New("missing comma")
	// ErrMissingQuote is a parse error returned when an expression is missing a quote.
	ErrMissingQuote = errors.New("missing quote")
	// ErrUnexpectedCharacter is a parse error returned when an expression contains an unexpected character.
	ErrUnexpectedCharacter = errors.New("unexpected character")
	// ErrBadType is an eval error returned when a argument has wrong type.
	ErrBadType = errors.New("bad type")
	// ErrMissingArgument is an eval error returned when a argument is missing.
	ErrMissingArgument = errors.New("missing argument")
	// ErrMissingTimeseries is an eval error returned when a time series argument is missing.
	ErrMissingTimeseries = errors.New("missing time series argument")
	// ErrSeriesDoesNotExist is an eval error returned when a requested time series argument does not exist.
	ErrSeriesDoesNotExist = errors.New("no timeseries with that name")
)

type Expr interface {
	IsName() bool
	IsFunc() bool
	IsConst() bool
	IsString() bool
	Type() ExprType
	Target() string
	SetTarget(string)
	MutateTarget(string) Expr
	FloatValue() float64
	StringValue() string
	SetValString(string)
	MutateValString(string) Expr
	Args() []Expr
	NamedArgs() map[string]Expr
	RawArgs() string
	SetRawArgs(args string)
	MutateRawArgs(args string) Expr
	Metrics() []MetricRequest

	GetIntervalArg(n int, defaultSign int) (int32, error)

	GetStringArg(n int) (string, error)
	GetStringArgDefault(n int, s string) (string, error)
	GetStringNamedOrPosArgDefault(k string, n int, s string) (string, error)

	GetFloatArg(n int) (float64, error)
	GetFloatArgDefault(n int, v float64) (float64, error)
	GetFloatNamedOrPosArgDefault(k string, n int, v float64) (float64, error)

	GetIntArg(n int) (int, error)
	GetIntArgs(n int) ([]int, error)
	GetIntArgDefault(n int, d int) (int, error)
	GetIntNamedOrPosArgDefault(k string, n int, d int) (int, error)

	GetNamedArg(name string) Expr

	GetBoolNamedOrPosArgDefault(k string, n int, b bool) (bool, error)
	GetBoolArgDefault(n int, b bool) (bool, error)

	toExpr() interface{}
}

var _ Expr = &expr{}

func Parse(e string) (Expr, string, error) {
	return ParseExpr(e)
}

func NewTargetExpr(target string) Expr {
	e := &expr{
		target: target,
		argString: target,
	}
	return e
}

func NewNameExpr(name string) Expr {
	e := &expr{
		target: name,
		etype: EtName,
		argString: name,
	}
	return e
}

func NewConstExpr(value float64) Expr {
	e := &expr{
		val: value,
		etype: EtConst,
		argString: fmt.Sprintf("%v", value),
	}
	return e
}

func NewValueExpr(value string) Expr {
	e := &expr{
		valStr: value,
		etype: EtString,
		argString: value,
	}
	return e
}

type ArgName string
type ArgValue string
type NamedArgs map[string]interface{}

func NewExpr(target string, vaArgs... interface{}) Expr {
	var nArgsFinal map[string]*expr
	args, nArgs := sliceExpr(vaArgs)
	if args == nil {
		fmt.Printf("Unsupported argument list for target=%v\n", target)
	}

	var a []*expr
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, arg.RawArgs())
		a = append(a, arg)
	}

	if nArgs != nil {
		nArgsFinal = make(map[string]*expr)
		for k, v := range nArgs {
			nArgsFinal[k] = v
			argStrs = append(argStrs, k + "=" + v.RawArgs())
		}
	}

	e := &expr{
		target:    target,
		etype:     EtFunc,
		args:      a,
		argString: strings.Join(argStrs, ","),
	}

	if nArgsFinal != nil {
		e.namedArgs = nArgsFinal
	}

	return e
}

func NewExprTyped(target string, args []Expr) Expr {
	var a []*expr
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, arg.Target())
		a = append(a, arg.toExpr().(*expr))
	}

	e := &expr{
		target:    target,
		etype:     EtFunc,
		args:      a,
		argString: strings.Join(argStrs, ","),
	}

	return e
}