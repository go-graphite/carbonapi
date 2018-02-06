package parser

import (
	"errors"
	"strings"
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
	FloatValue() float64
	StringValue() string
	SetString(string)
	Args() []Expr
	NamedArgs() map[string]Expr
	RawArgs() string
	SetRawArgs(args string)
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
	}
	return e
}

func NewNameExpr(name string) Expr {
	e := &expr{
		target: name,
		etype: EtName,
	}
	return e
}

func NewConstExpr(value float64) Expr {
	e := &expr{
		val: value,
		etype: EtConst,
	}
	return e
}

func NewValueExpr(value string) Expr {
	e := &expr{
		valStr: value,
		etype: EtString,
	}
	return e
}

type NameArg string
type ValueArg string

func NewExpr(target string, vaArgs... interface{}) Expr {
	args := sliceExpr(vaArgs)
	var a []*expr
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, arg.Target())
		a = append(a, &expr{target: arg.Target()})
	}

	e := &expr{
		target:    target,
		etype:     EtFunc,
		args:      a,
		argString: strings.Join(argStrs, ","),
	}

	return e
}

func NewExprNamed(target string, namedArgsIface map[string]interface{}, vaArgs... interface{}) Expr {
	args := sliceExpr(vaArgs)
	namedArgs := mapExpr(namedArgsIface)
	var a []*expr
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, arg.Target())
		a = append(a, &expr{target: arg.Target()})
	}

	nArgs := make(map[string]*expr)
	if namedArgs != nil {
		for k, v := range namedArgs {
			nArgs[k] = v.toExpr().(*expr)
			argStrs = append(argStrs, k + "=" + v.Target())
		}
	}

	e := &expr{
		target:    target,
		etype:     EtFunc,
		args:      a,
		argString: strings.Join(argStrs, ","),
		namedArgs: nArgs,
	}

	return e
}

func NewExprTyped(target string, args []Expr) Expr {
	var a []*expr
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, arg.Target())
		a = append(a, &expr{target: arg.Target()})
	}

	e := &expr{
		target:    target,
		etype:     EtFunc,
		args:      a,
		argString: strings.Join(argStrs, ","),
	}

	return e
}