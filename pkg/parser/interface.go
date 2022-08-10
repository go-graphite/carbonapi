package parser

import (
	"errors"
	"fmt"
	"strings"
)

// MetricRequest contains all necessary data to request a metric.
type MetricRequest struct {
	Metric string
	From   int64
	Until  int64
}

// ExprType defines a type for expression types constants (e.x. functions, values, constants, parameters, strings)
type ExprType int

const (
	// EtName is a const for 'Series Name' type expression
	EtName ExprType = iota
	// EtFunc is a const for 'Function' type expression
	EtFunc
	// EtConst is a const for 'Constant' type expression
	EtConst
	// EtString is a const for 'String' type expression
	EtString
	// EtBool is a constant for 'Bool' type expression
	EtBool
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
	// ErrUnknownTimeUnits is an eval error returned when a time unit is unknown to system
	ErrUnknownTimeUnits = errors.New("unknown time units")
)

// NodeOrTag structure contains either Node (=integer) or Tag (=string)
// They are distinguished by "IsTag" = true in case it's tag.
type NodeOrTag struct {
	IsTag bool
	Value interface{}
}

// Expr defines an interface to talk with expressions
type Expr interface {
	// IsName checks if Expression is 'Series Name' expression
	IsName() bool
	// IsFunc checks if Expression is 'Function' expression
	IsFunc() bool
	// IsConst checks if Expression is 'Constant' expression
	IsConst() bool
	// IsString checks if Expression is 'String' expression
	IsString() bool
	// IsBool checks if Expression is 'Bool' expression
	IsBool() bool
	// Type returns type of the expression
	Type() ExprType
	// Target returns target value for expression
	Target() string
	// SetTarget changes target for the expression
	SetTarget(string)
	// MutateTarget changes target for the expression and returns new interface. Please note that it doesn't copy object yet
	MutateTarget(string) Expr
	// ToString returns string representation of expression
	ToString() string

	// FloatValue returns float value for expression.
	FloatValue() float64

	// StringValue returns value of String-typed expression (will return empty string for ConstExpr for example).
	StringValue() string
	// SetValString changes value of String-typed expression
	SetValString(string)
	// MutateValString changes ValString for the expression and returns new interface. Please note that it doesn't copy object yet
	MutateValString(string) Expr

	// Arg returns argument with index (parsed, as Expr interface as well)
	Arg(int) Expr
	// Args returns slice of arguments (parsed, as Expr interface as well)
	Args() []Expr
	// ArgsLen return arguments count
	ArgsLen() int
	// NamedArgs returns map of named arguments. E.x. for nonNegativeDerivative(metric1,maxValue=32) it will return map{"maxValue": constExpr(32)}
	NamedArgs() map[string]Expr
	// NamedArg returns named argument and boolean flag for check arg exist.
	NamedArg(string) (Expr, bool)
	// RawArgs returns string that contains all arguments of expression exactly the same order they appear
	RawArgs() string
	// SetRawArgs changes raw argument list for current expression.
	SetRawArgs(args string)
	// MutateRawArgs changes raw argument list for the expression and returns new interface. Please note that it doesn't copy object yet
	MutateRawArgs(args string) Expr

	// Metrics returns list of metric requests
	Metrics() []MetricRequest

	// GetIntervalArg returns interval typed argument.
	GetIntervalArg(n int, defaultSign int) (int32, error)

	// GetIntervalNamedOrPosArgDefault returns interval typed argument that can be passed as a named argument or as position or replace it with default if none found.
	GetIntervalNamedOrPosArgDefault(k string, n, defaultSign int, v int64) (int64, error)

	// GetStringArg returns n-th argument as string.
	GetStringArg(n int) (string, error)
	// GetStringArgs returns n-th argument as slice of strings.
	GetStringArgs(n int) ([]string, error)
	// GetStringArgDefault returns n-th argument as string. It will replace it with Default value if none present.
	GetStringArgDefault(n int, s string) (string, error)
	// GetStringNamedOrPosArgDefault returns specific positioned string-typed argument or replace it with default if none found.
	GetStringNamedOrPosArgDefault(k string, n int, s string) (string, error)

	// GetFloatArg returns n-th argument as float-typed (if it's convertible to float)
	GetFloatArg(n int) (float64, error)
	// GetFloatArgDefault returns n-th argument as float. It will replace it with Default value if none present.
	GetFloatArgDefault(n int, v float64) (float64, error)
	// GetFloatNamedOrPosArgDefault returns specific positioned float64-typed argument or replace it with default if none found.
	GetFloatNamedOrPosArgDefault(k string, n int, v float64) (float64, error)

	// GetIntArg returns n-th argument as int-typed
	GetIntArg(n int) (int, error)
	// GetIntArgs returns n-th argument as slice of ints
	GetIntArgs(n int) ([]int, error)
	// GetIntArgDefault returns n-th argument as int. It will replace it with Default value if none present.
	GetIntArgDefault(n int, d int) (int, error)
	// GetIntArgWithIndication returns n-th argument as int. If argument wasn't present, second return value will be `false`. Even if there was error in parsing data, but it was there, second value will be `true`
	GetIntArgWithIndication(n int) (int, bool, error)
	// GetIntNamedOrPosArgWithIndication returns specific positioned int-typed argument. If argument wasn't present, second return value will be `false`. Even if there was error in parsing data, but it was there, second value will be `true`
	GetIntNamedOrPosArgWithIndication(k string, n int) (int, bool, error)
	// GetIntNamedOrPosArgDefault returns specific positioned int-typed argument or replace it with default if none found.
	GetIntNamedOrPosArgDefault(k string, n int, d int) (int, error)

	GetNamedArg(name string) Expr

	// GetBoolArgDefault returns n-th argument as bool. It will replace it with Default value if none present.
	GetBoolArgDefault(n int, b bool) (bool, error)
	// GetBoolNamedOrPosArgDefault returns specific positioned bool-typed argument or replace it with default if none found.
	GetBoolNamedOrPosArgDefault(k string, n int, b bool) (bool, error)

	// GetNodeOrTagArgs returns the last arguments starting from the n-th as a slice of NodeOrTag structures. If `single` is `true`, only the n-th argument is taken.
	GetNodeOrTagArgs(n int, single bool) ([]NodeOrTag, error)

	IsInterfaceNil() bool

	toExpr() interface{}
}

var _ Expr = &expr{}

// Parse parses string as an expression.
func Parse(e string) (Expr, string, error) {
	return ParseExpr(e)
}

// NewTargetExpr Creates new expression with specified target only.
func NewTargetExpr(target string) Expr {
	e := &expr{
		target:    target,
		argString: target,
	}
	return e
}

// NewNameExpr Creates new expression with specified name only.
func NewNameExpr(name string) Expr {
	e := &expr{
		target:    name,
		etype:     EtName,
		argString: name,
	}
	return e
}

// NewConstExpr Creates new Constant expression.
func NewConstExpr(value float64) Expr {
	e := &expr{
		val:       value,
		etype:     EtConst,
		argString: fmt.Sprintf("%v", value),
	}
	return e
}

// NewValueExpr Creates new Value expression.
func NewValueExpr(value string) Expr {
	e := &expr{
		valStr:    value,
		etype:     EtString,
		argString: value,
	}
	return e
}

// ArgName is a type for Name Argument
type ArgName string

// ArgValue is a type for Value Argument
type ArgValue string

// NamedArgs is a type for Hashmap of Named Arguments.
type NamedArgs map[string]interface{}

// NewExpr creates a new expression with specified target and arguments. It will do best it can to identify type of argument
func NewExpr(target string, vaArgs ...interface{}) Expr {
	var nArgsFinal map[string]*expr
	args, nArgs := sliceExpr(vaArgs)
	if args == nil {
		panic(fmt.Sprintf("unsupported argument list for target=%v\n", target))
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
			argStrs = append(argStrs, k+"="+v.RawArgs())
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

// NewExprTyped creates a new expression with specified target and arguments. Strictly typed one.
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
