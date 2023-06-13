package parser

import (
	"bytes"
	"fmt"
	"github.com/go-graphite/carbonapi/expr/holtwinters"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ansel1/merry"
)

// expression parser

type expr struct {
	target    string
	etype     ExprType
	val       float64
	valStr    string
	args      []*expr // positional
	namedArgs map[string]*expr
	argString string
}

func (e *expr) IsName() bool {
	return e.etype == EtName
}

func (e *expr) IsFunc() bool {
	return e.etype == EtFunc
}

func (e *expr) IsConst() bool {
	return e.etype == EtConst
}

func (e *expr) IsString() bool {
	return e.etype == EtString
}

func (e *expr) IsBool() bool {
	return e.etype == EtBool
}

func (e *expr) Type() ExprType {
	return e.etype
}

func (e *expr) ToString() string {
	switch e.etype {
	case EtFunc:
		return e.target + "(" + e.argString + ")"
	case EtConst:
		return e.valStr
	case EtString:
		s := e.valStr
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `'`, `\'`)
		return "'" + s + "'"
	case EtBool:
		return fmt.Sprint(e.val)
	}

	return e.target
}

func (e *expr) SetTarget(target string) {
	e.target = target
}

func (e *expr) MutateTarget(target string) Expr {
	e.SetTarget(target)
	return e
}

func (e *expr) Target() string {
	return e.target
}

func (e *expr) FloatValue() float64 {
	return e.val
}

func (e *expr) StringValue() string {
	return e.valStr
}

func (e *expr) SetValString(value string) {
	e.valStr = value
}

func (e *expr) MutateValString(value string) Expr {
	e.SetValString(value)
	return e
}

func (e *expr) RawArgs() string {
	return e.argString
}

func (e *expr) SetRawArgs(args string) {
	e.argString = args
}

func (e *expr) MutateRawArgs(args string) Expr {
	e.SetRawArgs(args)
	return e
}

func (e *expr) Args() []Expr {
	ret := make([]Expr, len(e.args))
	for i := 0; i < len(e.args); i++ {
		ret[i] = e.args[i]
	}
	return ret
}

func (e *expr) Arg(i int) Expr {
	return e.args[i]
}

func (e *expr) ArgsLen() int {
	return len(e.args)
}

func (e *expr) NamedArgs() map[string]Expr {
	ret := make(map[string]Expr)
	for k, v := range e.namedArgs {
		ret[k] = v
	}
	return ret
}

func (e *expr) NamedArg(name string) (Expr, bool) {
	expr, exist := e.namedArgs[name]
	return expr, exist
}

func (e *expr) Metrics(from, until int64) []MetricRequest {
	switch e.etype {
	case EtName:
		return []MetricRequest{{Metric: e.target, From: from, Until: until}}
	case EtConst, EtString:
		return nil
	case EtFunc:
		var r []MetricRequest
		for _, a := range e.args {
			r = append(r, a.Metrics(from, until)...)
		}

		switch e.target {
		case "transformNull":
			referenceSeriesExpr := e.GetNamedArg("referenceSeries")
			if !referenceSeriesExpr.IsInterfaceNil() {
				r = append(r, referenceSeriesExpr.Metrics(from, until)...)
			}
		case "timeShift":
			offs, err := e.GetIntervalArg(1, -1)
			if err != nil {
				return nil
			}
			for i := range r {
				r[i].From += int64(offs)
				r[i].Until += int64(offs)
			}
		case "timeStack":
			offs, err := e.GetIntervalArg(1, -1)
			if err != nil {
				return nil
			}

			start, err := e.GetIntArg(2)
			if err != nil {
				return nil
			}

			end, err := e.GetIntArg(3)
			if err != nil {
				return nil
			}

			var r2 []MetricRequest
			for _, v := range r {
				for i := int64(start); i < int64(end); i++ {
					fromNew := v.From + i*int64(offs)
					untilNew := v.Until + i*int64(offs)
					r2 = append(r2, MetricRequest{
						Metric: v.Metric,
						From:   fromNew,
						Until:  untilNew,
					})
				}
			}

			return r2
		case "holtWintersForecast", "holtWintersConfidenceBands", "holtWintersConfidenceArea":
			bootstrapInterval, err := e.GetIntervalNamedOrPosArgDefault("bootstrapInterval", 2, 1, holtwinters.DefaultBootstrapInterval)
			if err != nil {
				return nil
			}

			for i := range r {
				r[i].From -= bootstrapInterval
			}
		case "holtWintersAberration":
			bootstrapInterval, err := e.GetIntervalNamedOrPosArgDefault("bootstrapInterval", 2, 1, holtwinters.DefaultBootstrapInterval)
			if err != nil {
				return nil
			}

			// For this function, we also need to pull data with an adjusted From time,
			// so additional requests are added with the adjusted start time based on the
			// bootstrapInterval
			for i := range r {
				adjustedReq := MetricRequest{
					Metric: r[i].Metric,
					From:   r[i].From - bootstrapInterval,
					Until:  r[i].Until,
				}
				r = append(r, adjustedReq)
			}
		case "movingAverage", "movingMedian", "movingMin", "movingMax", "movingSum", "exponentialMovingAverage":
			if len(e.args) < 2 {
				return nil
			}
			if e.args[1].etype == EtString {
				offs, err := e.GetIntervalArg(1, 1)
				if err != nil {
					return nil
				}
				for i := range r {
					fromNew := r[i].From - int64(offs)
					r[i].From = fromNew
				}
			}
		case "hitcount":
			if len(e.args) < 2 {
				return nil
			}

			alignToInterval, err := e.GetBoolNamedOrPosArgDefault("alignToInterval", 2, false)
			if err != nil {
				return nil
			}
			if alignToInterval {
				bucketSizeInt32, err := e.GetIntervalArg(1, 1)
				if err != nil {
					return nil
				}

				interval := int64(bucketSizeInt32)
				// This is done in order to replicate the behavior in Graphite web when alignToInterval is set,
				// in which new data is fetched with the adjusted start time.
				for i, _ := range r {
					start := r[i].From
					for _, v := range []int64{86400, 3600, 60} {
						if interval >= v {
							start -= start % v
							break
						}
					}

					r[i].From = start
				}
			}
		case "smartSummarize":
			if len(e.args) < 2 {
				return nil
			}

			alignToInterval, err := e.GetStringNamedOrPosArgDefault("alignTo", 3, "")
			if err != nil {
				return nil
			}

			if alignToInterval != "" {
				for i, _ := range r {
					newStart, err := StartAlignTo(r[i].From, alignToInterval)
					if err != nil {
						return nil
					}
					r[i].From = newStart
				}
			}
		}
		return r
	}

	return nil
}

func (e *expr) GetIntervalArg(n, defaultSign int) (int32, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	if e.args[n].etype != EtString {
		return 0, ErrBadType
	}

	seconds, err := IntervalString(e.args[n].valStr, defaultSign)
	if err != nil {
		return 0, ErrBadType
	}

	return seconds, nil
}

func (e *expr) GetIntervalNamedOrPosArgDefault(k string, n, defaultSign int, v int64) (int64, error) {
	var val string
	var err error
	if a := e.getNamedArg(k); a != nil {
		val, err = a.doGetStringArg()
		if err != nil {
			return 0, ErrBadType
		}
	} else {
		if len(e.args) <= n {
			return v, nil
		}

		if e.args[n].etype != EtString {
			return 0, ErrBadType
		}
		val = e.args[n].valStr
	}

	seconds, err := IntervalString(val, defaultSign)
	if err != nil {
		return 0, ErrBadType
	}

	return int64(seconds), nil
}

func (e *expr) GetStringArg(n int) (string, error) {
	if len(e.args) <= n {
		return "", ErrMissingArgument
	}

	return e.args[n].doGetStringArg()
}

func (e *expr) GetStringArgs(n int) ([]string, error) {
	if len(e.args) <= n {
		return nil, ErrMissingArgument
	}

	strs := make([]string, 0, len(e.args)-n)

	for i := n; i < len(e.args); i++ {
		a, err := e.GetStringArg(i)
		if err != nil {
			return nil, err
		}
		strs = append(strs, a)
	}

	return strs, nil
}

func (e *expr) GetStringArgDefault(n int, s string) (string, error) {
	if len(e.args) <= n {
		return s, nil
	}

	return e.args[n].doGetStringArg()
}

func (e *expr) GetStringNamedOrPosArgDefault(k string, n int, s string) (string, error) {
	if a := e.getNamedArg(k); a != nil {
		return a.doGetStringArg()
	}

	return e.GetStringArgDefault(n, s)
}

func (e *expr) GetFloatArg(n int) (float64, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	return e.args[n].doGetFloatArg()
}

func (e *expr) GetFloatArgDefault(n int, v float64) (float64, error) {
	if len(e.args) <= n {
		return v, nil
	}

	return e.args[n].doGetFloatArg()
}

func (e *expr) GetFloatNamedOrPosArgDefault(k string, n int, v float64) (float64, error) {
	if a := e.getNamedArg(k); a != nil {
		return a.doGetFloatArg()
	}

	return e.GetFloatArgDefault(n, v)
}

func (e *expr) GetIntArg(n int) (int, error) {
	if len(e.args) <= n {
		return 0, ErrMissingArgument
	}

	return e.args[n].doGetIntArg()
}

func (e *expr) GetIntArgs(n int) ([]int, error) {
	if len(e.args) < n {
		return nil, ErrMissingArgument
	}

	ints := make([]int, 0, len(e.args)-n)

	for i := n; i < len(e.args); i++ {
		a, err := e.GetIntArg(i)
		if err != nil {
			return nil, err
		}
		ints = append(ints, a)
	}

	return ints, nil
}

func (e *expr) GetIntArgDefault(n, d int) (int, error) {
	if len(e.args) <= n {
		return d, nil
	}

	return e.args[n].doGetIntArg()
}

func (e *expr) GetIntArgWithIndication(n int) (int, bool, error) {
	if len(e.args) <= n {
		return 0, false, nil
	}

	v, err := e.args[n].doGetIntArg()
	return v, true, err
}

func (e *expr) GetIntNamedOrPosArgWithIndication(k string, n int) (int, bool, error) {
	if a := e.getNamedArg(k); a != nil {
		v, err := a.doGetIntArg()
		return v, true, err
	}

	return e.GetIntArgWithIndication(n)
}

func (e *expr) GetIntNamedOrPosArgDefault(k string, n, d int) (int, error) {
	if a := e.getNamedArg(k); a != nil {
		return a.doGetIntArg()
	}

	return e.GetIntArgDefault(n, d)
}

func (e *expr) GetNamedArg(name string) Expr {
	return e.getNamedArg(name)
}

func (e *expr) GetBoolNamedOrPosArgDefault(k string, n int, b bool) (bool, error) {
	if a := e.getNamedArg(k); a != nil {
		return a.doGetBoolArg()
	}

	return e.GetBoolArgDefault(n, b)
}

func (e *expr) GetBoolArgDefault(n int, b bool) (bool, error) {
	if len(e.args) <= n {
		return b, nil
	}

	return e.args[n].doGetBoolArg()
}

func (e *expr) GetNodeOrTagArgs(n int, single bool) ([]NodeOrTag, error) {
	// if single==false, zero nodes is OK
	if single && len(e.args) <= n || len(e.args) < n {
		return nil, ErrMissingArgument
	}

	nodeTags := make([]NodeOrTag, 0, len(e.args)-n)

	var err error
	until := len(e.args)
	if single {
		until = n + 1
	}
	for i := n; i < until; i++ {
		var nodeTag NodeOrTag
		nodeTag.Value, err = e.GetIntArg(i)
		if err != nil {
			// Try to parse it as String
			nodeTag.Value, err = e.GetStringArg(i)
			if err != nil {
				return nil, err
			}
			nodeTag.IsTag = true
		}
		nodeTags = append(nodeTags, nodeTag)
	}

	return nodeTags, nil
}

func (e *expr) IsInterfaceNil() bool {
	return e == nil
}

func (e *expr) insertFirstArg(exp *expr) error {
	if e.etype != EtFunc {
		return fmt.Errorf("pipe to not a function")
	}

	newArgs := []*expr{exp}
	e.args = append(newArgs, e.args...)

	if e.argString == "" {
		e.argString = exp.ToString()
	} else {
		e.argString = exp.ToString() + "," + e.argString
	}

	return nil
}

func skipWhitespace(e string) string {
	skipTo := len(e)
	for i, r := range e {
		if !unicode.IsSpace(r) {
			skipTo = i
			break
		}
	}
	return e[skipTo:]
}

func parseExprWithoutPipe(e string) (Expr, string, error) {
	e = skipWhitespace(e)

	if e == "" {
		return nil, "", ErrMissingExpr
	}

	if '0' <= e[0] && e[0] <= '9' || e[0] == '-' || e[0] == '+' {
		val, valStr, e, err := parseConst(e)
		r, _ := utf8.DecodeRuneInString(e)
		if !unicode.IsLetter(r) {
			return &expr{val: val, etype: EtConst, valStr: valStr}, e, err
		}
	}

	if e[0] == '\'' || e[0] == '"' {
		val, e, err := parseString(e)
		return &expr{valStr: val, etype: EtString}, e, err
	}

	name, e := parseName(e)

	if name == "" {
		return nil, e, ErrMissingArgument
	}

	nameLower := strings.ToLower(name)
	if nameLower == "false" || nameLower == "true" {
		return &expr{valStr: nameLower, etype: EtBool, target: nameLower}, e, nil
	}

	if e != "" && e[0] == '(' {
		// TODO(civil): Tags: make it a proper Expression
		if name == "seriesByTag" {
			argString, _, _, e, err := parseArgList(e)
			return &expr{target: name + "(" + argString + ")", etype: EtName}, e, err
		}
		exp := &expr{target: name, etype: EtFunc}

		argString, posArgs, namedArgs, e, err := parseArgList(e)
		exp.argString = argString
		exp.args = posArgs
		exp.namedArgs = namedArgs

		return exp, e, err
	}

	return &expr{target: name}, e, nil
}

func parseExprInner(e string) (Expr, string, error) {
	exp, e, err := parseExprWithoutPipe(e)
	if err != nil {
		return exp, e, err
	}
	return pipe(exp.(*expr), e)
}

// ParseExpr actually do all the parsing. It returns expression, original string and error (if any)
func ParseExpr(e string) (Expr, string, error) {
	exp, e, err := parseExprInner(e)
	if err != nil {
		return exp, e, err
	}
	exp, err = defineMap.expandExpr(exp.(*expr))
	return exp, e, err
}

func pipe(exp *expr, e string) (*expr, string, error) {
	e = skipWhitespace(e)

	if e == "" || e[0] != '|' {
		return exp, e, nil
	}

	wr, e, err := parseExprWithoutPipe(e[1:])
	if err != nil {
		return exp, e, err
	}
	if wr == nil {
		return exp, e, nil
	}

	err = wr.(*expr).insertFirstArg(exp)
	if err != nil {
		return exp, e, err
	}
	exp = wr.(*expr)

	return pipe(exp, e)
}

// IsNameChar checks if specified char is actually a valid (from graphite's protocol point of view)
func IsNameChar(r byte) bool {
	return false ||
		'a' <= r && r <= 'z' ||
		'A' <= r && r <= 'Z' ||
		'0' <= r && r <= '9' ||
		r == '.' || r == '_' ||
		r == '-' || r == '*' ||
		r == '?' || r == ':' ||
		r == '[' || r == ']' ||
		r == '^' || r == '$' ||
		r == '<' || r == '>' ||
		r == '&' || r == '#' ||
		r == '/' || r == '%' ||
		r == '@'
}

func IsDigit(r byte) bool {
	return '0' <= r && r <= '9'
}

func parseArgList(e string) (string, []*expr, map[string]*expr, string, error) {
	var (
		posArgs   []*expr
		namedArgs map[string]*expr
	)
	eOrig := e

	if e[0] != '(' {
		panic("arg list should start with paren")
	}

	var argStringBuffer bytes.Buffer

	e = e[1:]

	// check for empty args
	t := skipWhitespace(e)
	if t != "" && t[0] == ')' {
		return "", posArgs, namedArgs, t[1:], nil
	}

	charNum := 1
	for {
		var arg Expr
		var err error
		charNum++

		argString := e
		arg, e, err = parseExprInner(e)
		if err != nil {
			return "", nil, nil, e, err
		}

		if e == "" {
			return "", nil, nil, "", ErrMissingComma
		}

		// we now know we're parsing a key-value pair
		if arg.IsName() && e[0] == '=' {
			e = e[1:]
			argCont, eCont, errCont := parseExprInner(e)
			if errCont != nil {
				return "", nil, nil, eCont, errCont
			}

			if eCont == "" {
				return "", nil, nil, "", ErrMissingComma
			}

			if !argCont.IsConst() && !argCont.IsName() && !argCont.IsString() && !argCont.IsBool() {
				return "", nil, nil, eCont, ErrBadType
			}

			if namedArgs == nil {
				namedArgs = make(map[string]*expr)
			}

			exp := &expr{
				etype:  argCont.Type(),
				val:    argCont.FloatValue(),
				valStr: argCont.StringValue(),
				target: argCont.Target(),
			}
			namedArgs[arg.Target()] = exp

			e = eCont
			if argStringBuffer.Len() > 0 {
				argStringBuffer.WriteByte(',')
			}
			argStringBuffer.WriteString(argString[:len(argString)-len(e)])
			charNum += len(argString) - len(e)
		} else {
			exp := arg.toExpr().(*expr)
			posArgs = append(posArgs, exp)

			if argStringBuffer.Len() > 0 {
				argStringBuffer.WriteByte(',')
			}
			if exp.IsFunc() {
				expString := exp.ToString()
				argStringBuffer.WriteString(expString)
				charNum += len(expString)
			} else {
				argStringBuffer.WriteString(argString[:len(argString)-len(e)])
				charNum += len(argString) - len(e)
			}
		}

		// after the argument, trim any trailing spaces
		e = skipWhitespace(e)

		if e[0] == ')' {
			return argStringBuffer.String(), posArgs, namedArgs, e[1:], nil
		}

		if e[0] != ',' && e[0] != ' ' {
			return "", nil, nil, "", merry.Wrap(ErrUnexpectedCharacter).WithUserMessagef("string_to_parse=`%v`, character_number=%v, character=`%v`", eOrig, charNum, string(e[0]))
		}

		e = e[1:]
	}
}

func parseConst(s string) (float64, string, string, error) {
	var i int
	// All valid characters for a floating-point constant
	// Just slurp them all in and let ParseFloat sort 'em out
	for i < len(s) && (IsDigit(s[i]) || s[i] == '.' || s[i] == '+' || s[i] == '-' || s[i] == 'e' || s[i] == 'E') {
		i++
	}

	v, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, "", "", err
	}

	return v, s[:i], s[i:], err
}

// RangeTables is an array of *unicode.RangeTable
var RangeTables []*unicode.RangeTable

var disallowedCharactersInMetricName = map[rune]struct{}{
	'(':  struct{}{},
	')':  struct{}{},
	'"':  struct{}{},
	'\'': struct{}{},
	' ':  struct{}{},
	'/':  struct{}{},
}

func unicodeRuneAllowedInName(r rune) bool {
	if _, ok := disallowedCharactersInMetricName[r]; ok {
		return false
	}

	return true
}

func parseName(s string) (string, string) {
	var (
		braces, i, w int
		r            rune
		isEscape     bool
		isDefault    bool
	)

	buf := bytes.NewBuffer(make([]byte, 0, len(s)))

FOR:
	for braces, i, w = 0, 0, 0; i < len(s); i += w {
		if s[i] != '\\' {
			err := buf.WriteByte(s[i])
			if err != nil {
				break FOR
			}
		}
		isDefault = false
		w = 1
		if IsNameChar(s[i]) {
			continue
		}

		switch s[i] {
		case '\\':
			if isEscape {
				err := buf.WriteByte(s[i])
				if err != nil {
					break FOR
				}
				isEscape = false
				continue
			}
			isEscape = true
		case '{':
			if isEscape {
				isDefault = true
			} else {
				braces++
			}
		case '}':
			if isEscape {
				isDefault = true
			} else {
				if braces == 0 {
					break FOR
				}
				braces--
			}
		case ',':
			if isEscape {
				isDefault = true
			} else if braces == 0 {
				break FOR
			}
		/* */
		case '=':
			// allow metric name to end with any amount of `=` without treating it as a named arg or tag
			if !isEscape {
				if len(s) < i+2 || s[i+1] == '=' || s[i+1] == ',' || s[i+1] == ')' {
					continue
				}
			}
			fallthrough
		/* */
		default:
			isDefault = true
		}
		if isDefault {
			r, w = utf8.DecodeRuneInString(s[i:])
			if unicodeRuneAllowedInName(r) && unicode.In(r, RangeTables...) {
				continue
			}
			if !isEscape {
				break FOR
			}
			isEscape = false
			continue
		}
	}

	if i == len(s) {
		return buf.String(), ""
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

func StartAlignTo(start int64, alignTo string) (int64, error) {
	var newDate time.Time
	re := regexp.MustCompile(`^[0-9]+`)
	alignTo = re.ReplaceAllString(alignTo, "")

	startDate := time.Unix(start, 0).UTC()
	switch {
	case strings.HasPrefix(alignTo, "y"):
		newDate = time.Date(startDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	case strings.HasPrefix(alignTo, "mon"):
		newDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	case strings.HasPrefix(alignTo, "w"):
		if !IsDigit(alignTo[len(alignTo)-1]) {
			return start, ErrInvalidInterval
		}
		newDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		dayOfWeek, err := strconv.Atoi(alignTo[len(alignTo)-1:])
		if err != nil {
			return start, ErrInvalidInterval
		}

		startDayOfWeek := int(startDate.Weekday())
		daysToSubtract := startDayOfWeek - dayOfWeek
		if daysToSubtract < 0 {
			daysToSubtract += 7
		}
		newDate = newDate.AddDate(0, 0, -daysToSubtract)
	case strings.HasPrefix(alignTo, "d"):
		newDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	case strings.HasPrefix(alignTo, "h"):
		newDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), 0, 0, 0, time.UTC)
	case strings.HasPrefix(alignTo, "min"):
		newDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(), 0, 0, time.UTC)
	case strings.HasPrefix(alignTo, "s"):
		newDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(), startDate.Second(), 0, time.UTC)
	default:
		return start, ErrInvalidInterval
	}
	return newDate.Unix(), nil
}
