package parser

import (
	"fmt"

	"runtime/debug"
)

func (e *expr) doGetIntArg() (int, error) {
	if e.etype != EtConst {
		return 0, ErrBadType
	}

	return int(e.val), nil
}

func (e *expr) getNamedArg(name string) *expr {
	if a, ok := e.namedArgs[name]; ok {
		return a
	}

	return nil
}

func (e *expr) doGetFloatArg() (float64, error) {
	if e.etype != EtConst {
		return 0, ErrBadType
	}

	return e.val, nil
}

func (e *expr) doGetStringArg() (string, error) {
	if e.etype != EtString {
		return "", ErrBadType
	}

	return e.valStr, nil
}

func (e *expr) doGetBoolArg() (bool, error) {
	if e.etype != EtString && e.etype != EtBool {
		return false, ErrBadType
	}

	switch e.valStr {
	case "False", "false":
		return false, nil
	case "True", "true":
		return true, nil
	}

	return false, ErrBadType
}

func (e *expr) toExpr() interface{} {
	return e
}

func mergeNamedArgs(arg1, arg2 map[string]*expr) map[string]*expr {
	res := make(map[string]*expr)

	for k, v := range arg1 {
		res[k] = v
	}

	for k, v := range arg2 {
		res[k] = v
	}

	return res
}

func sliceExpr(args []interface{}) ([]*expr, map[string]*expr) {
	var res []*expr
	var nArgs map[string]*expr
	for _, a := range args {
		switch v := a.(type) {
		case ArgName:
			res = append(res, NewNameExpr(string(v)).toExpr().(*expr))
		case ArgValue:
			res = append(res, NewValueExpr(string(v)).toExpr().(*expr))
		case float64:
			res = append(res, NewConstExpr(v).toExpr().(*expr))
		case int:
			res = append(res, NewConstExpr(float64(v)).toExpr().(*expr))
		case string:
			res = append(res, NewTargetExpr(v).toExpr().(*expr))
		case *expr:
			res = append(res, v)
		case Expr:
			res = append(res, v.toExpr().(*expr))
		case NamedArgs:
			nArgsNew := mapExpr(v)
			nArgs = mergeNamedArgs(nArgs, nArgsNew)
		default:
			panic(fmt.Sprintf("BUG! THIS SHOULD NEVER HAPPEN! Unknown type=%T\n%v\n", a, string(debug.Stack())))
		}
	}

	return res, nArgs
}

func mapExpr(m NamedArgs) map[string]*expr {
	if m == nil || len(m) == 0 {
		return nil
	}
	res := make(map[string]*expr)
	for k, a := range m {
		switch v := a.(type) {
		case ArgName:
			res[k] = NewNameExpr(string(v)).toExpr().(*expr)
		case ArgValue:
			res[k] = NewValueExpr(string(v)).toExpr().(*expr)
		case float64:
			res[k] = NewConstExpr(v).toExpr().(*expr)
		case int:
			res[k] = NewConstExpr(float64(v)).toExpr().(*expr)
		case string:
			res[k] = NewTargetExpr(v).toExpr().(*expr)
		case Expr:
			res[k] = v.toExpr().(*expr)
		case *expr:
			res[k] = v
		default:
			return nil
		}
	}

	return res
}
