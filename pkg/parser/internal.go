package parser

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
	if e.etype != EtName {
		return false, ErrBadType
	}

	// names go into 'target'
	switch e.target {
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

func sliceExpr(args... interface{}) []Expr {
	var res []Expr
	for _, a := range args {
		switch v := a.(type) {
		case NameArg:
			res = append(res, NewNameExpr(string(v)))
		case ValueArg:
			res = append(res, NewValueExpr(string(v)))
		case float64:
			res = append(res, NewConstExpr(v))
		case int:
			res = append(res, NewConstExpr(float64(v)))
		case string:
			res = append(res, NewTargetExpr(v))
		case Expr:
			res = append(res, v)
		default:
			return nil
		}
	}

	return res
}

func mapExpr(m map[string]interface{}) map[string]Expr {
	if m == nil || len(m) == 0 {
		return nil
	}
	res := make(map[string]Expr)
	for k, a := range m {
		switch v := a.(type) {
		case NameArg:
			res[k] = NewNameExpr(string(v))
		case ValueArg:
			res[k] = NewValueExpr(string(v))
		case float64:
			res[k] = NewConstExpr(v)
		case int:
			res[k] = NewConstExpr(float64(v))
		case string:
			res[k] = NewTargetExpr(v)
		case Expr:
			res[k] = v
		default:
			return nil
		}
	}

	return res
}