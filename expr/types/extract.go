package types

import (
	"unicode"
	"unicode/utf8"

	"github.com/go-graphite/carbonapi/pkg/parser"
)

var allowedCharactersInMetricName = map[byte]struct{}{
	'=': struct{}{},
	'@': struct{}{},
}

func byteAllowedInName(b byte) bool {
	_, ok := allowedCharactersInMetricName[b]
	return ok
}

// ExtractNameLoc extracts name start:end location out of function list with . Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractNameLoc(s string) (int, int) {
	// search for a metric name in 's'
	// metric name is defined to be a Series of name characters terminated by a ',' or ')'

	var (
		start, braces, i, w int
		r                   rune
	)

FOR:
	for braces, i, w = 0, 0, 0; i < len(s); i += w {

		w = 1
		if parser.IsNameChar(s[i]) || byteAllowedInName(s[i]) {
			continue
		}

		switch s[i] {
		// If metric name have tags, we want to skip them
		case ';':
			break FOR
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
		case ')':
			break FOR
		default:
			r, w = utf8.DecodeRuneInString(s[i:])
			if unicode.In(r, parser.RangeTables...) {
				continue
			}
			start = i + 1
		}
	}

	return start, i
}

// ExtractName extracts name out of function list. Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractName(s string) string {
	start, end := ExtractNameLoc(s)
	return s[start:end]
}
